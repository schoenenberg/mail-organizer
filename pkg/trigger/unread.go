package trigger

import (
	"errors"

	"github.com/emersion/go-imap"
)

type Unread bool

func (u Unread) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		return false, nil
	} else {
		return false, errors.New("envelope missing")
	}
}



