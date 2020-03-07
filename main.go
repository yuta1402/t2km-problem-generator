package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sclevine/agouti"
)

func main() {
	capabilities := agouti.NewCapabilities()
	capabilities["phantomjs.page.settings.userAgent"] = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.120 Safari/537.36"
	options := agouti.Desired(capabilities)

	driver := agouti.PhantomJS(options)
	defer driver.Stop()
	driver.Start()

	page, err := driver.NewPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	page.Navigate("https://example.com/")

	time.Sleep(1 * time.Second)

	src, err := page.HTML()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	fmt.Println(src)
}
