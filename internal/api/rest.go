package api

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
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
	w.WriteHeader(http.StatusOK)
}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
