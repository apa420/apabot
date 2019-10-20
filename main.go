package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gempir/go-twitch-irc"
)

// Config file structs
type Config struct {
	Account  Account   `json:"account"`
	Channels []Channel `json:"channels"`
}

type Account struct {
	Username   string `json:"username"`
	UserID     string `json:"userid"`
	OauthToken string `json:"oauthToken"`
	Owner      string `json:"owner"`
}

type Channel struct {
	ChannelName string `json:"channelName"`
}

// Bot
type Bot struct {
	//Config   Config
	Client     *twitch.Client
	Username   string
	UserID     string
	OauthToken string
	Channels   []Channel
	Owner      string
	UserBucket int
	ModBucket  int
}

// TODO: test
func newBot() *Bot {
	config := loadConfig()
	bot := &Bot{
		Client:     twitch.NewClient(config.Account.Username, config.Account.OauthToken),
		Username:   config.Account.Username,
		UserID:     config.Account.UserID,
		OauthToken: config.Account.OauthToken,
		Channels:   config.Channels,
		Owner:      config.Account.Owner,
		UserBucket: 20,
		ModBucket:  100,
	}
	return bot
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadConfig() Config {
	jsonFile, err := os.Open("config.json")
	check(err)
	defer jsonFile.Close()

	var config Config
	json.NewDecoder(jsonFile).Decode(&config)

	return config
}

func connectToChannels(client *twitch.Client, channels []Channel) {
	for i := 0; i < len(channels); i++ {
		client.Join(channels[i].ChannelName)
		client.Say(channels[i].ChannelName, ":)")
	}
}

func sendMessage(target string, message string, client *twitch.Client) {
	if message[0] == '.' || message[0] == '/' {
		client.Say(target, "Don't use commands")
		return
	}
	client.Say(target, message)
}

func sendOwnerMessage(target string, message string, client *twitch.Client) {
	client.Say(target, message)
}

func handleMessage(message twitch.PrivateMessage, bot *Bot) {
	if bot.UserBucket == 0 || bot.ModBucket == 0 {
		return
	}
	bot.UserBucket--
	if message.Action && message.Tags["display-name"] == bot.Owner {
		sendOwnerMessage(message.Channel, ".me monkaS 🚨 ALERT", bot.Client)
	}
}

func main() {
	bot := newBot()

	connectToChannels(bot.Client, bot.Channels)

	bot.Client.OnConnect(func() {
		fmt.Println("connected")
	})

	bot.Client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		channelID := message.Tags["room-id"]
		if channelID == "" {
			fmt.Printf("Missing room-id tag in message")
			return
		}

		if message.Tags["user-id"] == bot.UserID {
			return
		}
		handleMessage(message, bot)

	})

	fmt.Println("Finished loading")
	// TODO: Start bucket system
	/*go func() {
		t := time.NewTimer(30 * time.Second)
		for {
			<-t.C
			bot.UserBucket = 30
			bot.ModBucket = 100

			t.Reset(30 * time.Second)
		}
	}()*/
	err := bot.Client.Connect()
	check(err)
}
