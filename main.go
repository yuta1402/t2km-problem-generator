package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuta1402/t2km-problem-generator/contest"
	"github.com/yuta1402/t2km-problem-generator/problem"
)

type RequestData struct {
	Text string `json:"text"`
}

func parsePoints(pointsStr string) ([]float64, error) {
	points := []float64{}

	if pointsStr == "" {
		return points, errors.New("points is empty")
	}

	slice := strings.Split(pointsStr, "-")

	for _, s := range slice {
		p, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}

		points = append(points, p)
	}

	return points, nil
}

func postSlack(cc *contest.CoordinatedContest, apiURL string) (*http.Response, error) {
	text := "*「" + cc.Option.Name + "」開催のお知らせ*\n" +
		"日時: " + cc.Option.StartTime.Format("2006/01/02 15:04") + "-\n" +
		"会場: " + cc.URL + "\n" +
		"参加できる方は:ok: 絵文字、参加できない方は:ng: 絵文字でお知らせください。"

	d := RequestData{
		Text: text,
	}

	json, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(apiURL, "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func makeStartTime(now time.Time, startWeekday int, startTimeStr string) (time.Time, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return time.Time{}, err
	}

	now = time.Now().In(jst)

	diffDay := (startWeekday - int(now.Weekday()) + 7) % 7
	startDay := now.Day() + diffDay

	a := strings.Split(startTimeStr, ":")
	if len(a) != 2 {
		return time.Time{}, errors.New("start time parse error")
	}

	startHour, err := strconv.Atoi(a[0])
	if err != nil {
		return time.Time{}, err
	}

	startMin, err := strconv.Atoi(a[1])
	if err != nil {
		return time.Time{}, err
	}

	startTime := time.Date(now.Year(), now.Month(), startDay, startHour, startMin, 0, 0, jst)
	return startTime, nil
}

func main() {
	now := time.Now()
	rand.Seed(now.UnixNano())

	var (
		id           string
		password     string
		pointsStr    string
		startWeekday int
		startTimeStr string
		durationMin  int
		penaltyMin   int
		apiURL       string
	)

	flag.StringVar(&id, "id", "", "id of atcoder virtual contest")
	flag.StringVar(&password, "password", "", "password of atcoder virtual contest")
	flag.StringVar(&pointsStr, "points", "", "problem points (e.g. 100-200-300-400)")
	flag.IntVar(&startWeekday, "start-weekday", int(now.Weekday()), "start weekday Sun=0, Mon=1, ... (default now.Weekday())")
	flag.StringVar(&startTimeStr, "start-time", "", "start time (e.g. 18:00)")
	flag.IntVar(&durationMin, "duration", 100, "duration [min] (default: 100)")
	flag.IntVar(&penaltyMin, "penalty", 5, "penalty time [min] (default: 5)")
	flag.StringVar(&apiURL, "api", "", "API of slack")
	flag.Parse()

	startTime, err := makeStartTime(now, startWeekday, startTimeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	startTime = contest.CorrectTime(startTime)

	points, err := parsePoints(pointsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse points error: %s\n", err)
		return
	}

	problems, err := problem.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	probs := problems.RandomSelectByPoints(points)
	fmt.Println(probs)

	cg, err := contest.NewContestGenerator(id, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize contest generator error: %s\n", err)
		return
	}
	defer cg.Close()

	err = cg.Login()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to login")
		return
	}

	option := contest.Option{
		NamePrefix:  "tmp contest",
		Description: "",
		StartTime:   startTime,
		DurationMin: time.Duration(durationMin) * time.Minute,
		PenaltyMin:  penaltyMin,
		Private:     true,
		Problems:    probs,
	}

	cc, err := cg.Generate(option)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fmt.Println(cc.URL)

	res, err := postSlack(cc, apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fmt.Println(res.Status)
}
