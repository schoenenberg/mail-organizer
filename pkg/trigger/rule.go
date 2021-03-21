package trigger

import (
	"github.com/emersion/go-imap"
)

type RuleEvaluator interface {
	Evaluate(msg imap.Message) (bool, error)
}