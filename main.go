package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuta1402/t2km-problem-generator/contest"
	"github.com/yuta1402/t2km-problem-generator/problem"
)

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

func main() {
	rand.Seed(time.Now().UnixNano())

	id := flag.String("id", "", "id of atcoder virtual contest")
	password := flag.String("password", "", "password of atcoder virtual contest")
	pointsStr := flag.String("points", "", "problem points (e.g. 100-200-300-400)")
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

	err = cg.Generate(option)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
}
