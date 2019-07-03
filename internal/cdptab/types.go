package cdptab

import (
	"context"

	"github.com/elastifeed/es-scraper/internal/storage"
)

// TabState is a type for state of a tab. Whether it is accepting new request or still busy with another request
type tabState int

// Possible states for a tab.
const (
	Accepting tabState = iota
	Busy
)
const mercuryParser = "http://localhost:8080/mercury/html"

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

// MercuryPayload is a simple struct containign data to be sent to mercury.
type MercuryPayload struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
}