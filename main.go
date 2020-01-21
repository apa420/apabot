package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gempir/go-twitch-irc"
	//		"net/http"
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
	//Config		Config
	Client     *twitch.Client
	Username   string
	UserID     string
	OauthToken string
	Channels   []Channel
	Owner      string
	NormalMsg  [20]time.Time
	ModMsg     [100]time.Time
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
		NormalMsg:	[20]time.Time{},
		ModMsg:		[100]time.Time{},
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

func sendMessage(target string, message string, bot *Bot) {
	if !throttleMessage(bot){
		bot.Client.Say(target, message)
	}
}

func handleMessage(message twitch.PrivateMessage, bot *Bot) {
	if message.Action && message.Tags["display-name"] == bot.Owner {
		bot.NormalMsg = append([]time{Time.now()}, bot.NormalMsg[1:19])
		//bot.NormalMsg = append([]time.Time.now(), bot.NormalMsg...)
		//bot.NormalMsg = [time.Time.now()] + bot.NormalMsg[1:20]
		//bot.NormalMsg = append(bot.NormalMsg[1:20],
		//bot.NormalMsg = append(bot.NormalMsg, time.Time{time.Now()})
		sendMessage(message.Channel, ".me monkaS 🚨 ALERT", bot)
	}
}

func inBetweenTimes(start, end, check time.Time) bool{
	return check.After(start) && check.Before(end)
}

func throttleMessage(bot *Bot)(bool){
	if inBetweenTimes(bot.NormalMsg[19], bot.NormalMsg[19].Add(30*time.Second), time.Now()){
		return true
	}
	return false
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
	err := bot.Client.Connect()
	check(err)
}
