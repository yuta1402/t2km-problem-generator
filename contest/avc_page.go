package contest

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
	"github.com/yuta1402/t2km-problem-generator/problem"
)

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

type AVCPage struct {
	driver  *agouti.WebDriver
	cookies []*http.Cookie
}

type ContestOption struct {
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	PenaltyMin  int
	Private     bool
	Problems    problem.Problems
}

type CoordinatedContest struct {
	Option ContestOption
	URL    string
}

type ParticipatedContest struct {
	Name         string
	StartTimeStr string
	EndTimeStr   string
}

func NewAVCPage() (*AVCPage, error) {
	capabilitiesOption := CapabilitiesOption()

	driver := agouti.PhantomJS(capabilitiesOption)
	if err := driver.Start(); err != nil {
		return nil, err
	}

	avcPage := &AVCPage{
		driver:  driver,
		cookies: nil,
	}

	return avcPage, nil
}

func (avcPage *AVCPage) Close() {
	avcPage.driver.Stop()
}

func (avcPage *AVCPage) Login(id string, password string) error {
	p, err := avcPage.driver.NewPage()
	if err != nil {
		return err
	}

	u, err := url.Parse(AtCoderVirtualContestEndpoint)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "/login")
	if err := p.Navigate(u.String()); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	if err := p.FindByName("id").Fill(id); err != nil {
		return err
	}

	if err := p.FindByName("password").Fill(password); err != nil {
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

	cookies, err := p.GetCookies()
	if err != nil {
		return err
	}

	avcPage.cookies = cookies

	return nil
}

func (avcPage *AVCPage) NewPage() (*agouti.Page, error) {
	p, err := avcPage.driver.NewPage()
	if err != nil {
		return nil, err
	}

	for _, c := range avcPage.cookies {
		p.SetCookie(c)
	}

	return p, nil
}

func (avcPage *AVCPage) NewPageWithPath(urlPath string) (*agouti.Page, error) {
	p, err := avcPage.NewPage()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(AtCoderVirtualContestEndpoint)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, urlPath)
	if err := p.Navigate(u.String()); err != nil {
		return nil, err
	}

	return p, nil
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

func (avcPage *AVCPage) CoordinateContest(option ContestOption) (*CoordinatedContest, error) {
	p, err := avcPage.NewPageWithPath("/coordinate")
	if err != nil {
		return nil, err
	}

	startDay, startHour, startMinute := makeDayHourMinute(option.StartTime)
	endDay, endHour, endMinute := makeDayHourMinute(option.EndTime)

	// <input>の入力項目を処理
	{
		m := []struct {
			name  string
			value string
		}{
			{"title", option.Name},
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
				return nil, err
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
				return nil, err
			}
		}
	}

	if option.Private {
		if err := p.FindByName("private").Check(); err != nil {
			return nil, err
		}
	}

	if err := p.Find("body > div.container > form > div:nth-child(6) > button").Submit(); err != nil {
		return nil, err
	}

	for _, prob := range option.Problems {
		url, err := prob.URL()
		if err != nil {
			return nil, err
		}

		urlElement := p.Find("body > div.container > div > form:nth-child(5) > div > input")
		if err := urlElement.Fill(url); err != nil {
			return nil, err
		}

		submitElement := p.Find("body > div.container > div > form:nth-child(5) > button")
		if err := submitElement.Submit(); err != nil {
			return nil, err
		}
	}

	contestURL, err := p.URL()
	if err != nil {
		return nil, err
	}

	contestURL = strings.ReplaceAll(contestURL, "setting", "contest")

	cc := &CoordinatedContest{
		Option: option,
		URL:    contestURL,
	}

	return cc, nil
}

func (avcPage *AVCPage) GetParticipatedContests() ([]ParticipatedContest, error) {

	p, err := avcPage.NewPageWithPath("/participated")
	if err != nil {
		return nil, err
	}

	html, err := p.HTML()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))

	contests := []ParticipatedContest{}

	doc.Find("body > div > table > tbody > tr").Each(func(index int, s *goquery.Selection) {
		if s.Children().Size() != 3 {
			return
		}

		name := strings.TrimSpace(s.Children().Eq(0).Text())
		startTimeStr := strings.TrimSpace(s.Children().Eq(1).Text())
		endTimeStr := strings.TrimSpace(s.Children().Eq(2).Text())

		contests = append(contests, ParticipatedContest{
			Name:         name,
			StartTimeStr: startTimeStr,
			EndTimeStr:   endTimeStr,
		})
	})

	return contests, nil
}
