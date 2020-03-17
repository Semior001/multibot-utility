package bot

import (
	"context"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-pkgz/syncs"
)

//go:generate mockery -inpkg -name Bot -case snake

// User defines user info of the Message
type User struct {
	ID          string
	Username    string // might be "" for users without usernames
	DisplayName string
	IsAdmin     bool
}

// Message to pass data from/to bot
type Message struct {
	ID             string
	ChatID         string
	From           User
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

// Bot describes a particular bot, that reacts on messages and sends whatever
type Bot interface {
	OnMessage(msg Message) *Response // nil if nothing to send
	Help() string                    // returns help message - how to use this bot
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

	wg := syncs.NewSizedGroup(4)

	texts := make(chan string)
	var pin int32
	var unpin int32
	var preview int32
	var reply int32
	var banInterval time.Duration = 0

	var mutex = &sync.Mutex{}

	for _, bot := range *m {
		// if we pass the variable into the goroutine, all goroutines
		// will take the same variable, bot is a SINGLE variable
		bot := bot
		// declaring goroutine
		wg.Go(func(ctx context.Context) {
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
		})
	}

	go func() {
		wg.Wait()
		close(texts)
	}()

	var lines []string
	for r := range texts {
		log.Printf("[DEBUG] compose %q", r)
		lines = append(lines, r)
	}

	log.Printf("[DEBUG] answers %d, send %v", len(lines), len(lines) > 0)

	if len(lines) <= 0 {
		return nil
	}

	return &Response{
		Text:        strings.Join(lines, "\n"),
		Pin:         atomic.LoadInt32(&pin) > 0,
		Unpin:       atomic.LoadInt32(&unpin) > 0,
		Preview:     atomic.LoadInt32(&preview) > 0,
		Reply:       atomic.LoadInt32(&reply) > 0,
		BanInterval: banInterval,
	}
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
