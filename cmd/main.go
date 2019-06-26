package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/elastifeed/es-scraper/internal/api"
	"github.com/elastifeed/es-scraper/internal/cdp"
	"github.com/elastifeed/es-scraper/internal/storage"
)

/*
 Entrypoint for es scraper. Configuration is done via enviroment:

 - API_BIND_SCRAPE: IP and port for server (eg. ":8080")
*/
func main() {
	flag.Parse()

	// Initiate storage backend
	s, err := storage.NewS3(&aws.Config{
		// @TODO get credentials and endpoint from environment, these are just test credentials :-P
		Credentials:      credentials.NewStaticCredentials("K279UGQBCW1RM3G1IITH", "s11DBCiqv9hnqoJ9drpEAQJkkBO2EP0Gv7u6MgLf", ""),
		Endpoint:         aws.String("http://localhost:30098"),
		Region:           aws.String("us-east-1"), // Somehow this is needed
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}, "elastitest", "http://localhost:30098/")

	if err != nil {
		log.Fatal(err)
	}

	r := api.InitRouter(s)

	server := &http.Server{
		Handler:      r,
		Addr:         ":8080", //os.Getenv("API_BIND_SCRAPE"),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	ctx, cancel := cdp.Launch() // Start a new headless chrome browser
	defer cancel()              // Defer closing the browser until main ends.

	go func() { // Run the server in a non - blocking goroutine
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	log.Print("Set up endpoint on", os.Getenv("API_BIND_SCRAPE"))

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	// Create a context based on the context of the browser. So, if the browser is closed, the server will shut down.
	context, stop := context.WithCancel(ctx)
	defer stop()

	<-c // Block until we recieve a signal on c

	server.Shutdown(context) // Shutdown the server gracefully.
}
