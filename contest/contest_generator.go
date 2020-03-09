package contest

import (
	"fmt"
	"net/url"
	"path"
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

func makeDateStr(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%04d/%02d/%02d", y, m, d)
}

func makeDayHourMinute(t time.Time) (string, string, string) {
	d := makeDateStr(t)
	h := strconv.Itoa(t.Hour())
	m := strconv.Itoa(t.Minute())
	return d, h, m
}

// 時刻が5分刻みになるように補正 (AtCoderVirtualContestの仕様)
func CorrectTime(t time.Time) time.Time {
	t = t.Add(time.Duration(5-(t.Minute()%5)) * time.Minute)
	return t
}

func (cg *ContestGenerator) Generate(option Option) error {
	p, err := cg.avcPage.NewPage()
	if err != nil {
		return err
	}

	u, err := url.Parse(AtCoderVirtualContestEndpoint)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "/coordinate")
	if err := p.Navigate(u.String()); err != nil {
		return err
	}

	startDay, startHour, startMinute := makeDayHourMinute(option.StartTime)
	endDay, endHour, endMinute := makeDayHourMinute(option.StartTime.Add(option.DurationMin))

	// <input>の入力項目を処理
	{
		m := []struct {
			name  string
			value string
		}{
			{"title", option.NamePrefix},
			{"description", option.Description},
			{"start_day", startDay},
			{"end_day", endDay},
			{"penalty", strconv.Itoa(option.PenaltyMin)},
		}

		for _, o := range m {
			e := p.FindByName(o.name)

			// Send ESC key for hidden calendar
			e.SendKeys("\uE00C")

			if err := e.Fill(o.value); err != nil {
				return err
			}
		}
	}

	// <select>の入力項目を処理
	{
		m := []struct {
			name  string
			value string
		}{
			{"start_hour", startHour},
			{"start_minute", startMinute},
			{"end_hour", endHour},
			{"end_minute", endMinute},
		}

		for _, o := range m {
			e := p.FindByName(o.name)

			if err := e.Select(o.value); err != nil {
				return err
			}
		}
	}

	if option.Private {
		if err := p.FindByName("private").Check(); err != nil {
			return err
		}
	}

	if err := p.Find("body > div.container > form > div:nth-child(6) > button").Submit(); err != nil {
		return err
	}

	for _, prob := range option.Problems {
		url, err := prob.URL()
		if err != nil {
			return err
		}

		urlElement := p.Find("body > div.container > div > form:nth-child(5) > div > input")
		if err := urlElement.Fill(url); err != nil {
			return err
		}

		submitElement := p.Find("body > div.container > div > form:nth-child(5) > button")
		if err := submitElement.Submit(); err != nil {
			return err
		}
	}

	contestURL, err := p.URL()
	if err != nil {
		return err
	}

	contestURL = strings.ReplaceAll(contestURL, "setting", "contest")
	fmt.Println(contestURL)

	return nil
}

func (cg *ContestGenerator) GetLastContestIndex(contestNamePrefix string) (int, error) {
	return cg.avcPage.GetLastContestIndex(contestNamePrefix)
}
