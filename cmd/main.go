package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/elastifeed/es-scraper/internal/api"
	"github.com/elastifeed/es-scraper/internal/cdp"
)

/*
 Entrypoint for es scraper. Configuration is done via enviroment:

 - API_BIND_SCRAPE: IP and port for server (eg. ":8080")
*/
func main() {
	flag.Parse()
	r := api.InitRouter()

	server := &http.Server{
		Handler:      r,
		Addr:         os.Getenv("API_BIND_SCRAPE"),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	ctx, cancel := cdp.Launch() // Start a new headless chrome browser
	defer cancel()              // Defer closing the browser until main ends.

	go func() { // Run the server in a non - blocking goroutine
		if err := server.ListenAndServe(); err != nil {
			glog.Fatal(err)
		}
	}()
	glog.Info("Set up endpoint on", os.Getenv("API_BIND_SCRAPE"))

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	// Create a context based on the context of the browser. So, if the browser is closed, the server will shut down.
	context, stop := context.WithCancel(ctx)
	defer stop()

	<-c // Block until we recieve a signal on c

	server.Shutdown(context) // Shutdown the server gracefully.

}
