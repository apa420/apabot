package main;

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gempir/go-twitch-irc"
);

// Config file structs

// TODO: test
func newBot() *Bot {
    config := loadConfig();
    bot := &Bot {
        Client:      twitch.NewClient(config.Account.Username, config.Account.OauthToken),
        Username:    config.Account.Username,
        UserID:      config.Account.UserID,
        OauthToken:  config.Account.OauthToken,
        ClientID:    config.Account.ClientID,
        GithubToken: config.Github.GithubToken,
        GistUrl:     config.Github.ScheduleUrl,
        Channels:    config.Channels,
        Owner:       config.Account.Owner,
        NormalMsg:   [20]time.Time{},
        ModMsg:      [100]time.Time{},
        PrvMsg:      "",
        PrvMsgIdx:   0,
    };
    return bot;
};

func loadConfig() Config {
    jsonFile, err := os.Open("config.json");
    check(err);
    defer jsonFile.Close();

    var config Config;
    json.NewDecoder(jsonFile).Decode(&config);

    return config;
}

func isChannelLive(channelID string, clientID string) bool {

    client := &http.Client{};
    req, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/streams/"+channelID, nil);
    check(err);
    req.Header.Set("Accept", "application/vnd.twitchtv.v5+json");
    req.Header.Set("Client-ID", clientID);
    check(err);

    resp, err := client.Do(req);
    check(err);

    defer resp.Body.Close();

    type T struct {
        Stream interface{};
    }
    var t T;
    err = json.NewDecoder(resp.Body).Decode(&t);
    check(err);
    return t.Stream != nil;
}

func connectToChannels(client *twitch.Client, channels []Channel) {
    for i := 0; i < len(channels); i++{
        client.Join(channels[i].ChannelName);
        client.Say(channels[i].ChannelName, ":)");
    }
}

func sendMessage(target string, message string, bot *Bot) {
    if (message[0] == '.' || message[0] == '/') {
        message = ". " + message;
    }
    if (len(message) > 247) {
        message = message[0:247];
    }
    if (bot.PrvMsg == message) {
        message += " \U000E0000";
    }
    bot.Client.Say(target, message);
    bot.PrvMsg = message;
}

func banUser(target string, user string, bot *Bot) {
    bot.Client.Say(target, ".ban " + user);
}

func handleMessage(message twitch.PrivateMessage, bot *Bot) {
    if (message.Message[0] == '/') {

        if (throttleNormalMessage(bot)) {
            return;
        }

        bot.NormalMsg[bot.PrvMsgIdx] = time.Now();
        bot.PrvMsgIdx = (bot.PrvMsgIdx + 1) % 20;

        commandName := strings.SplitN(message.Message, " ", 2)[0][1:];

        switch commandName {
        case "":
            sendMessage(message.Channel, "Why, hello there :)", bot);
        case "get":
            if (message.Tags["display-name"] == bot.Owner) {
                // Do http request
                scheduleArray := getSchedule(bot.GistUrl);
                fmt.Println(scheduleArray);

                sendMessage(message.Channel, "Sent request", bot);
            }
        case "update":
            if (message.Tags["display-name"] == bot.Owner) {
                // Test schedule message
                schedule := Schedule {
                    Title:   "This is a nice test title",
                    Twitch:  "https://twitch.tv/apa420",
                    Project: "https://github.com/apa420/apabot",
                    IntTime: time.Now().UnixNano() / (1000*1000),
                    Time:    time.Now(),
                };

                if (updateSchedule(schedule, bot.GistUrl, bot.GithubToken, bot.GistUrl)) {
                    sendMessage(message.Channel, "Request succeeded!", bot);
                } else {
                    sendMessage(message.Channel, "Request failed!", bot);
                }

            }
case "isLive":
            if (message.Tags["display-name"] == bot.Owner) {
                if (isChannelLive(message.RoomID, bot.ClientID)) {
                    sendMessage(message.Channel, "Live :)", bot);
                } else {
                    sendMessage(message.Channel, "Offline :/", bot);
                }
            }
        case "echo":
            if (message.Tags["display-name"] == bot.Owner) {
                if (len(strings.SplitN(message.Message, " ", 2)) > 1) {
                    sendMessage(message.Channel, strings.SplitN(message.Message, " ", 2)[1], bot);
                } else {
                    sendMessage(message.Channel, "Can't return empty string", bot);
                }

            }
        case "tf":
            sendMessage(message.Channel, ":tf:", bot);
        case "help":
            sendMessage(message.Channel, "Bot made by apa420 https://github.com/apa420/apabot", bot);
        case "ban":
            if (len(strings.SplitN(message.Message, " ", 2)) > 1) {
                if (message.Tags["display-name"] == bot.Owner) {
                    banUser(message.Channel, strings.SplitN(message.Message, " ", 2)[1], bot);
                }
                sendMessage(message.Channel, strings.SplitN(message.Message, " ", 2)[1] + " has been banned!", bot);
            } else {
                sendMessage(message.Channel, message.Tags["display-name"] + " has been banned!", bot);
            }
        default:
            sendMessage(message.Channel, "Soon ™", bot);
        }
    }
}

func throttleNormalMessage(bot *Bot) bool {

    if (bot.NormalMsg[(bot.PrvMsgIdx+19)%20].Add(1500 * time.Millisecond).After(time.Now())) {
        return true;
    }
    if (bot.NormalMsg[bot.PrvMsgIdx].Add(30 * time.Second).After(time.Now())) {
        return true;
    }
    return false;
}

func main() {
    bot := newBot();

    connectToChannels(bot.Client, bot.Channels);

    bot.Client.OnConnect(func() {
        fmt.Println("connected");
    })

    bot.Client.OnPrivateMessage(func(message twitch.PrivateMessage) {
        channelID := message.Tags["room-id"];
        if (channelID == "") {
            fmt.Printf("Missing room-id tag in message");
            return;
        }

        if (message.Tags["user-id"] == bot.UserID) {
            return;
        }
        handleMessage(message, bot);

    })

    fmt.Println("Finished loading");
    err := bot.Client.Connect();
    check(err);
}
