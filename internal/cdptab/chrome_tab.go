package cdptab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/storage"
	"k8s.io/klog"
)

var navLock sync.Mutex

// NewBrowserTab creates a new tab and returns it.
func NewBrowserTab(id uint, mercuryURL string, store storage.Storager, parentContext *context.Context) ChromeTab {

	// Create the new context
	ctx, cancel := chromedp.NewContext(*parentContext)

	// Ensure the tab is actually started.
	if err := chromedp.Run(ctx); err != nil {
		panic(err)
	}

	klog.Infof("[%d] Tab %d started with mercury at %s", id, id, mercuryURL)

	return ChromeTab{
		ID:         id,
		Store:      store,
		Context:    &ctx,
		Stop:       &cancel,
		URL:        "",
		MercuryURL: mercuryURL,
		State:      Accepting,
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
	// Navigate and especially waiting for the page to be ready is a cause for a race condition.
	// By locking the navigate funciton we ensure that this can't happen.
	navLock.Lock()
	defer navLock.Unlock()
	tab.URL = url
	klog.Infof("[%d] Reached target navigate.", tab.ID)
	defer klog.Infof("[%d] Navigated to url:\t%s", tab.ID, tab.URL)

	task := chromedp.ActionFunc(func(ctx context.Context) error {
		emulation.SetUserAgentOverride(userAgent).Do(ctx) // Set user agent
		_, _, _, err := page.Navigate(url).Do(ctx)        // Navigate
		if err != nil {
			klog.Errorf("[%d] Navigation error. %s", tab.ID, err)
			return err
		}
		wait := waitLoadedOrTimeout(ctx)                          // Wait
		_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx) // Get the size of the website
		if err != nil {
			return err
		}
		tab.ContentSize = contentSize
		return wait
	})
	return chromedp.Run(*tab.Context, task)
}

