package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"k8s.io/klog"

	"github.com/elastifeed/es-scraper/internal/cdp"
	"github.com/elastifeed/es-scraper/internal/cdptab"
	"github.com/gorilla/mux"
)

// InitRouter initializes the router and defines routes.
func InitRouter() *mux.Router {
	r := mux.NewRouter()

	// Register parametrised route
	r.HandleFunc("/scrape/{action}", handler).Methods("POST")

	// Return the initialized router to the caller
	return r
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	action := vars["action"]

	if !isValidAction(action) {
		klog.Warning("Recieved request for invalid aciton \"%s\"", action)
		w.WriteHeader(http.StatusNotFound)
		w.Write(*responseError(errors.New("Path not found")))
		return
	}

	url, err := decodeRequest(r) // Decode the incoming
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}

	// Enqueue request, make a channel for the result and block until the result has arrived
	callback := make(chan cdptab.ChromeTabReturns)
	cdp.Enqueue(action, url, callback)
	result := <-callback

	if result.Err != nil {
		klog.Error(result.Err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(result.Err))
		return
	}

	data, err := json.Marshal(result.Data)
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(*responseError(err))
		return
	}
	klog.Infof("Got result for %s on %s", action, url)
	//klog.Info(string(data))

	n, err := w.Write(data)
	if err != nil {
		klog.Error("Written error", err)
	}
	klog.Infof("Written: n : %d", n)
	//io.WriteString(w, string(data))
}

func isValidAction(action string) bool {
	switch action {
	case "all", "screenshot", "pdf", "content":
		return true
	}
	return false
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
