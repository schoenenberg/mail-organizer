package trigger

import (
	"errors"
	"strings"

	"github.com/emersion/go-imap"
)

type SubjectStartsWith string

func (s SubjectStartsWith) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return strings.HasPrefix(msg.Envelope.Subject, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}



