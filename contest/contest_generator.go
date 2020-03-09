package contest

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yuta1402/t2km-problem-generator/problem"
)

type ContestGenerator struct {
	ID       string
	Password string

	avcPage *AVCPage
}

type Option struct {
	NamePrefix  string
	Description string
	StartTime   time.Time
	DurationMin time.Duration
	PenaltyMin  int
	Private     bool
	Problems    problem.Problems
}

func NewContestGenerator(id string, password string) (*ContestGenerator, error) {
	avcPage, err := NewAVCPage()
	if err != nil {
		return nil, err
	}

	cg := &ContestGenerator{
		ID:       id,
		Password: password,
		avcPage:  avcPage,
	}

	return cg, nil
}

func (cg *ContestGenerator) Close() {
	cg.avcPage.Close()
}

func (cg *ContestGenerator) Login() error {
	if err := cg.avcPage.Login(cg.ID, cg.Password); err != nil {
		return err
	}

	return nil
}

// 時刻が5分刻みになるように補正 (AtCoderVirtualContestの仕様)
func CorrectTime(t time.Time) time.Time {
	t = t.Add(time.Duration(5-(t.Minute()%5)) * time.Minute)
	return t
}

func (cg *ContestGenerator) Generate(option Option) error {
	contestOption := ContestOption{
		Name:        option.NamePrefix,
		Description: option.Description,
		StartTime:   option.StartTime,
		EndTime:     option.StartTime.Add(option.DurationMin),
		PenaltyMin:  5,
		Private:     true,
		Problems:    option.Problems,
	}

	cc, err := cg.avcPage.CoordinateContest(contestOption)
	fmt.Println(cc.URL)

	return err
}

func (cg *ContestGenerator) GetLastContestIndex(contestNamePrefix string) (int, error) {
	contests, err := cg.avcPage.GetParticipatedContests()
	if err != nil {
		return 0, err
	}

	maxIndex := 0

	for _, c := range contests {
		if strings.Contains(c.Name, contestNamePrefix) {
			indexStr := strings.ReplaceAll(c.Name, contestNamePrefix, "")
			indexStr = strings.TrimSpace(indexStr)
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				continue
			}

			if index > maxIndex {
				maxIndex = index
			}
		}
	}

	return maxIndex, nil
}
