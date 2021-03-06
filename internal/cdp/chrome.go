package cdp

import (
	"context"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/cdptab"
	"github.com/elastifeed/es-scraper/internal/storage"
)

const numTabs = 4

// BrowserTabs contains a list of tabs in this browser.
var BrowserTabs = struct {
	tabs []cdptab.ChromeTab
	mux  sync.Mutex
}{tabs: []cdptab.ChromeTab{}}

// Launch starts a new headless browser and returns the function to cancel that browser.
func Launch(mercuryURL string, store storage.Storager) (context.Context, context.CancelFunc) {
	allocctx, ccl := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:3000")

	// Set up the multi tab request execution enviroment

	// Make the task queue.
	queue = make(chan task)
	for i := 0; i < numTabs; i++ {
		// Create and keep track of tabs
		BrowserTabs.tabs = append(BrowserTabs.tabs, cdptab.NewBrowserTab(uint(i), mercuryURL, store, &allocctx))
		// Start the task queue workers
		go queueWorker(queue)
	}

	return allocctx, ccl
}
