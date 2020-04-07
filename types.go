package main;

import "time"
import "github.com/gempir/go-twitch-irc"

func check(e error) {
    if (e != nil) {
        panic(e);
    }
}

type Gist struct {
    Files struct {
        ScheduleJson struct {
            Content string `json:"content"`
        } `json:"schedule.json"`
    } `json:"files"`
}

type Config struct {
    Account  Account   `json:"account"`
    Github   Github    `json:"github"`
    Channels []Channel `json:"channels"`
};

type Account struct {
    Username   string `json:"username"`
    UserID     string `json:"userid"`
    OauthToken string `json:"oauthToken"`
    ClientID   string `json:"clientID"`
    Owner      string `json:"owner"`
};

type Github struct {
    GithubToken string `json:"oauthToken"`
    ScheduleUrl string `json:"scheduleUrl"`
};

type Channel struct {
    ChannelName string `json:"channelName"`
};

type ScheduleArray struct {
    Schedule []Schedule `json:"schedule"`
};

type Schedule struct {
    Title   string    `json:"title"`
    Twitch  string    `json:"twitch"`
    Project string    `json:"project"`
    IntTime int64     `json:"time"`
    Time    time.Time `json:"-"`
};


// Bot
type Bot struct {
    //Config    Config
    Client      *twitch.Client
    Username    string
    UserID      string
    OauthToken  string
    ClientID    string
    GithubToken string
    GistUrl     string
    Channels    []Channel
    Owner       string
    NormalMsg   [20]time.Time
    ModMsg      [100]time.Time
    PrvMsg      string
    PrvMsgIdx   int8
};
