package bot;

import "encoding/json"
import "fmt"
import "net/http"
import "os"
import "strings"
import "strconv"
import "time"

import "github.com/gempir/go-twitch-irc/v2"

// Config file structs

// TODO: test
func newBot() *Bot {
    config := loadConfig();
    bot := &Bot {
        Client:       twitch.NewClient(config.Account.Username, config.Account.OauthToken),
        Username:     config.Account.Username,
        UserID:       config.Account.UserID,
        OauthToken:   config.Account.OauthToken,
        ClientID:     config.Account.ClientID,
        GithubToken:  config.Github.GithubToken,
        GistUrl:      config.Github.ScheduleUrl,
        Channels:     config.Channels,
        Owner:        config.Account.Owner,
        NormalMsg:    [20]time.Time{},
        ModMsg:       [100]time.Time{},
        PrvMsg:       make(map[string]string),
        PrvMsgIdx:    0,
        Schedule:     getSchedule(config.Github.ScheduleUrl),
    };

    for i := 0; i < len(bot.Channels); i++ {
        bot.PrvMsg[bot.Channels[i].ChannelName] = "";
    }

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

    if (throttleNormalMessage(bot)) {
        return;
    }

    bot.NormalMsg[bot.PrvMsgIdx] = time.Now();
    bot.PrvMsgIdx = (bot.PrvMsgIdx + 1) % 20;

    if (message[0] == '.' || message[0] == '/') {
        message = ". " + message;
    }

    if (len(message) > 247) {
        message = message[0:247];

    }

    if (bot.PrvMsg[target] == message) {
        message += " \U000E0000";
    }

    bot.Client.Say(target, message);
    bot.PrvMsg[target] = message;
}

func throttleNormalMessage(bot *Bot) bool {

    if (bot.NormalMsg[(bot.PrvMsgIdx + 19) % 20].Add(1500 * time.Millisecond).After(time.Now())) {
        return true;
    }
    if (bot.NormalMsg[bot.PrvMsgIdx].Add(30 * time.Second).After(time.Now())) {
        return true;
    }
    return false;
}

func banUser(target string, user string, bot *Bot) {
    bot.Client.Say(target, ".ban " + user);
}

func handleMessage(message twitch.PrivateMessage, bot *Bot) {

    if (message.Message[0] == '/') {

        commandName := strings.SplitN(message.Message, " ", 2)[0][1:];

        msgLen := len(strings.SplitN(message.Message, " ", -1));

        switch commandName {
        case "":
            sendMessage(message.Channel, "Why, hello there :)", bot);

        case "regex":
            sendMessage(message.Channel, "Regex I use for chatterino: https://gist.github.com/apa420/2e1003636949f90ab440fa119c8ddcc2", bot);
        case "sch":
            sendMessage(message.Channel, "Schedule: apa420.github.io", bot);

        case "schupd":
            if (message.Tags["display-name"] == bot.Owner) {

                sortSchedule(&bot.Schedule);

                if (sendSchedule(&bot.Schedule, bot.GithubToken, bot.GistUrl,
                                 isChannelLive(message.RoomID, bot.ClientID))) {
                    sendMessage(message.Channel, "Request succeeded!", bot);
                } else {
                    sendMessage(message.Channel, "Request failed!", bot);
                }
            }

        case "schget":
            if (message.Tags["display-name"] == bot.Owner) {
                // Do http request
                bot.Schedule = getSchedule(bot.GistUrl);
                fmt.Println(bot.Schedule);

                sendMessage(message.Channel, "Sent request", bot);
            }
        case "schcle":
            if (message.Tags["display-name"] == bot.Owner) {
                cleanSchedule(&bot.Schedule);
                sendMessage(message.Channel, "Cleaning", bot);
            }
        case "schadd":
            if (message.Tags["display-name"] == bot.Owner) {
                // /update Cool test stream, apabot, 2d 2000

                if (msgLen > 1) {

                    messageStrip := strings.SplitN(message.Message, " ", 2)[1];
                    content := strings.SplitN(messageStrip, ", ", -1);

                    hour := 0;
                    second := 0;
                    if (len(content) > 2) {

                        title := content[0];
                        project := "https://github.com/apa420/" + content[1];
                        schTime := time.Now();

                        for _, element := range strings.SplitN(content[2], " ", -1) {
                            if (len(element) > 1) {
                                if (element[len(element)-1] == 'd') {


                                    days, err := strconv.Atoi(element[0:len(element)-1]);
                                    check(err);

                                    schTime = schTime.AddDate(0, 0, days);
                                } else if (len(element) == 4) {
                                    hour, _ = strconv.Atoi(element[0:2]);
                                    second, _ = strconv.Atoi(element[2:4]);
                                }
                            }
                        }
                        location, err := time.LoadLocation("Europe/Stockholm");
                        check(err);
                        schTime = time.Date(schTime.Year(),
                                            schTime.Month(),
                                            schTime.Day(),
                                            hour,
                                            second,
                                            0,
                                            0,
                                            location);

                        schedule := Schedule {
                            Title:   title,
                            Twitch:  "https://twitch.tv/apa420",
                            Project: project,
                            IntTime: schTime.UnixNano() / (1000*1000),
                            Time:    schTime,
                        };

                        if (addScheduleEntry(schedule, &bot.Schedule, bot.GithubToken,
                                             bot.GistUrl, isChannelLive(message.RoomID,
                                             bot.ClientID))) {
                            sendMessage(message.Channel, "Request succeeded!", bot);
                        } else {
                            sendMessage(message.Channel, "Request failed!", bot);
                        }
                    }
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
                if (len(strings.SplitN(message.Message, " ", 1)) > 0) {
                    sendMessage(message.Channel, strings.SplitN(message.Message, " ", -1)[0], bot);
                } else {
                    sendMessage(message.Channel, "Can't return empty string", bot);
                }

            }

        case "tf":
            sendMessage(message.Channel, ":tf:", bot);

        case "bot":
            sendMessage(message.Channel, "Bot made by apa420 written in Golang repo: https://github.com/apa420/apabot", bot);

        case "ban":
            if (len(strings.SplitN(message.Message, " ", 2)) > 1) {
                if (message.Tags["display-name"] == bot.Owner) {
                    banUser(message.Channel, strings.SplitN(message.Message, " ", 2)[1], bot);
                }
                sendMessage(message.Channel, strings.SplitN(message.Message, " ", 2)[1] + " has been banned!", bot);
            } else {
                sendMessage(message.Channel, message.Tags["display-name"] + " has been banned!", bot);
            }

        case "dank":
            if (len(strings.SplitN(message.Message, " ", 2)) > 1) {
                sendMessage(message.Channel, strings.SplitN(message.Message, " ", 2)[1] + " FeelsDankMan", bot);
            } else {
                sendMessage(message.Channel, message.Tags["display-name"] + " FeelsDankMan", bot);
            }

        default:
            sendMessage(message.Channel, "Soon ™", bot);
        }
    }
}


func Run() {
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
