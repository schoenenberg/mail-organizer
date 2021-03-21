package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/emersion/go-imap"
	move "github.com/emersion/go-imap-move"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"gopkg.in/yaml.v3"

	"github.com/schoenenberg/go-mail-organizer/pkg/trigger"
)

type Account struct {
	Host         string
	Port         uint16
	Username     string
	Password     string
	TLSMode      bool
	StartTLSMode bool
	Rules        []Rule
}

func (a *Account) transform() {
	for ruleIdx, _ := range a.Rules {
		for disRuleIdx, _ := range a.Rules[ruleIdx].Trigger.DisjunctiveRules {
			r := a.Rules[ruleIdx].Trigger.DisjunctiveRules[disRuleIdx].External
			rules := make([]trigger.RuleEvaluator, 0)
			if r.SubjectExact != nil {
				rules = append(rules, r.SubjectExact)
			}
			if r.SubjectStartsWith != nil {
				rules = append(rules, r.SubjectStartsWith)
			}
			if r.SubjectEndsWith != nil {
				rules = append(rules, r.SubjectEndsWith)
			}
			if r.SubjectContains != nil {
				rules = append(rules, r.SubjectContains)
			}
			if r.FromDomainContains != nil {
				rules = append(rules, r.FromDomainContains)
			}
			if r.FromPersonalNameContains != nil {
				rules = append(rules, r.FromPersonalNameContains)
			}
			if r.DateBefore != nil {
				rules = append(rules, r.DateBefore)
			}
			a.Rules[ruleIdx].Trigger.DisjunctiveRules[disRuleIdx].internal = rules
			a.Rules[ruleIdx].Trigger.DisjunctiveRules[disRuleIdx].External = nil
		}
	}
}

type Rule struct {
	Name    string
	Enabled bool
	Trigger TriggerRuleset
	Action  Action
}

type TriggerRuleset struct {
	DisjunctiveRules []DisjunctiveRules `yaml:"or"`
}

type DisjunctiveRules struct {
	internal []trigger.RuleEvaluator
	External *ConjunctiveRuleSet `yaml:"and"`
}

type ConjunctiveRuleSet struct {
	SubjectContains          *trigger.SubjectContains          `yaml:"SubjectContains,omitempty"`
	SubjectStartsWith        *trigger.SubjectStartsWith        `yaml:"SubjectStartsWith,omitempty"`
	SubjectEndsWith          *trigger.SubjectEndsWith          `yaml:"SubjectEndsWith,omitempty"`
	SubjectExact             *trigger.SubjectExact             `yaml:"SubjectExact,omitempty"`
	FromPersonalNameContains *trigger.FromPersonalNameContains `yaml:"FromPersonalNameContains,omitempty"`
	FromDomainContains       *trigger.FromDomainContains       `yaml:"FromDomainContains,omitempty"`
	DateBefore               *trigger.DateBefore               `yaml:"DateBefore,omitempty"`
}

type Action struct {
	MoveIntoMailbox *string `yaml:"MoveIntoMailbox,omitempty"`
}

type UidClient struct{ *uidplus.Client }

type IMAPClient struct {
	*client.Client
	*UidClient
	*move.MoveClient
}

func NewIMAPClient(client *client.Client, uidClient *uidplus.Client, moveClient *move.MoveClient) *IMAPClient {
	return &IMAPClient{Client: client, UidClient: &UidClient{uidClient}, MoveClient: moveClient}
}

func main() {
	// Either use provides args or list all yaml files in directory
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	} else {
		var err error
		args, err = filepath.Glob("*.yml")
		if err != nil {
			log.Fatalln(err)
		}

	}

	for _, i := range args {
		b, err := ioutil.ReadFile(i)
		if err != nil {
			log.Fatalln("File not found:", i)
		}

		var acc Account
		err = yaml.Unmarshal(b, &acc)
		if err != nil {
			log.Fatalln("Not parsable config file:", string(b))
		}

		acc.transform()
		log.Printf("%+v", acc)

		_, err = run(acc)
		if err != nil {
			log.Println("Error on running", acc.Host, "-", err.Error())
		}
	}
}

func createClient(acc Account) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", acc.Host, acc.Port)

	if acc.TLSMode {
		return client.DialTLS(addr, nil)
	} else {
		return client.Dial(addr)
	}
}

func run(acc Account) (uint, error) {
	c, err := createClient(acc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	ic := NewIMAPClient(c, uidplus.NewClient(c), move.NewClient(c))

	_, err = ic.SupportUidPlus()
	if err != nil {
		return 0, err
	}

	if acc.StartTLSMode {
		if err := c.StartTLS(nil); err != nil {
			log.Fatal(err)
		}
	}

	// Login
	if err := c.Login(acc.Username, acc.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	res, err := ic.ListMailboxes()
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range res {
		fmt.Println("-", r)
	}

	// Select INBOX
	mbox, err := ic.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid}, messages)
	}()

	actionMap := make(map[Action][]uint32, 0)
	for msg := range messages {
		act := DetermineAction(msg, acc.Rules)

		if act != nil {
			uidActList, ok := actionMap[*act]
			if !ok {
				slice := make([]uint32, 1)
				slice[0] = msg.Uid
				actionMap[*act] = slice
			} else {
				uidActList = append(uidActList, msg.Uid)
				actionMap[*act] = uidActList
			}
		}
	}

	for k, v := range actionMap {
		if targetMailbox := k.MoveIntoMailbox; targetMailbox != nil {
			var set imap.SeqSet
			for _, m := range v {
				set.AddNum(m)
			}
			log.Println("Moving UIDs", set.String())
			err = ic.UidMove(set, *targetMailbox)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	return 0, nil
}

func (c *IMAPClient) ListMailboxes() ([]string, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	resultMailboxes := make([]string, 0)
	for m := range mailboxes {
		resultMailboxes = append(resultMailboxes, m.Name)
	}

	if err := <-done; err != nil {
		return nil, err
	}

	return resultMailboxes, nil
}

func (c *IMAPClient) UidMove(set imap.SeqSet, dest string) error {
	return c.UidMoveWithFallback(&set, dest)
}

func DetermineAction(msg *imap.Message, ruleset []Rule) *Action {
	for _, rule := range ruleset {
		if !rule.Enabled {
			continue
		}
		for _, disRule := range rule.Trigger.DisjunctiveRules {
			numOfRules := len(disRule.internal)
			numOfSuccesses := 0
			for _, conRule := range disRule.internal {
				eval, err := conRule.Evaluate(*msg)

				if err != nil {
					log.Println("Error during evaluation of Rule", err)
					break
				} else if !eval {
					break
				} else {
					numOfSuccesses += 1
				}
			}

			if numOfSuccesses == numOfRules && numOfSuccesses > 0 {
				return &rule.Action
			} else {
				continue
			}
		}
	}
	return nil
}
