package main

import (
	"log"
	"os"

	"gopkg.in/telegram-bot-api.v4"

	"github.com/pkg/errors"
	rpio "github.com/stianeikeland/go-rpio"
)

const (
	rele1Pin = 15
	rele2Pin = 14 // don't know why first and second are swapped
	rele3Pin = 18
	rele4Pin = 23
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}

func run() error {
	if err := setup(); err != nil {
		return errors.Wrap(err, "failed to setup GPIO pins")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			case "switch":
				rpio.Pin(rele1Pin).Toggle()
				msg.Text = "done"
			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}
	}
	return nil
}

func setup() error {
	if err := rpio.Open(); err != nil {
		return errors.Wrap(err, "failed to init rpio")
	}

	for _, i := range []int{rele1Pin, rele2Pin, rele3Pin, rele4Pin} {
		pin := rpio.Pin(i)
		pin.Output()
		pin.High()
	}

	return nil
}
