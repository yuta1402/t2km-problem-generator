package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuta1402/t2km-problem-generator/problems"
)

func parsePoints(pointsStr string) ([]float64, error) {
	slice := strings.Split(pointsStr, "-")

	points := []float64{}

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

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	switch os.Args[1] {
	case "list":
		pointsStr := listCmd.String("points", "", "problem points (e.g. 100-200-300-400)")
		listCmd.Parse(os.Args[2:])

		fmt.Println(*pointsStr)
		points, err := parsePoints(*pointsStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse points error: %s\n", err)
			return
		}

		problems, err := problems.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		probs := problems.RandomSelectByPoints(points)
		fmt.Println(probs)
	}
}
