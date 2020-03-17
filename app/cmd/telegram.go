package cmd

import (
	"context"
	"io/ioutil"
	"log"
	"path"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Semior001/multibotUtility/app/bot"
	"github.com/Semior001/multibotUtility/app/ctrl"
	"github.com/Semior001/multibotUtility/app/store/groups"
	bolt "github.com/coreos/bbolt"
)

// TelegramCmd runs the multibot instance over telegram
type TelegramCmd struct {
	Telegram struct {
		Token    string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		UserName string `long:"username" env:"USERNAME" description:"telegram bot username" default:"test"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
}

// Execute runs the telegram bot
func (s TelegramCmd) Execute(_ []string) error {
	loc, err := ioutil.TempDir("", "test_groups_multibot")
	if err != nil {
		log.Fatalf("failed to create temp dir at test_groups_multibot %+v", err)
	}
	svc, err := groups.NewBoltDB(path.Join(loc, "groups_bot_test.db"), bolt.Options{})
	if err != nil {
		log.Fatalf("failed to create boltdb at test_groups_multibot %+v", err)
	}
	tbapi, err := tgbotapi.NewBotAPI(s.Telegram.Token)
	if err != nil {
		log.Fatalf("failed to create boltdb at test_groups_multibot %+v", err)
	}
	t := ctrl.TelegramBotCtrl{
		Token: s.Telegram.Token,
		Bots: &bot.MultiBot{
			bot.NewGroupBot(svc, true),
		},
		TbAPI:    tbapi,
		UserName: s.Telegram.UserName,
	}
	err = t.Run(context.TODO())
	if err != nil {
		log.Fatalf("telegrambotctrl execution stopped, trace: %+v", err)
	}
	return nil
}
