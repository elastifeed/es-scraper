package cdptab

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/storage"
	"k8s.io/klog"
)

// NewBrowserTab creates a new tab and returns it.
func NewBrowserTab(id uint, mercuryURL string, store storage.Storager, parentContext *context.Context) ChromeTab {

	// Create the new context
	ctx, cancel := chromedp.NewContext(*parentContext, chromedp.WithLogf(klog.Infof))

	// Ensure the tab is actually started.
	if err := chromedp.Run(ctx); err != nil {
		panic(err)
	}

	// Set metrics
	metrics := emulation.SetDeviceMetricsOverride(1024, 1024, 1.0, true)
	if err := chromedp.Run(ctx, metrics); err != nil {
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
	tab.URL = url
	chromedp.Run(*tab.Context, chromedp.Navigate(url))
	klog.Infof("[%d] Navigated to url:\t%s", tab.ID, tab.URL)

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
			klog.Infof("[%d] Taking screenshot...", tab.ID)
			viewport := page.Viewport{
				X:      0,
				Y:      0,
				Width:  1024,
				Height: 1024,
				Scale:  1,
			}
			var err error
			result, err = page.CaptureScreenshot().WithClip(&viewport).Do(ctx)
			klog.Infof("[%d] Taking screenshot done.", tab.ID)
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
		//chromedp.WaitReady("#document"),
		// Use a chromedp.ActionFunc to build an executable function
		chromedp.ActionFunc(func(ctx context.Context) error { // The context is set when Run calls Do for each each Action
			klog.Infof("[%d] Rendering pdf...", tab.ID)

			var err error
			result, _, err = page.PrintToPDF().WithTransferMode(page.PrintToPDFTransferModeReturnAsBase64).Do(ctx)
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
	var html string
	if tab.URL == "" {
		ch <- ChromeTabReturns{nil, errors.New("Tried retrieving content from empty page")}
		return
	}

	// Retrieve the entire page's html
	klog.Infof("[%d] Extracting html...", tab.ID)
	chromedp.Run(*tab.Context, chromedp.OuterHTML("html", &html))
	klog.Infof("[%d] Extracting html done.", tab.ID)

	// Encode the data for json
	r, w := io.Pipe()
	defer r.Close()
	//data := []byte(fmt.Sprintf("\"{\"url\":\"%s\", \"html\":\"%s\"}\"", tab.URL, html))
	go func() {
		payload := MercuryPayload{URL: tab.URL, HTML: html}
		json.NewEncoder(w).Encode(&payload)
		defer w.Close()
	}()

	// Post to the mercury api for the content
	client := &http.Client{}
	req, _ := http.NewRequest("POST", tab.MercuryURL, r)
	req.Header.Set("Content-Type", "application/json")

	klog.Infof("[%d] Retrieving content...", tab.ID)
	resp, err := client.Do(req)
	if err != nil {
		ch <- ChromeTabReturns{nil, err}
		return
	}
	defer resp.Body.Close()     // We need to close the body when we are done
	if resp.StatusCode != 200 { // Rely on the return value of mercury to determine if a parse failed
		ch <- ChromeTabReturns{result, errors.New("Error occured while parsing with mercury")}
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
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

// Scrape runs a full scrape on a page.
func (tab *ChromeTab) Scrape(ch chan ChromeTabReturns) {
	screen := make(chan ChromeTabReturns)
	pdf := make(chan ChromeTabReturns)
	con := make(chan ChromeTabReturns)

	// Run each action in their own routine, collect and combine results before sending.
	klog.Infof("[%d] Full scrape...", tab.ID)
	go tab.Screenshot(screen)
	go tab.Pdf(pdf)
	go tab.Content(con)

	s := <-screen // @TODO Does this actually run parallel? Or do they block each other
	p := <-pdf
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
