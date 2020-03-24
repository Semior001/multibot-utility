package ctrl

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/Semior001/multibotUtility/app/bot"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

// TelegramBotCtrl is an implementation of bot ctrl
// to execute bot commands in the Telegram messenger
type TelegramBotCtrl struct {
	Token    string
	Bots     bot.Bot
	TbAPI    *tgbotapi.BotAPI
	UserName string
}

// Run starts bots to listen for messages
func (t *TelegramBotCtrl) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := t.TbAPI.GetUpdatesChan(u)
	if err != nil {
		return errors.Wrap(err, "failed to start telegram bot listener")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return errors.New("telegram updates chan closed")
			}
			if update.Message == nil { // ignore any non-message updates
				continue
			}
			if update.Message.Chat == nil { // ignore messages not from chat
				continue
			}
			if update.Message.Text == "" { // ignore messages without text
				continue
			}

			fromChat := update.Message.Chat.ID
			msg := t.convertMessage(update.Message)

			log.Printf("[DEBUG] incoming msg: %+v", msg)

			resp := t.Bots.OnMessage(msg)

			if err := t.sendBotResponse(resp, fromChat); err != nil {
				log.Printf("[WARN] failed to respond on update, %v", err)
			}
		}
	}
}

// sendBotResponse sends bot's answer to tg channel and saves it to log
func (t *TelegramBotCtrl) sendBotResponse(resp *bot.Response, chatID int64) error {
	if resp == nil {
		return nil
	}

	log.Printf("[DEBUG] bot response - %+v, pin: %t", resp.Text, resp.Pin)
	tbMsg := tgbotapi.NewMessage(chatID, resp.Text)
	tbMsg.ParseMode = tgbotapi.ModeMarkdown
	tbMsg.DisableWebPagePreview = !resp.Preview
	res, err := t.TbAPI.Send(tbMsg)
	if err != nil {
		return errors.Wrapf(err, "can't send message to telegram %q", resp.Text)
	}

	if resp.Pin {
		_, err = t.TbAPI.PinChatMessage(tgbotapi.PinChatMessageConfig{
			ChatID:              chatID,
			MessageID:           res.MessageID,
			DisableNotification: true,
		})
		if err != nil {
			return errors.Wrap(err, "can't pin message to telegram")
		}
	}

	if resp.Unpin {
		_, err = t.TbAPI.UnpinChatMessage(tgbotapi.UnpinChatMessageConfig{ChatID: chatID})
		if err != nil {
			return errors.Wrap(err, "can't unpin message to telegram")
		}
	}

	return nil
}

// convertMessage transforms a telegram message into internal struct
func (t *TelegramBotCtrl) convertMessage(msg *tgbotapi.Message) bot.Message {
	res := bot.Message{
		ID:     strconv.Itoa(msg.MessageID),
		ChatID: strconv.FormatInt(msg.Chat.ID, 10),
		Sent:   time.Unix(int64(msg.Date), 0),
		Text:   msg.Text,
	}

	if msg.From != nil {
		res.From = &bot.User{
			ID:          strconv.Itoa(msg.From.ID),
			Username:    msg.From.UserName,
			DisplayName: msg.From.FirstName + " " + msg.From.LastName,
			IsBot:       msg.From.IsBot,
		}

		res.From.IsAdmin = t.isUserAdmin(msg)
	}

	// checking that it is a bot addition
	if t.isBotAddedToChat(msg.NewChatMembers) {
		res.AddedBotToChat = true
	}

	return res
}

// isUserAdmin detects on the message data is a sender of message an admin
// todo blocking call
func (t *TelegramBotCtrl) isUserAdmin(msg *tgbotapi.Message) bool {
	// if all members are admins - we do not have to get list of all users
	if msg.Chat.AllMembersAreAdmins {
		return true
	}

	admins, err := t.TbAPI.GetChatAdministrators(tgbotapi.ChatConfig{
		ChatID: msg.Chat.ID,
	})

	if err != nil {
		log.Printf("[WARN] failed to retrieve admins for chat %d: %+v", msg.Chat.ID, err)
		return false
	}

	// if not all members are admins, then we check users contained in the list of users
	if chatMemberContainsUsername(admins, msg.From.UserName) {
		return true
	}

	return false
}

// chatMemberContainsUsername checks is the username in slice of users
func chatMemberContainsUsername(members []tgbotapi.ChatMember, username string) bool {
	for _, u := range members {
		if u.User == nil {
			continue
		}
		if u.User.UserName == username {
			return true
		}
	}
	return false
}

// isBotAddedToChat checks that this bot was added to the new chat
func (t *TelegramBotCtrl) isBotAddedToChat(newMembers *[]tgbotapi.User) bool {
	if newMembers == nil {
		return false
	}

	for _, u := range *newMembers {
		if u.IsBot && u.UserName == t.UserName {
			return true
		}
	}

	return false
}
