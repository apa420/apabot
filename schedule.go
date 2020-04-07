package main;

import "net/http"
import "encoding/json"
import "time"
import "fmt"
import "io/ioutil"
import "bytes"
import "strings"

func getSchedule(gistUrl string) ScheduleArray {
    resp, err := http.Get(gistUrl);
    check(err);

    defer resp.Body.Close();

    var scheduleArray ScheduleArray;

    err = json.NewDecoder(resp.Body).Decode(&scheduleArray);
    check(err);

    for i := 0; i < len(scheduleArray.Schedule); i++ {
        scheduleArray.Schedule[i].Time = time.Unix(
            0,
            scheduleArray.Schedule[i].IntTime * (1000*1000));
    }
    return scheduleArray;
}

func updateSchedule(scheduleAddition Schedule, url string, githubOAuth string, gistUrl string) bool {

    // Get old schedule
    scheduleArray := getSchedule(gistUrl);
    scheduleArray.Schedule = append(scheduleArray.Schedule, scheduleAddition);

    gist := Gist{};
    content, err := json.MarshalIndent(scheduleArray, "", " ");
    gist.Files.ScheduleJson.Content = string(content);


    //buffer, err := json.Marshal(&gist);
    //check(err);
    buffer, err := json.MarshalIndent(&gist, "", " ");
    check(err);

    reader := bytes.NewReader(buffer);


    client := &http.Client{};
    req, err := http.NewRequest(
                     "PATCH",
                     "https://api.github.com/gists/" + strings.SplitN(url, "/", 6)[4],
                     reader);

    check(err);

    req.Header.Set("Authorization", "token " + githubOAuth);
    req.Header.Set("Accept", "application/vnd.github.v3+json");

    resp, err := client.Do(req);
    check(err);

    defer resp.Body.Close();

    str, err := ioutil.ReadAll(resp.Body);
    fmt.Printf("%s", str);
    fmt.Println();

    check(err);
    return resp.StatusCode == 200;
}
