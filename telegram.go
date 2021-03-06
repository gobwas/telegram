package telegram

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/telegram-bot-api.v2"
	"net/http"
	"net/url"
)

const (
	InlineQueryResultArticleType = "article"
	ParseModeMarkdown            = "Markdown"
	ParseModeHTML                = "HTML"
)

type Config struct {
	Token   string
	Debug   bool
	Polling *Polling
	WebHook *WebHook
}

type Polling struct {
	Offset  int
	Timeout int
}

type WebHook struct {
	URL    url.URL
	Cert   string
	Listen Listen
}

type TLS struct {
	Key  string
	Cert string
}

type Listen struct {
	Addr string
	Port int
	TLS  *TLS
}

type Application struct {
	Router
	bot    *tgbotapi.BotAPI
	config Config
}

func New(config Config) (app *Application, err error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return
	}
	bot.Debug = config.Debug

	if config.WebHook == nil && config.Polling == nil {
		return nil, fmt.Errorf("telegram: could not listen for updates: polling or webhook config fields should be set")
	}

	return &Application{
		bot:    bot,
		config: config,
	}, nil
}

func (self *Application) Bot() *tgbotapi.BotAPI {
	return self.bot
}

func (self *Application) Listen() error {
	var updates <-chan tgbotapi.Update
	fatal := make(chan error)

	if self.config.WebHook != nil {
		c := self.config.WebHook
		webHookConfig := tgbotapi.WebhookConfig{URL: &c.URL}
		if c.Cert != "" {
			webHookConfig.Certificate = c.Cert
		}
		if _, err := self.bot.SetWebhook(webHookConfig); err != nil {
			return err
		}

		updates = self.bot.ListenForWebhook("/" + c.URL.Path)
		go func() {
			addr := fmt.Sprintf("%s:%d", c.Listen.Addr, c.Listen.Port)
			if c.Listen.TLS != nil {
				fatal <- http.ListenAndServeTLS(addr, c.Listen.TLS.Cert, c.Listen.TLS.Key, nil)
			} else {
				fatal <- http.ListenAndServe(addr, nil)
			}
		}()
	} else if self.config.Polling != nil {
		c := self.config.Polling
		u := tgbotapi.NewUpdate(c.Offset)
		u.Timeout = c.Timeout

		if _, err := self.bot.RemoveWebhook(); err != nil {
			return err
		}

		ch, err := self.bot.GetUpdatesChan(u)
		if err != nil {
			return err
		}

		updates = ch
	} else {
		return errors.New("malformed configuration: Polling or Webhook directives are required")
	}

	for {
		select {
		case err := <-fatal:
			return err
		case update := <-updates:
			ctx := context.Background()
			go self.HandleUpdate(ctx, self.bot, update)
		}
	}

	return nil
}
