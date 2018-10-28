package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"time"
)

const channel = "#ctlug.tw.test"

type Config struct {
	Tg_Apikey	string `json:"tg_apikey"`
	Tg_groupid	string `json:"tg_groupid"`
	Irc_Hostname	string `json:"irc_hostname"`
	Irc_Nick	string `json:"irc_nick"`
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
	irccon := irc.IRC(config.Irc_Nick, "IRCTestSSL")
	irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			log.Printf("%s", event.Message())
			log.Printf("%s", event.Nick)
			log.Printf("%s", event.Arguments[0])

			irccon.Privmsg(channel, event.Message())
		}(event)
	})

	err := irccon.Connect(config.Irc_Hostname)
	if err != nil {
		log.Panic(err)
		return
	}

	irccon.Loop()
}

func tg(config Config) {
	bot, err := tgbotapi.NewBotAPI(config.Tg_Apikey)
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
