package bot

import "time"

// User defines user info of the Message
type User struct {
	ID          string
	Username    string // might be "" for users without usernames
	DisplayName string
	IsAdmin     bool
}

// Message to pass data from/to bot
type Message struct {
	ID     string
	ChatID string
	From   User
	Sent   time.Time
	Text   string `json:",omitempty"`
}

// Response describes bot's answer on particular message
type Response struct {
	Text        string
	Pin         bool          // enable pin
	Preview     bool          // enable web preview
	BanInterval time.Duration // bot banning user set the interval
}

// Bot describes a particular bot, that reacts on messages and sends whatever
type Bot interface {
	OnMessage(msg Message) *Response // nil if nothing to send
	ReactsOn() []string              // reacts on returns the list of messages that will trigger the
	Help() string                    // returns help message - how to use this bot
}
