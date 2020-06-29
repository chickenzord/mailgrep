package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chickenzord/mailgrep"
	"github.com/chickenzord/mailgrep/filter"
	"github.com/emersion/go-imap"
	"github.com/joho/godotenv"
)

type Printer interface {
	Print(*imap.Message) string
}

type SubjectPrinter struct {
}

func (e *SubjectPrinter) Print(msg *imap.Message) string {
	return msg.Envelope.Subject
}

func main() {
	Start()
}

func Start() {
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
	if hostname == "" {
		log.Fatal("IMAP_HOSTNAME is required")
	}

	// Load filters
	printer := &SubjectPrinter{}
	filters := []filter.Filter{}
	if sender != "" {
		filters = append(filters, filter.SenderAddress(sender))
	}
	if within > 0 {
		filters = append(filters, filter.Within(within))
	}

	messages, err := mailgrep.ListEmail(
		&mailgrep.ImapConfig{
			Address:  address,
			Username: username,
			Password: password,
		},
		&mailgrep.ListRequest{
			Mailbox: mailbox,
			Filters: filters,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messages {
		fmt.Println(printer.Print(&msg))
	}
}
