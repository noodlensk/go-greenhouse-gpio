package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"gopkg.in/telegram-bot-api.v4"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
	rpio "github.com/stianeikeland/go-rpio"
)

// Rele represents one rele on rele board
type Rele struct {
	Name        string
	Pin         int
	On          bool
	Scheduled   bool
	SwitchOnAt  string // 13:00 format
	SwitchOffAt string
}

// ReleBoard is set of Rele
type ReleBoard []*Rele

// Toggle rele state
func (r *Rele) Toggle() {
	pin := rpio.Pin(r.Pin)
	pin.Toggle()
	log.Printf("state: %+v", pin.Read())
	if pin.Read() == rpio.High {
		r.On = false
	} else {
		r.On = true
	}
}

// SwitchOn enables power on rele
func (r *Rele) SwitchOn() {
	pin := rpio.Pin(r.Pin)
	pin.Low()
	if pin.Read() == rpio.High {
		r.On = false
	} else {
		r.On = true
	}
}

// SwitchOff disables power on rele
func (r *Rele) SwitchOff() {
	pin := rpio.Pin(r.Pin)
	pin.High()
	if pin.Read() == rpio.High {
		r.On = false
	} else {
		r.On = true
	}
}

var myReleBoard ReleBoard
var config struct {
	ReleBoard ReleBoard
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}

func run() error {
	viper.SetConfigName("config") // name of config file (without extension)
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	viper.AddConfigPath(dir)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	if err = viper.Unmarshal(&config); err != nil {
		log.Fatalf("failed to scan config to obj: %v", err)
	}
	myReleBoard = config.ReleBoard
	if err := setupRpio(); err != nil {
		return errors.Wrap(err, "failed to setupRpio GPIO pins")
	}
	c := cron.New()
	for _, r := range myReleBoard {
		if r.SwitchOnAt != "" {
			time := strings.Split(r.SwitchOnAt, ":")
			hour, err := strconv.ParseInt(time[0], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "failed to read SwitchOnAt for %s", r.Name)
			}
			minute, err := strconv.ParseInt(time[1], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "failed to read SwitchOnAt for %s", r.Name)
			}
			if err := c.AddFunc(
				fmt.Sprintf("0 %d %d * * *", minute, hour),
				func() {
					r.SwitchOn()
					log.Printf("Running cron job \n")
				},
			); err != nil {
				return errors.Wrapf(err, "failed to setup cron job for %s", r.Name)
			}
		}
	}
	c.Start()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		return errors.Wrap(err, "failed to init telegram bot API")
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

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "switch":
				found := false
				for _, r := range myReleBoard {
					if r.Name == update.Message.CommandArguments() {
						r.Toggle()
						log.Printf("Toggle: %+v", r)
						found = true
						break
					}
				}
				if found {
					msg.Text = "Done"
				} else {
					msg.Text = "Unknown rele ID"
				}
			case "status":
				for _, r := range myReleBoard {
					msg.Text += fmt.Sprintf("Name: %+v\n", r)
				}
			default:
				msg.Text = "I don't know that command"
			}

			bot.Send(msg)
		}
	}
	return nil
}

func setupRpio() error {
	if err := rpio.Open(); err != nil {
		return errors.Wrap(err, "failed to init rpio")
	}

	for _, r := range myReleBoard {
		pin := rpio.Pin(r.Pin)
		pin.Output()
		pin.High()
	}

	return nil
}
