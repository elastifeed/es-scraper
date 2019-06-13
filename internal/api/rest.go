package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastifeed/es-scraper/internal/cdp"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// InitRouter initializes the router and defines routes.
func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Grab a subrouter for the route "/scrape"
	base := r.PathPrefix("/scrape").Methods("POST").Subrouter()
	// Define all the routes
	base.HandleFunc("/", allHandler)
	base.HandleFunc("/screenshot", screenshotHandler)
	base.HandleFunc("/pdf", pdfHandler)

	// Return the initialized router to the caller
	return r
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	glog.Info(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func contentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	url, err := decodeRequest(r) // Decode the incoming
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}

	filePath, err := cdp.Screenshot(url) // Take the screenshot
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("{\"thumbnail_path\" : \"%s\"}", filePath)
	w.Write([]byte(resp))

}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	url, err := decodeRequest(r) // Decode the incoming
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}

	filePath, err := cdp.Pdf(url) // Render the Pdf
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("{\"pdf_path\" : \"%s\"}", filePath)
	w.Write([]byte(resp))
}

func decodeRequest(r *http.Request) (string, error) {

	// Struct to define the shape of an incoming request
	var request struct {
		URL string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if decoder.Decode(&request) != nil {
		return "", errors.New("json: Error decoding json from request")
	}

	return request.URL, nil

}

func responseError(err error) *[]byte {
	msg := []byte(fmt.Sprintf("{\"status\": \"bad request \n%s\"}", err.Error()))
	return &msg
}
