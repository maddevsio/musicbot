package config

import (
	"os"

	"github.com/gen1us2k/log"
	"github.com/urfave/cli"
)

type (
	MusicBotConfig struct {
		TelegramBotToken   string
		TelegramWebhookURL string
		HTTPBindAddr       string
		YoutubeAPIKey      string
	}
	Configuration struct {
		data *MusicBotConfig
		app  *cli.App
	}
)

// Version stores current service version
var (
	Version            string
	TelegramBotToken   string
	TelegramWebhookURL string
	HTTPBindAddr       string
	YoutubeAPIKey      string
	LogLevel           string
)

// New is constructor and creates a new copy of Configuration
func New() *Configuration {
	Version = "0.1dev"
	app := cli.NewApp()
	app.Name = "Music bot"
	app.Usage = "Send music from youtube to telegram"
	return &Configuration{
		data: &MusicBotConfig{},
		app:  app,
	}
}

func (c *Configuration) fillConfig() *MusicBotConfig {
	return &MusicBotConfig{
		TelegramBotToken:   TelegramBotToken,
		TelegramWebhookURL: TelegramWebhookURL,
		HTTPBindAddr:       HTTPBindAddr,
		YoutubeAPIKey:      YoutubeAPIKey,
	}
}

// Run is wrapper around cli.App
func (c *Configuration) Run() error {
	c.app.Before = func(ctx *cli.Context) error {
		log.SetLevel(log.MustParseLevel(LogLevel))
		return nil
	}
	c.app.Flags = c.setupFlags()
	return c.app.Run(os.Args)
}

// App is public method for Configuration.app
func (c *Configuration) App() *cli.App {
	return c.app
}

func (c *Configuration) setupFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "http_bind_addr",
			Value:       ":8090",
			Usage:       "Set address to bind http server",
			EnvVar:      "HTTP_BIND_ADDR",
			Destination: &HTTPBindAddr,
		},
		cli.StringFlag{
			Name:        "youtube_api_key",
			Value:       "Aiza...",
			Usage:       "Set youtube api key",
			EnvVar:      "YOUTUBE_API_KEY",
			Destination: &YoutubeAPIKey,
		},
		cli.StringFlag{
			Name:        "telegram_bot_token",
			Value:       "",
			Usage:       "Set telegram bot access token",
			EnvVar:      "TELEGRAM_TOKEN",
			Destination: &TelegramBotToken,
		},
		cli.StringFlag{
			Name:        "telegram_web_hook_url",
			Value:       "",
			Usage:       "Set telegram bot webhook url",
			EnvVar:      "TELEGRAM_WEBHOOK_URL",
			Destination: &TelegramWebhookURL,
		},
		cli.StringFlag{
			Name:        "loglevel",
			Value:       "debug",
			Usage:       "set log level",
			Destination: &LogLevel,
			EnvVar:      "LOG_LEVEL",
		},
	}

}

// GetConfig returns filled MusicBotConfig
func (c *Configuration) GetConfig() *MusicBotConfig {
	c.data = c.fillConfig()
	return c.data
}
