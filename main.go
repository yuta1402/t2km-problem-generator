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

func main() {
	rand.Seed(time.Now().UnixNano())

	id := flag.String("id", "", "id of atcoder virtual contest")
	password := flag.String("password", "", "password of atcoder virtual contest")
	pointsStr := flag.String("points", "", "problem points (e.g. 100-200-300-400)")
	apiURL := flag.String("api", "", "API of slack")
	flag.Parse()

	points, err := parsePoints(*pointsStr)
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

	cg, err := contest.NewContestGenerator(*id, *password)
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

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	now := time.Now().In(jst)
	startTime := contest.CorrectTime(now)
	durationMin := time.Duration(100) * time.Minute

	option := contest.Option{
		NamePrefix:  "tmp contest",
		Description: "",
		StartTime:   startTime,
		DurationMin: durationMin,
		PenaltyMin:  5,
		Private:     true,
		Problems:    probs,
	}

	cc, err := cg.Generate(option)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fmt.Println(cc.URL)

	res, err := postSlack(cc, *apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fmt.Println(res.Status)
}
