package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	currentWorkingDirectory, _ := os.Getwd()

	// Flags
	port := flag.Int("port", 8080, "Serve static files on this port, falling back to the proxy if the file does not exist.")
	flag.IntVar(port, "p", 8080, "Short form of --port")

	staticFilesDirectory := flag.String("dir", currentWorkingDirectory, fmt.Sprintf("Where to look for static files, defaults to current working directory (%v in this case)", currentWorkingDirectory))
	flag.StringVar(staticFilesDirectory, "d", currentWorkingDirectory, "Short form of --dir")

	proxyDestination := flag.String("remote", "http://localhost:80", "If static files are not found, forward the request to the remote")
	flag.StringVar(proxyDestination, "r", "http://localhost:80", "Short form of --remote")

	help := flag.Bool("help", false, "Help")
	flag.BoolVar(help, "h", false, "Help")

	verbose := flag.Bool("verbose", false, "Verbose logging")
	flag.BoolVar(verbose, "v", false, "Alias for --verbose")

	livereload := flag.Bool("livereload", false, "Support live reload of the pages")
	flag.BoolVar(livereload, "l", false, "Alias for --livereload")

	// Parse our flags!
	flag.Parse()

	flag.Usage = func() {
		fmt.Printf("\ncorsair is a small webserver to help write software that makes REST calls to a server, without having to run on the server\n\nUsage: %s [options]\n\nOPTIONS:\n", os.Args[0])
		flag.PrintDefaults()
		readme, _ := Asset("Readme.md")
		fmt.Printf("\n" + string(readme))
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// do we have a valid directory?
	if _, err := os.Stat(*staticFilesDirectory); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Directory %v does not exist\n", *staticFilesDirectory)
		flag.Usage()
		os.Exit(1)
	}

	// Can we parse the URL?
	var destination *url.URL
	var err error
	destination, err = url.Parse(*proxyDestination)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse %s\n  Error: %v\n", *proxyDestination, err)
		flag.Usage()
		os.Exit(1)
	}

	// Do we live reload?
	if *livereload {
		livereloader(*staticFilesDirectory)
	}

	log.Printf("Starting corsair in %v on port %v forwarding to %v://%v", *staticFilesDirectory, *port, destination.Scheme, destination.Host)
	log.Printf("Visit:\n\n    http://localhost:%d\n\nTo get started", *port)

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
