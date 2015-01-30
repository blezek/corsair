package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Start up the server, never return
func startServer(directory string, destination *url.URL, port int) {
	// Set up a proxy object, and let it be chatty if needed
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Do we have a file?
		var file string
		pathWithoutLeadingSlash := req.URL.Path[1:]
		if strings.HasSuffix(pathWithoutLeadingSlash, "/") {
			// look for index.html
			file = filepath.Join(directory, pathWithoutLeadingSlash, "index.html")
		} else {
			file = filepath.Join(directory, pathWithoutLeadingSlash)
		}
		logger.Debug("Get request for %v looking for file at %v\n", req.URL, file)

		if _, err := os.Stat(file); err == nil {
			// Serve the file
			http.ServeFile(w, req, file)
			logger.Debug("Found file %v and serving", file)
		} else {
			// Proxy...
			req.URL.Scheme = destination.Scheme
			req.URL.Host = destination.Host
			logger.Debug("Proxy request to %v", req.URL)

			proxy.ServeHTTP(w, req)
		}
	})

	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
