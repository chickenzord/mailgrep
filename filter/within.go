package filter

import (
	"time"

	"github.com/emersion/go-imap"
)

type within struct {
	within time.Duration
}

func Within(duration time.Duration) Filter {
	return &within{within: duration}
}

func (f *within) Filter(msg *imap.Message) bool {
	return msg.Envelope.Date.After(time.Now().Add(-f.within))
}
