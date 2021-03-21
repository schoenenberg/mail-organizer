package trigger

import (
	"errors"
	"strings"

	"github.com/emersion/go-imap"
)

type SubjectExact string

func (s SubjectExact) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return strings.EqualFold(msg.Envelope.Subject, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}



