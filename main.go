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

// Config represents main config for app
type Config struct {
	ReleBoard    ReleBoard
	AllowedUsers []string
}

// Toggle rele state
func (r *Rele) Toggle() {
	pin := rpio.Pin(r.Pin)
	pin.Toggle()
	log.Printf("%s: state: %+v", r.Name, pin.Read())
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

var config Config

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
		r := r // hack to avoid issue
		if r.SwitchOnAt != "" {
			log.Printf("schedule switch on at %s for rele %s\n", r.SwitchOnAt, r.Name)
			hour, minute, err := parseTime(r.SwitchOnAt)
			if err != nil {
				return errors.Wrapf(err, "failed to parse time %s", r.SwitchOnAt)
			}
			if err := c.AddFunc(
				fmt.Sprintf("0 %d %d * * *", minute, hour),
				func() {
					r.SwitchOn()
					log.Printf("Cron: switch on rele %s at %s\n", r.Name, r.SwitchOnAt)
				},
			); err != nil {
				return errors.Wrapf(err, "failed to setup cron job for %s", r.Name)
			}
		}
		if r.SwitchOffAt != "" {
			log.Printf("schedule switch off at %s for rele %s\n", r.SwitchOffAt, r.Name)
			hour, minute, err := parseTime(r.SwitchOffAt)
			if err != nil {
				return errors.Wrapf(err, "failed to parse time %s", r.SwitchOffAt)
			}
			if err := c.AddFunc(
				fmt.Sprintf("0 %d %d * * *", minute, hour),
				func() {
					r.SwitchOff()
					log.Printf("Cron: switch off rele %s at %s\n", r.Name, r.SwitchOffAt)
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

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		isAllowed := false
		for _, user := range config.AllowedUsers {
			if user == update.Message.From.UserName {
				isAllowed = true
			}
		}
		if !isAllowed {
			continue
		}
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

// parse string "05:23" as hour = 5, minute = 23
func parseTime(s string) (hour, minute int64, err error) {
	time := strings.Split(s, ":")
	hour, err = strconv.ParseInt(time[0], 10, 64)
	if err != nil {
		return
	}
	minute, err = strconv.ParseInt(time[1], 10, 64)
	if err != nil {
		return
	}
	return
}
