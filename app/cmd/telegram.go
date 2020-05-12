package cmd

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Semior001/multibot-utility/app/bot"
	"github.com/Semior001/multibot-utility/app/ctrl"
	"github.com/Semior001/multibot-utility/app/store/groups"
	bolt "github.com/coreos/bbolt"
)

// TelegramCmd runs the multibot instance over telegram
type TelegramCmd struct {
	Telegram struct {
		Token    string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
		UserName string `long:"username" env:"USERNAME" description:"telegram bot username" default:"test"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Db struct {
		Location string `long:"location" env:"LOCATION" description:"location of boltdb sotrage" required:"true"`
	} `group:"db" namespace:"db" env-namespace:"DB"`
}

// Execute runs the telegram bot
func (s TelegramCmd) Execute(_ []string) error {
	svc, err := groups.NewBoltDB(s.Db.Location, bolt.Options{})
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
			bot.NewGroupBot(bot.GroupBotParams{
				Store:              svc,
				RespondAllCommands: true,
			}),
		},
		API:      tbapi,
		UserName: s.Telegram.UserName,
	}
	err = t.Run(context.TODO())
	if err != nil {
		log.Fatalf("telegrambotctrl execution stopped, trace: %+v", err)
	}
	return nil
}
