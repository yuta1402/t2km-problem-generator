package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/yuta1402/t2km-problem-generator/problems"
)

func main() {
	problems, err := problems.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	fmt.Println(len(problems))

	rand.Seed(time.Now().UnixNano())

	probs := problems.RandomSelectByPoints([]float64{100, 200, 300, 400, 500})
	fmt.Println(probs)
}
