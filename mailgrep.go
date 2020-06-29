package mailgrep

import (
	"log"
	"sort"

	"github.com/chickenzord/mailgrep/filter"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type ImapConfig struct {
	Address  string
	Username string
	Password string
}

type ListRequest struct {
	Mailbox string
	Filters []filter.Filter
}

func sortByDateDesc(messages []imap.Message) {
	sort.Slice(messages[:], func(i, j int) bool {
		return messages[i].Envelope.Date.After(messages[j].Envelope.Date)
	})
}

func ListEmail(cfg *ImapConfig, req *ListRequest) ([]imap.Message, error) {
	// Connect to IMAP
	c, err := client.DialTLS(cfg.Address, nil)
	if err != nil {
		return nil, err
	}
	defer c.Logout()
	if err := c.Login(cfg.Username, cfg.Password); err != nil {
		return nil, err
	}

	// Switch mailbox
	mbox, err := c.Select(req.Mailbox, false)
	if err != nil {
		log.Fatal(err)
	}
	if mbox.Messages == 0 {
		return []imap.Message{}, nil
	}

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)
	result := make(chan *imap.Message, mbox.Messages)
	err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, result)
	if err != nil {
		return nil, err
	}

	// Filter and Sort
	messages := []imap.Message{}
	for msg := range result {
		ok := true
		for _, filter := range req.Filters {
			if !filter.Filter(msg) {
				ok = false
				break
			}
		}
		if ok {
			messages = append(messages, *msg)
		}
	}
	sortByDateDesc(messages)

	return messages, nil
}
