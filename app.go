package main

import (
	"crypto/tls"
	"encoding/json"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
)

type Config struct {
	Tg_Apikey    string `json:"tg_apikey"`
	Tg_groupid   int64  `json:"tg_groupid"`
	Irc_Hostname string `json:"irc_hostname"`
	Irc_Channel  string `json:"irc_channel"`
	Irc_Nick     string `json:"irc_nick"`
	Admin        string `json:"admin"`
}

func (config *Config) load() {
	dat, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panic(err)
	}

	json.Unmarshal(dat, &config)
}

func main() {
	config := Config{}
	config.load()

	irc_chan := make(chan string, 100)
	tg_chan := make(chan string, 100)

	// telegram
	bot, err := tgbotapi.NewBotAPI(config.Tg_Apikey)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go func() {
		updates, err := bot.GetUpdatesChan(u)

		if err != nil {
			log.Panic(err)
		}

		for update := range updates {
			log.Printf("[%s] %s ", update.Message.From.UserName, update.Message.Text)
			log.Printf("%d", update.Message.Chat.ID)

			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			if update.Message.Chat.ID != config.Tg_groupid {
				continue
			}

			tg_chan <- "<" + update.Message.From.UserName + "> " + update.Message.Text

		}
	}()

	// irc
	irccon := irc.IRC(config.Irc_Nick, "IRCTestSSL")
	irccon.VerboseCallbackHandler = true
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(config.Irc_Channel) })

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			log.Printf("%s", event.Message())
			log.Printf("%s", event.Nick)
			log.Printf("%s", event.Arguments[0])
			irc_chan <- "<" + event.Nick + "> " + event.Message()
		}(event)
	})

	err = irccon.Connect(config.Irc_Hostname)

	if err != nil {
		log.Panic(err)
	}

	go func(irc_chan <-chan string) {
		for msg := range irc_chan {
			bot.Send(tgbotapi.NewMessage(config.Tg_groupid, msg))
		}
	}(irc_chan)

	go func(tg_chan <-chan string) {
		for msg := range tg_chan {
			irccon.Privmsg(config.Irc_Channel, msg)
		}
	}(tg_chan)

	irccon.Loop()
}
