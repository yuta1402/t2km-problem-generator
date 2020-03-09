package problem

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"path"
)

const (
	AtCoderContestEndpoint  = "https://atcoder.jp/contests/"
	AtCoderProblemsEndpoint = "https://kenkoooo.com/atcoder/resources/"
)

type Problem struct {
	ID        string  `json:"id"`
	ContestID string  `json:"contest_id"`
	Title     string  `json:"title"`
	Point     float64 `json:"point"`
}

func (p Problem) URL() (string, error) {
	u, err := url.Parse(AtCoderContestEndpoint)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, p.ContestID, "/tasks", p.ID)
	return u.String(), nil
}

type Problems []Problem

func New() (Problems, error) {
	u, err := url.Parse(AtCoderProblemsEndpoint)
	if err != nil {
		return nil, err
	}

	// u.Path = path.Join(u.Path, "/problems.json")
	u.Path = path.Join(u.Path, "/merged-problems.json")

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var problems Problems

	err = json.NewDecoder(resp.Body).Decode(&problems)
	if err != nil {
		return nil, err
	}

	return problems, nil
}

func (p Problems) RandomSelectByPoints(points []float64) Problems {
	selected := Problems{}

	for _, point := range points {
		tmp := Problems{}

		for _, q := range p {
			if q.Point == point {
				tmp = append(tmp, q)
			}
		}

		// 対象の点数が存在しない場合はスキップ
		if len(tmp) == 0 {
			continue
		}

		i := rand.Intn(len(tmp))
		selected = append(selected, tmp[i])
	}

	return selected
}
