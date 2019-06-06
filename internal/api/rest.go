package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/elastifeed/es-scraper/internal/cdp"
)

// InitRouter initializes the router and defines routes.
func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Grab a subrouter for the route "/scrape"
	base := r.PathPrefix("/scrape").Subrouter()
	// Define all the routes
	base.HandleFunc("/", allHandler)
	base.HandleFunc("/content", contentHandler)
	base.HandleFunc("/thumbnail", thumbnailHandler)
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

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	url, err := decodeRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	}
	Screenshot(url)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"good request\"}"))

}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func decodeRequest(r *http.Request) (string, error) {

	// Struct to define the shape of an incoming request
	type request struct {
		URL string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if decoder.Decode(&request) != nil {
		return "", errors.New("json: Error decoding json from request")
	}

	return request.URL, nil

}
