package trigger

import (
	"errors"
	"strings"

	"github.com/emersion/go-imap"
)

type SubjectContains string

func (s SubjectContains) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return strings.Contains(msg.Envelope.Subject, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}
