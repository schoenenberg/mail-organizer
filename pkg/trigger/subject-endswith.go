package trigger

import (
	"errors"
	"strings"

	"github.com/emersion/go-imap"
)

type SubjectEndsWith string

func (s SubjectEndsWith) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return strings.HasSuffix(msg.Envelope.Subject, string(s)), nil
	} else {
		return false, errors.New("envelope missing")
	}
}
