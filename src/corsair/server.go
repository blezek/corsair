package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Start up the server, never return
func startServer(destination *url.URL) {
	// Set up a proxy object, and let it be chatty if needed
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Do we have a file?
		var file string
		pathWithoutLeadingSlash := req.URL.Path[1:]
		if strings.HasSuffix(pathWithoutLeadingSlash, "/") {
			// look for index.html
			file = filepath.Join(*staticFilesDirectory, pathWithoutLeadingSlash, "index.html")
		} else {
			file = filepath.Join(*staticFilesDirectory, pathWithoutLeadingSlash)
		}
		if *verbose {
			log.Printf("Get request for %v looking for file at %v\n", req.URL, file)
		}
		if _, err := os.Stat(file); err == nil {
			// Serve the file
			if *verbose {
				log.Printf("Found file %v and serving", file)
			}
			http.ServeFile(w, req, file)
		} else {
			// Proxy...
			req.URL.Scheme = destination.Scheme
			req.URL.Host = destination.Host
			if *verbose {
				log.Printf("Proxy request to %v", req.URL)
			}
			proxy.ServeHTTP(w, req)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
