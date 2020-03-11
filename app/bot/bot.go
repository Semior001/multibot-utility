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
	Text        string        // text of the message
	Pin         bool          // enable pin
	Preview     bool          // enable web preview
	ReplyTo     *Message      // message that we have to reply to, might be nil, if caused by other action
	BanInterval time.Duration // bot banning user set the interval
}

// Bot describes a particular bot, that reacts on messages and sends whatever
type Bot interface {
	OnMessage(msg Message) *Response // nil if nothing to send
	Help() string                    // returns help message - how to use this bot
}
