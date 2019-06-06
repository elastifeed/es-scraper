package cdp

import (
	"io/ioutil"

	"github.com/chromedp/chromedp"
	"github.com/golang/glog"
)

func Screenshot(url string) {
	var result []byte
	tasks := chromedp.Tasks{
		chromedp.CaptureScreenshot(&result),
	}
	ExecuteOnPage(url, tasks)
	if err := saveFile(&result); err != nil {
		glog.Fatal(err)
	}
}

func pdf(result *[]byte, url string) {
	// @TODO
}

func content(result *[]byte, url string) {
	// @TODO
}

// Saves the given data to the disk.
func saveFile(data *[]byte) error {
	return ioutil.WriteFile(makeFilename(), *data, 0644)
}

func makeFilename() string { return "/tmp/chromedp_test.png" }
