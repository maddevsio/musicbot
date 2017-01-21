package bot

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/musicbot/config"
	"github.com/rylio/ytdl"
	"gopkg.in/telegram-bot-api.v4"
)

type API interface {
	HandleBot(echo.Context) error
	GetAudio(echo.Context) error
}

type MusicAPI struct {
	echo       *echo.Echo
	appConfig  *config.MusicBotConfig
	yClient    *youtube.Service
	waitGroup  sync.WaitGroup
	botAPIURL  string
	bot        *tgbotapi.BotAPI
	updateChan chan tgbotapi.Update
}

func NewBotAPI(conf *config.MusicBotConfig) (*MusicAPI, error) {
	a := &MusicAPI{}
	a.echo = echo.New()
	a.appConfig = conf
	client := &http.Client{
		Transport: &transport.APIKey{Key: conf.YoutubeAPIKey},
	}
	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}
	a.yClient = service
	a.updateChan = make(chan tgbotapi.Update, 100)
	a.botAPIURL = fmt.Sprintf("/%s", conf.TelegramBotToken)
	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)

	if err != nil {
		return nil, err
	}
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(a.appConfig.TelegramWebhookURL))
	if err != nil {
		return nil, err
	}
	a.bot = bot
	a.echo.Use(middleware.Logger())
	a.echo.POST("/", a.HandleBot)
	a.echo.GET(a.botAPIURL, a.HandleBot)
	a.echo.POST(a.botAPIURL, a.HandleBot)
	return a, nil
}

func (m *MusicAPI) HandleBot(c echo.Context) error {
	var update tgbotapi.Update
	if err := c.Bind(&update); err != nil {
		return err
	}
	m.updateChan <- update
	return nil
}
func (m *MusicAPI) GetAudio(url string) (string, error) {
	info, err := ytdl.GetVideoInfo(url)
	if err != nil {
		return "", err
	}
	format := info.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)
	u, err := info.GetDownloadURL(format[0])
	if err != nil {
		return "", err
	}
	err = m.Convert(info.Title, u.String())
	if err != nil {
		return "", err
	}
	return info.Title, nil
}
func (m *MusicAPI) Convert(title, url string) error {
	fileName := fmt.Sprintf("%s.mp3", title)
	ffmpegArgs := []string{
		"-i", url,
		"-headers", "User-Agent: Go-http-client/1.1",
		"-codec:a", "libmp3lame", "-qscale:a", "2", fileName,
	}
	cmd := exec.Command("ffmpeg", ffmpegArgs...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
func (m *MusicAPI) handleRun() {
	for update := range m.updateChan {
		log.Printf("%+v\n", update)

		if update.Message != nil {
			m.onMessage(update.Message)
		}
	}
	m.waitGroup.Done()
}
func (m *MusicAPI) onMessage(message *tgbotapi.Message) {
	messageText := strings.ToLower(message.Text)
	me, err := m.bot.GetMe()
	if err != nil {
		log.Printf("Error getting myself: %s", err)
	}
	if me.UserName != message.From.UserName {
		log.Println(messageText)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Started to search")
		m.bot.Send(msg)
		call := m.yClient.Search.List("id,snippet").Q(messageText).MaxResults(50)
		resp, err := call.Do()
		if err != nil {
			log.Println(err)
		}
		var url string
		for _, i := range resp.Items {
			switch i.Id.Kind {
			case "youtube#video":
				url = fmt.Sprintf("https://youtube.com/watch?v=%s", i.Id.VideoId)
				break
			default:
				continue
			}
		}
		msg = tgbotapi.NewMessage(message.Chat.ID, "Converting")
		m.bot.Send(msg)
		m.waitGroup.Add(1)
		go func(url string) {
			title, err := m.GetAudio(url)
			if err != nil {
				log.Println(err)
			}
			msg = tgbotapi.NewMessage(message.Chat.ID, title)
			m.bot.Send(msg)
			fileName := fmt.Sprintf("%s.mp3", title)
			file, err := os.Open(fileName)
			if err != nil {
				log.Println(err)
			}
			fi, err := file.Stat()
			if err != nil {
				log.Println(err)
			}
			amsg := tgbotapi.NewAudioUpload(message.Chat.ID, fileName)
			amsg.Title = title
			amsg.Duration = 10
			amsg.MimeType = "audio/mpeg"
			amsg.FileSize = int(fi.Size())
			m.bot.Send(amsg)
			os.Remove(fileName)
			m.waitGroup.Done()
		}(url)
	}
}

func (m *MusicAPI) Start() {
	m.waitGroup.Add(1)
	go func() {
		m.echo.Start(m.appConfig.HTTPBindAddr)
		m.waitGroup.Done()
	}()
	m.waitGroup.Add(1)
	go m.handleRun()

}

func (m *MusicAPI) WaitStop() {
	m.waitGroup.Wait()
}
