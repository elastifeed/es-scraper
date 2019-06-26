package cdptab

import (
	"context"
	"errors"
	"log"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/storage"
)

// TabState is a type for state of a tab. Whether it is accepting new request or still busy with another request
type tabState int

// Possible states for a tab.
const (
	Accepting tabState = iota
	Busy
)

// ChromeTab describes the relevant data of a tab in a browser.
type ChromeTab struct {
	ID      uint                // Id of the browser tab
	Context *context.Context    // Pointer to the context that describes this tab
	Stop    *context.CancelFunc // CancelFunc to cancel it's context
	URL     string              // The url that is loaded in this tab
	State   tabState            // The working state of this tab. Whether it is ready to accept a new request or still busy with another request
	Store   storage.Storager    // S3 Storage where PDFs etc get stored
}

// ChromeTabReturns is a datastructure for the results of any operation on a tab.
type ChromeTabReturns struct {
	Data map[string]interface{} // Map containing unmarshaled json data.
	Err  error
}

// NewBrowserTab creates a new tab and returns it.
func NewBrowserTab(id uint, store storage.Storager, parentContext *context.Context) ChromeTab {

	// Create the new context
	ctx, cancel := chromedp.NewContext(*parentContext, chromedp.WithLogf(log.Printf))

	// Ensure the tab is actually started.
	if err := chromedp.Run(ctx); err != nil {
		panic(err)
	}

	// Set metrics
	metrics := emulation.SetDeviceMetricsOverride(1024, 1024, 1.0, true)
	if err := chromedp.Run(ctx, metrics); err != nil {
		panic(err)
	}

	return ChromeTab{
		ID:      id,
		Store:   store,
		Context: &ctx,
		Stop:    &cancel,
		URL:     "",
		State:   Accepting,
	}

}

// Busy sets the tab to be unavailable to recieve new requests.
func (tab *ChromeTab) Busy() {
	tab.State = Busy
}

// Ready sets the tab to be ready to work on a request
func (tab *ChromeTab) Ready() {
	tab.State = Accepting
}

// Navigate the tab to a page
func (tab *ChromeTab) Navigate(url string) error {
	tab.URL = url
	chromedp.Run(*tab.Context, chromedp.Navigate(url))

	return nil
}

// Screenshot renders the current page as a png file and saves the result.
// returns: the path where the screenshot was saved
func (tab *ChromeTab) Screenshot(ch chan ChromeTabReturns) {
	var result []byte
	if tab.URL == "" {
		ch <- ChromeTabReturns{nil, errors.New("Tried screenshotting empty page")}
		return
	}

	tasks := chromedp.Tasks{
		// Build the neccessary function(s)
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("[%d] Taking screenshot...", tab.ID)
			viewport := page.Viewport{
				X:      0,
				Y:      0,
				Width:  1024,
				Height: 1024,
				Scale:  1,
			}
			var err error
			result, err = page.CaptureScreenshot().WithClip(&viewport).Do(ctx)
			log.Printf("[%d]... screenshot done.", tab.ID)
			return err
		}),
	}
	// Run the functions
	if err := chromedp.Run(*tab.Context, tasks); err != nil {
		ch <- ChromeTabReturns{nil, err}
		return
	}

	// Save the result
	savePath, saverr := tab.Store.Upload(result, "png")
	if saverr != nil {
		ch <- ChromeTabReturns{nil, saverr}
		return
	}

	data := map[string]interface{}{
		"screenshot": savePath,
	}

	ch <- ChromeTabReturns{data, nil}
}

// Pdf thakes an url and renders it as a pdf file.
func (tab *ChromeTab) Pdf(ch chan ChromeTabReturns) {
	var result []byte
	if tab.URL == "" {
		ch <- ChromeTabReturns{nil, errors.New("Tried rendering empty page")}
		return
	}

	tasks := chromedp.Tasks{
		//chromedp.WaitReady("#document"),
		// Use a chromedp.ActionFunc to build an executable function
		chromedp.ActionFunc(func(ctx context.Context) error { // The context is set when Run calls Do for each each Action
			log.Printf("[%d] Rendering pdf...", tab.ID)

			var err error
			result, _, err = page.PrintToPDF().WithTransferMode(page.PrintToPDFTransferModeReturnAsBase64).Do(ctx)
			log.Printf("[%d]... pdf done.", tab.ID)
			return err

		}),
	}
	if err := chromedp.Run(*tab.Context, tasks); err != nil {
		ch <- ChromeTabReturns{nil, err}
		return
	}
	savePath, saverr := tab.Store.Upload(result, "pdf")
	if saverr != nil {
		ch <- ChromeTabReturns{nil, saverr}
		return
	}

	data := map[string]interface{}{
		"pdf": savePath,
	}

	ch <- ChromeTabReturns{data, nil}
}

// Content retrieves the content on a page
func (tab *ChromeTab) Content(ch chan ChromeTabReturns) {

}

// Scrape runs a full scrape on a page.
func (tab *ChromeTab) Scrape(ch chan ChromeTabReturns) {
	screen := make(chan ChromeTabReturns)
	pdf := make(chan ChromeTabReturns)
	//con := make(chan ChromeTabReturns)

	// Run each action in their own routine, collect and combine results before sending.
	log.Printf("[%d] Running full scrape...", tab.ID)
	go tab.Screenshot(screen)
	go tab.Pdf(pdf)
	// go tab.Content(con) TODO

	s := <-screen // @TODO Does this actually run parallel? Or do they block each other
	p := <-pdf
	// c := <-con

	if s.Err != nil {
		log.Print(s.Err)
		ch <- ChromeTabReturns{nil, s.Err}
		return
	}
	if p.Err != nil {
		log.Print(p.Err)
		ch <- ChromeTabReturns{nil, p.Err}
		return
	}
	/*if c.Err != nil {
		ch <- ChromeTabReturns{nil, c.Err}
		return
	}*/

	//c.Data["screenshot"] = s.Data["screenshot"]
	//c.Data["pdf"] = p.Data["pdf"]
	p.Data["screenshot"] = s.Data["screenshot"]
	log.Printf("[%d] ... full scrape done.", tab.ID)

	ch <- p
}
