package bot;

import "net/http"
import "encoding/json"
import "time"
import "bytes"
import "strings"

func getSchedule(gistUrl string) []Schedule {
    resp, err := http.Get(gistUrl);
    check(err);

    defer resp.Body.Close();

    var gistContent GistContent;

    err = json.NewDecoder(resp.Body).Decode(&gistContent);
    check(err);
    scheduleArray := gistContent.Schedule;

    for i := 0; i < len(scheduleArray); i++ {
        scheduleArray[i].Time = time.Unix(
            0,
            scheduleArray[i].IntTime * (1000*1000));
    }
    return scheduleArray;
}

func sendSchedule(scheduleArray *[]Schedule, githubOAuth string, gistUrl string, isLive bool) bool {

    gist := Gist{};

    preContent := GistContent {
        IsLive: isLive,
        Schedule: *scheduleArray,
    }

    content, err := json.MarshalIndent(preContent, "", " ");
    gist.Files.ScheduleJson.Content = string(content);

    buffer, err := json.MarshalIndent(&gist, "", " ");
    check(err);

    reader := bytes.NewReader(buffer);


    client := &http.Client{};
    req, err := http.NewRequest(
                     "PATCH",
                     "https://api.github.com/gists/" + strings.SplitN(gistUrl, "/", 6)[4],
                     reader);

    check(err);

    req.Header.Set("Authorization", "token " + githubOAuth);
    req.Header.Set("Accept", "application/vnd.github.v3+json");

    resp, err := client.Do(req);
    check(err);

    defer resp.Body.Close();

    return resp.StatusCode == 200;
}

func addScheduleEntry(scheduleAddition Schedule, scheduleArray *[]Schedule, githubOAuth string, gistUrl string, isLive bool) bool {

    *scheduleArray = append(*scheduleArray, scheduleAddition);
    sortSchedule(*&scheduleArray);

    return sendSchedule(*&scheduleArray, githubOAuth, gistUrl, isLive);
}

func sortSchedule(scheduleArray *[]Schedule) {
    if (len(*scheduleArray) == 0) {
        return;
    }

    tempArray := []Schedule{(*scheduleArray)[0]};

    for i := 1; i < len(*scheduleArray); i++ {
        for j := 0; j <= i; j++ {
            if (i == j) {
                tempArray = append(tempArray, (*scheduleArray)[i]);
            }
            if (tempArray[j].Time.After((*scheduleArray)[i].Time)) {
                tempTempArray := tempArray[j:i-1];
                tempArray = append(tempArray[0:j], (*scheduleArray)[i]);
                tempArray = append(tempArray, tempTempArray...);
                break;
            }
        }

    }
    *scheduleArray = tempArray;
}

func cleanSchedule(scheduleArray *[]Schedule) {
    if (len(*scheduleArray) == 0) {
        return;
    }
    for i := len(*scheduleArray) - 1; i >= 0; i-- {
        if (time.Now().AddDate(0, 0, -1).After((*scheduleArray)[i].Time)) {
            if (i == len(*scheduleArray) -2) {
                scheduleArray = nil;
            } else {
                *scheduleArray = (*scheduleArray)[i+1:len(*scheduleArray)-1];
            }
            return;
        }
    }
}
