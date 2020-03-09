package contest

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sclevine/agouti"
	"github.com/yuta1402/t2km-problem-generator/problem"
)

type ContestGenerator struct {
	ID       string
	Password string

	driver *agouti.WebDriver
	page   *agouti.Page
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

const (
	AtCoderVirtualContestEndpoint = "https://not-522.appspot.com/"
	UserAgent                     = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.120 Safari/537.36"
)

func CapabilitiesOption() agouti.Option {
	capabilities := agouti.NewCapabilities()
	capabilities["phantomjs.page.settings.userAgent"] = UserAgent
	capabilitiesOption := agouti.Desired(capabilities)
	return capabilitiesOption
}

func NewContestGenerator(id string, password string) (*ContestGenerator, error) {
	capabilitiesOption := CapabilitiesOption()

	driver := agouti.PhantomJS(capabilitiesOption)
	if err := driver.Start(); err != nil {
		return nil, err
	}

	page, err := driver.NewPage()
	if err != nil {
		return nil, err
	}

	cg := &ContestGenerator{
		ID:       id,
		Password: password,
		driver:   driver,
		page:     page,
	}

	return cg, nil
}

func (cg *ContestGenerator) Close() {
	cg.driver.Stop()
}

func (cg *ContestGenerator) Login() error {
	p := cg.page

	u, err := url.Parse(AtCoderVirtualContestEndpoint)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "/login")
	if err := p.Navigate(u.String()); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	if err := p.FindByName("id").Fill(cg.ID); err != nil {
		return err
	}

	if err := p.FindByName("password").Fill(cg.Password); err != nil {
		return err
	}

	if err := p.Find("form > button").Submit(); err != nil {
		return err
	}

	url, err := p.URL()
	if err != nil {
		return err
	}

	// ログインページに戻されてしまった場合はログイン失敗
	if url == u.String() {
		return errors.New("failed to login")
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
	p := cg.page

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
	p := cg.page

	u, err := url.Parse(AtCoderVirtualContestEndpoint)
	if err != nil {
		return 0, err
	}

	u.Path = path.Join(u.Path, "/participated")
	if err := p.Navigate(u.String()); err != nil {
		return 0, err
	}

	contests := p.Find("body > div > table > tbody").All("tr > td:nth-child(1) > a")
	count, err := contests.Count()
	if err != nil {
		return 0, err
	}

	maxIndex := 0

	for i := 0; i < count; i++ {
		name, err := contests.At(i).Text()
		if err != nil {
			return 0, err
		}

		if strings.Contains(name, contestNamePrefix) {
			indexStr := strings.ReplaceAll(name, contestNamePrefix, "")
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
