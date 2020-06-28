package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/chickenzord/mailgrep/filter"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

type Printer interface {
	Print(*imap.Message) string
}

type SubjectPrinter struct {
}

func SortByDateDesc(messages []*imap.Message) {
	sort.Slice(messages[:], func(i, j int) bool {
		return messages[i].Envelope.Date.After(messages[j].Envelope.Date)
	})
}

func (e *SubjectPrinter) Print(msg *imap.Message) string {
	return msg.Envelope.Date.String() + " " + msg.Envelope.Subject
}

func main() {
	// Parse args
	var dotenv string
	var mailbox string
	var sender string
	var within time.Duration
	flag.StringVar(&dotenv, "dotenv", ".env", "DotEnv file to load")
	flag.StringVar(&mailbox, "mailbox", "INBOX", "IMAP Mailbox Name")
	flag.StringVar(&sender, "sender", "", "Sender Email Address")
	flag.DurationVar(&within, "within", 0, "Within the last X seconds")
	flag.Parse()

	// Load connection
	if dotenv != "" {
		godotenv.Overload(dotenv)
	}
	hostname := os.Getenv("IMAP_HOSTNAME")
	port := os.Getenv("IMAP_PORT")
	username := os.Getenv("IMAP_USERNAME")
	password := os.Getenv("IMAP_PASSWORD")
	address := fmt.Sprintf("%s:%s", hostname, port)

	// Load filters
	printer := &SubjectPrinter{}
	filters := []filter.Filter{}
	if sender != "" {
		filters = append(filters, filter.SenderAddress(sender))
	}
	if within > 0 {
		filters = append(filters, filter.Within(within))
	}

	// Connect to IMAP
	c, err := client.DialTLS(address, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()
	if err := c.Login(username, password); err != nil {
		log.Fatal(err)
	}

	// Switch mailbox
	mbox, err := c.Select(mailbox, false)
	if err != nil {
		log.Fatal(err)
	}
	if mbox.Messages == 0 {
		return
	}

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(1, mbox.Messages)
	result := make(chan *imap.Message, mbox.Messages)
	err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, result)
	if err != nil {
		log.Fatal(err)
	}

	// Filter and Sort
	messages := []*imap.Message{}
	for msg := range result {
		ok := true
		for _, filter := range filters {
			if !filter.Filter(msg) {
				ok = false
				break
			}
		}
		if ok {
			messages = append(messages, msg)
		}
	}
	SortByDateDesc(messages)

	// Print
	for _, msg := range messages {
		fmt.Println(printer.Print(msg))
	}
}
