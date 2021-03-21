package trigger

import (
	"errors"

	"github.com/emersion/go-imap"
)

type FromDomainContains string

func (s FromDomainContains) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return addressContainsDomain(msg.Envelope.From, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}

func addressContainsDomain(s []*imap.Address, e string) bool {
	for _, a := range s {
		if a != nil && a.HostName == e {
			return true
		}
	}
	return false
}