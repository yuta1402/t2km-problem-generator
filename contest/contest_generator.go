package contest

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/sclevine/agouti"
)

type ContestGenerator struct {
	ID       string
	Password string

	driver *agouti.WebDriver
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

	cg := &ContestGenerator{
		ID:       id,
		Password: password,
		driver:   driver,
	}

	return cg, nil
}

func (cg *ContestGenerator) Close() {
	cg.driver.Stop()
}

func (cg *ContestGenerator) Login() error {

	p, err := cg.driver.NewPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
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

	if err := p.FindByName("id").Fill(cg.ID); err != nil {
		fmt.Println("test")
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
