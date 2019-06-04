package cdp

import (
	"github.com/chromedp/chromedp"
)

func screenshot(result *[]byte, url string) {
	tasks := chromedp.Tasks{
		chromedp.CaptureScreenshot(result),
	}
	ExecuteOnPage(url, tasks)
}

func pdf(result *[]byte, url string) {
	// @TODO
}

func content(result *[]byte, url string) {
	// @TODO
}
