package cdp

import (
	"context"

	"github.com/golang/glog"

	"github.com/chromedp/chromedp"
)

var mainContext context.Context

// Launch starts a new headless browser and returns the function to cancel that browser.
func Launch() (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	mainContext, cancel = chromedp.NewContext(context.Background())
	return mainContext, cancel
}

// ExecuteOnPage takes a url and a list of actions which should be performed on that url.
func ExecuteOnPage(url string, actions ...chromedp.Action) {
	// First navigate and wait until the resource is ready before executing the other functions.
	navAndWait := []chromedp.Action{chromedp.Navigate(url), chromedp.WaitReady("body")}
	err := chromedp.Run(mainContext, append(navAndWait, actions...)...)
	if err != nil {
		glog.Error(err)
	}

}
