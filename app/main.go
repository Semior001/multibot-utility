package main

// golangci-lint warns on the use of go-flags without alias
//noinspection GoRedundantImportAlias
import (
	"fmt"
	"log"
	"os"

	"github.com/Semior001/multibot-utility/app/cmd"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
)

// Opts describes cli arguments and flags to execute a command
type Opts struct {
	TgCmd cmd.TelegramCmd `command:"telegram"`
	Dbg   bool            `long:"dbg" env:"DEBUG" description:"turn on debug mode"`
}

const version = "unknown"

var logFlags = log.Ldate | log.Ltime

func main() {
	fmt.Printf("multibot-utility version: %s\n", version)
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)

	p.CommandHandler = func(command flags.Commander, args []string) error {
		setupLog(opts.Dbg)
		err := command.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed to execute command %v", err)
		}
		return nil
	}

	// after failure command does not return non-zero code
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func setupLog(dbg bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: "INFO",
		Writer:   os.Stdout,
	}

	if dbg {
		logFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile
		filter.MinLevel = "DEBUG"
	}

	log.SetFlags(logFlags)

	log.SetOutput(filter)
}
