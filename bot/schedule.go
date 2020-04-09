package bot;

import "net/http"
import "encoding/json"
import "time"
import "bytes"
import "strings"
import "sort"

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

func sendSchedule(scheduleArray *ScheduleArray, githubOAuth string, gistUrl string) bool {

    gist := Gist{};
    content, err := json.MarshalIndent(*scheduleArray, "", " ");
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

func addScheduleEntry(scheduleAddition Schedule, scheduleArray *ScheduleArray, githubOAuth string, gistUrl string) bool {

    scheduleArray.Schedule = append(scheduleArray.Schedule, scheduleAddition);
    sortSchedule(*&scheduleArray);

    return sendSchedule(*&scheduleArray, githubOAuth, gistUrl);
}

func sortSchedule(scheduleArray *ScheduleArray) {
    sort.Sort(ScheduleSlice(scheduleArray.Schedule));
}

func cleanSchedule(scheduleArray *ScheduleArray) {
    if (len(scheduleArray.Schedule) == 0) {
        return;
    }
    for i := len(scheduleArray.Schedule) - 1; i >= 0; i-- {
        if (time.Now().AddDate(0, 0, -1).After(scheduleArray.Schedule[i].Time)) {
            if (i == len(scheduleArray.Schedule) -2) {
                scheduleArray.Schedule = nil;
            } else {
                scheduleArray.Schedule = scheduleArray.Schedule[i+1:len(scheduleArray.Schedule)-1];
            }
            return;
        }
    }
}
