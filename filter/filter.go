package filter

import "github.com/emersion/go-imap"

type Filter interface {
	Filter(*imap.Message) bool
}
