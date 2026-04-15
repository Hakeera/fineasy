package internal

import "time"

type Email struct {
	ID         int
	GmailID    string
	Subject    string
	From       string
	ReceivedAt time.Time
}
