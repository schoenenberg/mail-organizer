package trigger

import (
	"errors"
	"time"

	"github.com/emersion/go-imap"
)

type DateBefore string

func (s DateBefore) Evaluate(msg imap.Message) (bool, error) {
	if msg.Envelope != nil {
		dur, err := time.ParseDuration(string(s))
		if err != nil {
			return false, err
		}
		return msg.Envelope.Date.Add(dur).Before(time.Now()), nil
	} else {
		return false, errors.New("envelope missing")
	}
}
