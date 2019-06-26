package cdp

import (
	"context"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/cdptab"
	"github.com/elastifeed/es-scraper/internal/storage"
)

const userAgent = "Googlebot/2.1 (+http://www.google.com/bot.html)"
const numTabs = 2

// BrowserTabs contains a list of tabs in this browser.
var BrowserTabs = struct {
	tabs []cdptab.ChromeTab
	mux  sync.Mutex
}{tabs: []cdptab.ChromeTab{}}

// Launch starts a new headless browser and returns the function to cancel that browser.
func Launch(store storage.Storager) (context.Context, context.CancelFunc) {
	launchOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserAgent(userAgent),
		// chromedp.Headless, // for debugging uncomment to see a browser window
	}
	// Allocate the basis for a browser.
	allocctx, ccl := chromedp.NewExecAllocator(context.Background(), launchOpts...)

	// Set up the multi tab request execution enviroment

	// Make the task queue.
	queue = make(chan task)
	for i := 0; i < numTabs; i++ {
		// Create and keep track of tabs
		BrowserTabs.tabs = append(BrowserTabs.tabs, cdptab.NewBrowserTab(uint(i), store, &allocctx))
		// Start the task queue workers
		go queueWorker(queue)
	}

	// emulation.SetEmulatedMedia("screen").Do(mainContext) // @TODO Move to right position
	return allocctx, ccl
}
