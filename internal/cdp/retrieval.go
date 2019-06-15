package cdp

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/elastifeed/es-scraper/internal/storage"
)

// Screenshot takes an url and renders that url for use as a thumbnail.
func Screenshot(url string, s storage.Storager) (string, error) {
	var result []byte

	tasks := chromedp.Tasks{
		screenshotAction(&result),
	}
	if err := ExecuteOnPage(url, tasks); err != nil {
		return "", err
	}
	saveURL, err := s.Upload(result, ".png")
	if err != nil {
		return "", err
	}

	return saveURL, nil

}

// Pdf thakes an url and renders it as a pdf file.
func Pdf(url string, s storage.Storager) (string, error) {
	var result []byte

	tasks := chromedp.Tasks{
		//chromedp.WaitReady("#document"),
		pdfAction(&result),
	}
	if err := ExecuteOnPage(url, tasks); err != nil {
		return "", err
	}
	savePath, saverr := s.Upload(result, ".pdf")
	if saverr != nil {
		return "", saverr
	}

	return savePath, nil
}

// Scrape performs a full content retrieval and render on the page and returns !!@TODO!!
func Scrape(url string) {
	// var screenshotBuf, pdfBuf, thumbnailBuf []byte
}

func screenshotAction(resultBuf *[]byte) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		log.Print("Taking screenshot...")
		viewport := page.Viewport{
			X:      0,
			Y:      0,
			Width:  1024,
			Height: 1024,
			Scale:  1,
		}
		var err error
		*resultBuf, err = page.CaptureScreenshot().WithClip(&viewport).Do(ctx)
		log.Print("... screenshot done.")
		return err
	})
}

/*
	pdfAction returns a runnable chromedp.Action that renders the current context as a pdf file.
	The results are saved to the resultBuf
*/
func pdfAction(resultBuf *[]byte) chromedp.Action {
	// Use a chromedp.ActionFunc to build an executable function
	return chromedp.ActionFunc(func(ctx context.Context) error { // The context is set when Run calls Do for each each Action
		log.Print("Rendering pdf...")
		var err error
		*resultBuf, err = page.PrintToPDF().WithLandscape(true).Do(ctx)
		log.Print("... pdf done.")
		return err

	})
}
