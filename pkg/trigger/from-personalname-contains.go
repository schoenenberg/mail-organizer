package trigger

import (
	"errors"

	"github.com/emersion/go-imap"
)

type FromPersonalNameContains string

func (s FromPersonalNameContains) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return addressContainsPersonalName(msg.Envelope.From, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}

func addressContainsPersonalName(s []*imap.Address, e string) bool {
	for _, a := range s {
		if a != nil && a.PersonalName == e {
			return true
		}
	}
	return false
}