// waitLoadedOrTimeout blocks until a target receives a Page.loadEventFired or until a timeout is reached, at which point the page will be considered navigated.
func waitLoadedOrTimeout(ctx context.Context) error {
	// Modified version of waitLoaded in https://github.com/chromedp/chromedp/blob/5094b8b381a3f713ff230c638a601ff7c97479a3/nav.go
	ch := make(chan struct{})
	lctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(lctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventLoadEventFired); ok {
			cancel()
			close(ch)
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(60 * time.Second): // Navigation timeout
		klog.Warningf("[!] Navigation timed out. Continuing without Page.loadEventFired.")
		return nil
	}
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
		tab.setDeviceMetricsAction(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			klog.Infof("[%d] Taking screenshot...", tab.ID)
			viewport := page.Viewport{
				X:      tab.ContentSize.X,
				Y:      tab.ContentSize.Y,
				Width:  tab.ContentSize.Width,
				Height: tab.ContentSize.Height,
				Scale:  1,
			}
			var caperr error
			result, caperr = page.CaptureScreenshot().WithClip(&viewport).Do(ctx)
			klog.Infof("[%d] Taking screenshot done.", tab.ID)
			return caperr
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
	klog.Infof("[%d] Saved screenshot to %s", tab.ID, savePath)

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
		// Use a chromedp.ActionFunc to build an executable function
		tab.setDeviceMetricsAction(),
		chromedp.ActionFunc(func(ctx context.Context) error { // The context is set when Run calls Do for each each Action
			klog.Infof("[%d] Rendering pdf...", tab.ID)

			var err error
			//emulation.SetEmulatedMedia("print").Do(ctx)
			result, _, err = page.PrintToPDF().WithPrintBackground(true).
				WithTransferMode(page.PrintToPDFTransferModeReturnAsBase64).Do(ctx)
			klog.Infof("[%d] Rendering pdf done.", tab.ID)
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
	klog.Infof("[%d] Saved pdf to %s", tab.ID, savePath)

	data := map[string]interface{}{
		"pdf": savePath,
	}

	ch <- ChromeTabReturns{data, nil}
}

// Content retrieves the content on a page
func (tab *ChromeTab) Content(ch chan ChromeTabReturns) {
	var result map[string]interface{}
	if tab.URL == "" {
		ch <- ChromeTabReturns{nil, errors.New("Tried retrieving content from empty page")}
		return
	}

	// Post to the mercury api for the content

	klog.Infof("[%d] Retrieving content...", tab.ID)

	body, posterr := postAsJSON(tab.MercuryURL, MercuryPayload{URL: tab.URL})
	if posterr != nil {
		ch <- ChromeTabReturns{nil, posterr}
		return
	}
	json.Unmarshal(body, &result) // Unmarshal the body to work with it
	klog.Infof("[%d] Retrieving content done.", tab.ID)

	// Download the thumbnail image
	thumbnail, saverr := tab.thumbnail(&result)
	if saverr != nil {
		ch <- ChromeTabReturns{nil, saverr}
		return
	}

	result["thumbnail"] = thumbnail
	delete(result, "lead_image_url")

	ch <- ChromeTabReturns{result, nil}
}

// thumbnail downloads and saves a thumbnail to s3.
// returns: path to the downloaded thumnail
func (tab *ChromeTab) thumbnail(res *map[string]interface{}) (string, error) {
	address, ok := (*res)["lead_image_url"].(string)
	if ok == false {
		klog.Warningf("[%d] Had no thumbnail.", tab.ID)
		return "", nil
	}

	resp, err := http.Get(address)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	thumbnail, _ := ioutil.ReadAll(resp.Body)
	// Save the thumbnail
	savePath, saverr := tab.Store.Upload(thumbnail, "png")
	if saverr != nil {
		return "", saverr
	}
	klog.Infof("[%d] Saved thumbnail to %s.", tab.ID, savePath)

	return savePath, nil
}

func postAsJSON(url string, payload interface{}) ([]byte, error) {
	// Encode the data for json through a pipe
	var encodeErr error
	r, w := io.Pipe()
	defer r.Close()
	go func() {
		encodeErr = json.NewEncoder(w).Encode(&payload)
		defer w.Close()
	}()
	if encodeErr != nil {
		return nil, encodeErr
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, r)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 { // Rely on the return value of mercury to determine if a request failed annyway else
		return nil, fmt.Errorf("POST on %s with %s returned %s", url, payload, resp.Status)
	}
	defer resp.Body.Close()              // We need to close the body when we are done
	body, _ := ioutil.ReadAll(resp.Body) // Read the raw data
	return body, nil
}

func (tab *ChromeTab) setDeviceMetricsAction() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		width, height := int64(math.Ceil(tab.ContentSize.Width)), int64(math.Ceil(tab.ContentSize.Height))
		err := emulation.SetDeviceMetricsOverride(width, height, 1, false).
			WithScreenOrientation(&emulation.ScreenOrientation{
				Type:  emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}).Do(ctx)
		return err
	})
}

// Scrape runs a full scrape on a page.
func (tab *ChromeTab) Scrape(ch chan ChromeTabReturns) {
	screen := make(chan ChromeTabReturns)
	pdf := make(chan ChromeTabReturns)
	con := make(chan ChromeTabReturns)

	// Run each action in their own routine, collect and combine results before sending.
	klog.Infof("[%d] Full scrape...", tab.ID)
	go tab.Screenshot(screen)
	s := <-screen

	go tab.Pdf(pdf)
	p := <-pdf

	go tab.Content(con)
	c := <-con

	if s.Err != nil {
		ch <- ChromeTabReturns{nil, s.Err}
		return
	}
	if p.Err != nil {
		ch <- ChromeTabReturns{nil, p.Err}
		return
	}
	if c.Err != nil {
		ch <- ChromeTabReturns{nil, c.Err}
		return
	}

	c.Data["screenshot"] = s.Data["screenshot"]
	c.Data["pdf"] = p.Data["pdf"]
	klog.Infof("[%d] Full scrape done.", tab.ID)

	ch <- c
}
