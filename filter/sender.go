package filter

import "github.com/emersion/go-imap"

type sender struct {
	address string
}

func SenderAddress(address string) Filter {
	return &sender{address: address}
}

func (f *sender) Filter(msg *imap.Message) bool {
	for _, sender := range msg.Envelope.Sender {
		if f.address == sender.Address() {
			return true
		}

	}
	return false
}
