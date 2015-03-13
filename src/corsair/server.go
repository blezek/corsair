package main

import (
	"fmt"
	"github.com/elazarl/goproxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Start up the server, never return
func startServer(directory string, destination *url.URL, port int) {
	// Set up a proxy object, and let it be chatty if needed
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose

	snippit := fmt.Sprintf(`<script>document.write('<script src="http://' + (location.host || 'localhost').split(':')[0] + ':%d/livereload.js?snipver=1"></' + 'script>')</script>`, port)
	// Ignore case, look for "</body>", allow extra stuff in the tags
	expression, _ := regexp.Compile("(?i)</body[^>]+>")
	expression, _ = regexp.Compile("(?i)</body[^>]*>")

	handleRequest := func(w http.ResponseWriter, req *http.Request) {
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
			// Do we end with ".html?"
			if strings.HasSuffix(file, ".html") {
				// Inject some HTML before the </body> tag
				contents, _ := ioutil.ReadFile(file)
				contentString := expression.ReplaceAllString(string(contents), snippit+"$0")
				fmt.Fprint(w, contentString)
			} else {
				// spit it out man!
				http.ServeFile(w, req, file)
			}
			logger.Debug("Found file %v and serving", file)
		} else {
			// Proxy...
			req.URL.Scheme = destination.Scheme
			req.URL.Host = destination.Host
			logger.Debug("Proxy request to %v", req.URL)

			proxy.ServeHTTP(w, req)
		}
	}

	if addShutdownHook {
		http.HandleFunc("/corsair", func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case "DELETE":
				{
					logger.Info("Shutdown requested")
					os.Exit(0)
				}
			case "POST", "PUT":
				go requestLiveReload("/")
			default:
				handleRequest(w, req)
			}
		})
	}

	http.HandleFunc("/", handleRequest)

	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
