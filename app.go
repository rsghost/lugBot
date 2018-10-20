package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thoj/go-ircevent"
	"log"
	"crypto/tls"
	"time"
	"encoding/json"
	"io/ioutil"
)

const channel = "#ctlug.tw.test";

type Config struct {
	Apikey		string `json:"apikey"`
	Hostname	string `json:"hostname"`
	Nick		string `json:"nick"`
	Admin		string `json:"admin"`
}

func (config *Config) load() {
	dat, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panic(err)
	}

	json.Unmarshal(dat, &config)
}

func ircbot(config Config) {
        irccon := irc.IRC(config.Nick, "IRCTestSSL")
        irccon.VerboseCallbackHandler = true
        irccon.Debug = true
        irccon.UseTLS = true
        irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
        irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		log.Printf("%s", event.Message())
		log.Printf("%s", event.Nick)
		log.Printf("%s", event.Arguments[0])

		irccon.Privmsg(channel, event.Message())
	});

        err := irccon.Connect(config.Hostname)
	if err != nil {
		log.Panic(err)
		return
	}

        irccon.Loop()
}

func tg(config Config) {
	bot, err := tgbotapi.NewBotAPI(config.Apikey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.From.UserName != config.Admin {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}

func main() {
	config := Config{}
	config.load()

	go ircbot(config)
	go tg(config)

	time.Sleep(600 * time.Second)
}
