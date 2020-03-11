package cmd

// TelegramCmd runs the multibot instance over telegram
type TelegramCmd struct {
	Telegram struct {
		Token string `long:"token" env:"TOKEN" description:"telegram bot token" default:"test"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
}

func (s TelegramCmd) Execute(args []string) error {
	// todo
	panic("todo")
}
