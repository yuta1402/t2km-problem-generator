package contest

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
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

func (avcPage *AVCPage) GetLastContestIndex(contestNamePrefix string) (int, error) {
	p, err := avcPage.NewPageWithPath("/participated")
	if err != nil {
		return 0, err
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
