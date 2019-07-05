package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/elastifeed/es-scraper/internal/api"
	"github.com/elastifeed/es-scraper/internal/cdp"
	"github.com/elastifeed/es-scraper/internal/storage"
)

// getEnv is a helper to get a value from env or a default value
func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return def
}

/*
 Entrypoint for es scraper. Configuration is done via enviroment:

 - API_BIND_SCRAPE: IP and port for server (eg. ":8080")
*/
func main() {
	klog.InitFlags(nil)

	// Initiate storage backend
	store, err := storage.NewS3(&aws.Config{
		// @TODO get credentials and endpoint from environment, these are just test credentials :-P
		Credentials: credentials.NewStaticCredentials(
			getEnv("AWS_ACCESS_KEY_ID", ""),
			getEnv("AWS_SECRET_ACCESS_KEY", ""),
			"",
		),
		Endpoint:         aws.String(getEnv("AWS_ENDPOINT", "")),
		Region:           aws.String(getEnv("AWS_REGION", "us-east-1")), // Somehow this is needed
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}, getEnv("S3_BUCKET_NAME", ""), getEnv("S3_ENDPOINT", "http://localhost/"))

	if err != nil {
		klog.Fatal(err)
	}

	r := api.InitRouter()

	server := &http.Server{
		Handler: r,
		Addr:    getEnv("API_BIND", ":9090"),
	}

	ctx, cancel := cdp.Launch(getEnv("MERCURY_URL", "http://localhost:8080/mercury/html"), store) // Start a new headless chrome browser with s3 storage
	defer cancel()                                                                                // Defer closing the browser until main ends.

	go func() { // Run the server in a non - blocking goroutine
		if err := server.ListenAndServe(); err != nil {
			klog.Fatal(err)
		}
	}()
	klog.Info("Set up endpoint on", server.Addr)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Create a context based on the context of the browser. So, if the browser is closed, the server will shut down.
	context, stop := context.WithCancel(ctx)
	defer stop()

	<-c // Block until we recieve a signal on c

	server.Shutdown(context) // Shutdown the server gracefully.
}

