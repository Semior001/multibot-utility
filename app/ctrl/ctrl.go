package ctrl

import "github.com/Semior001/multibotUtility/app/bot"

// BotController describes general behavior of controllers
// to execute several bots simultaneously in any messenger
type BotController interface {
	DisableBot(bot bot.Bot)
	AddBot(bot bot.Bot)

	Run()
}
