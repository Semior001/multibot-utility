// Package bot contains all bots implementations and tests to them
// Multibot is a manager for all bots to simply distribute messages
// to all bots
package bot

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ChatType describes type of the chat, where the message
// came from
type ChatType int

// All recognizable chat types
const (
	ChatTypePrivate = 0
	ChatTypeGroup   = 1
	ChatTypeChannel = 2
)

//go:generate mockery -inpkg -name Bot -case snake

// Bot describes a particular bot, that reacts on messages and sends whatever
type Bot interface {
	OnMessage(msg Message) *Response // nil if nothing to send
	Help() string                    // returns help message - how to use this bot
}

// User defines user info of the Message
type User struct {
	ID          string
	Username    string // might be "" for users without usernames
	DisplayName string
	IsAdmin     bool
	IsBot       bool
}

// Message to pass data from/to bot
type Message struct {
	ID             string
	ChatID         string
	ChatType       ChatType
	From           *User
	Sent           time.Time
	Text           string `json:",omitempty"`
	AddedBotToChat bool
}

// Response describes bot's answer on particular message
type Response struct {
	Text        string        // text of the message
	Pin         bool          // enable pin
	Unpin       bool          // unpin current pinned message
	Preview     bool          // enable web preview
	Reply       bool          // message that we have to reply to, might be nil, if caused by other action
	BanInterval time.Duration // bot banning user set the interval
}

// IsEmpty checks that response is empty and we do not have to send it
func (r Response) IsEmpty() bool {
	return r.Text == "" && !r.Pin && !r.Unpin && !r.Preview && !r.Reply && r.BanInterval > 0
}

// MultiBot is bot that delivers messages to bots that it contains
type MultiBot []Bot

// OnMessage delivers the message to all bots and returns a response from
// any bot, that answers
func (m *MultiBot) OnMessage(msg Message) *Response {
	if contains([]string{"help", "/help", "help!"}, msg.Text) {
		return &Response{
			Text: m.Help(),
		}
	}

	wg := sync.WaitGroup{}

	texts := make(chan string, len(*m))
	var pin int32
	var unpin int32
	var preview int32
	var reply int32
	var banInterval time.Duration

	var mutex = &sync.Mutex{}

	wg.Add(len(*m))

	for _, bot := range *m {
		bot := bot
		go func() {
			if resp := bot.OnMessage(msg); resp != nil {
				texts <- resp.Text
				if resp.Pin {
					atomic.AddInt32(&pin, 1)
				}
				if resp.Preview {
					atomic.AddInt32(&preview, 1)
				}
				if resp.Unpin {
					atomic.AddInt32(&unpin, 1)
				}
				if resp.Reply {
					atomic.AddInt32(&reply, 1)
				}
				if resp.BanInterval > 0 {
					mutex.Lock()
					banInterval = resp.BanInterval
					mutex.Unlock()
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(texts)
	}()

	var lines []string
	for r := range texts {
		if strings.TrimSpace(r) == "" {
			continue
		}
		log.Printf("[DEBUG] compose %q", r)
		lines = append(lines, r)
	}

	log.Printf("[DEBUG] answers %d, send %v", len(lines), len(lines) > 0)

	resp := Response{
		Text:        strings.Join(lines, "\n"),
		Pin:         atomic.LoadInt32(&pin) > 0,
		Unpin:       atomic.LoadInt32(&unpin) > 0,
		Preview:     atomic.LoadInt32(&preview) > 0,
		Reply:       atomic.LoadInt32(&reply) > 0,
		BanInterval: banInterval,
	}

	if resp.IsEmpty() {
		return nil
	}

	return &resp
}

// Help composes help from all bots
func (m *MultiBot) Help() string {
	sb := strings.Builder{}
	for _, child := range *m {
		help := child.Help()
		if help != "" {
			// WriteString always returns nil err
			_, _ = sb.WriteString(help)
		}
	}
	return sb.String()
}

// contains returns true if the slice contains a given string
// despite any spaces at the start or end of searching string
func contains(s []string, e string) bool {
	e = strings.TrimSpace(e)
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}
