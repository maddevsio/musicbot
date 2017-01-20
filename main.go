package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/maddevsio/musicbot/bot"
	"github.com/maddevsio/musicbot/config"
	"github.com/urfave/cli"
)

func main() {
	app := config.New()
	app.App().Action = func(c *cli.Context) error {
		bot, err := bot.NewBotAPI(app.GetConfig())
		if err != nil {
			log.Fatal(err)
		}
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, os.Kill)
		defer signal.Stop(signalChan)

		go func() {
			<-signalChan
			log.Println("signal received, stopping...")

			os.Exit(0)
		}()
		bot.Start()
		bot.WaitStop()
		return nil
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
