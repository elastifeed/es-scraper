package cdptab

import (
	"context"

	"github.com/chromedp/cdproto/dom"
	"github.com/elastifeed/es-scraper/internal/storage"
)

// TabState is a type for state of a tab. Whether it is accepting new request or still busy with another request
type tabState int

// Possible states for a tab.
const (
	Accepting tabState = iota
	Busy
)
const userAgent = "Googlebot/2.1 (+http://www.google.com/bot.html)"

// ChromeTab describes the relevant data of a tab in a browser.
type ChromeTab struct {
	ID          uint                // Id of the browser tab
	Context     *context.Context    // Pointer to the context that describes this tab
	Stop        *context.CancelFunc // CancelFunc to cancel it's context
	URL         string              // The url that is loaded in this tab
	State       tabState            // The working state of this tab. Whether it is ready to accept a new request or still busy with another request
	Store       storage.Storager    // S3 Storage where PDFs etc get stored
	MercuryURL  string              // Mercury URL
	ContentSize *dom.Rect           // Size and position of the content in the frame
}

// ChromeTabReturns is a datastructure for the results of any operation on a tab.
type ChromeTabReturns struct {
	Data map[string]interface{} // Map containing unmarshaled json data.
	Err  error
}

// MercuryPayload is a simple struct containig data to be sent to mercury.
type MercuryPayload struct {
	URL string `json:"url"`
}

// ScreenshotPayload is a simple struct containing data to be sent for a screenshot
type ScreenshotPayload struct {
	URL     string                 `json:"url"`
	Options map[string]interface{} `json:"options"`
}
