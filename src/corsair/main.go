package main

import (
	"flag"
	"fmt"
<<<<<<< HEAD
	"github.com/dblezek/lrserver"
	"github.com/elazarl/goproxy"
	"io/ioutil"
=======
	"github.com/codegangsta/cli"
>>>>>>> Broke into subsections, adding proper cli parsing
	"log"
	"net/url"
	"os"
<<<<<<< HEAD
	"path/filepath"
	"regexp"
	"strings"
=======
>>>>>>> Broke into subsections, adding proper cli parsing
)

var (
	currentWorkingDirectory, _ = os.Getwd()
	// Flags
	port = flag.Int("port", 8080, "Serve static files on this port, falling back to the proxy if the file does not exist.")

	staticFilesDirectory = flag.String("dir", currentWorkingDirectory, fmt.Sprintf("Where to look for static files, defaults to current working directory (%v in this case)", currentWorkingDirectory))

	proxyDestination = flag.String("remote", "http://localhost:80", "If static files are not found, forward the request to the remote")

	help = flag.Bool("help", false, "Help")

	verbose = flag.Bool("verbose", false, "Verbose logging")

	livereload = flag.Bool("livereload", false, "Support live reload of the pages")
)

func init() {
	flag.IntVar(port, "p", 8080, "Short form of --port")
	flag.StringVar(staticFilesDirectory, "d", currentWorkingDirectory, "Short form of --dir")
	flag.StringVar(proxyDestination, "r", "http://localhost:80", "Short form of --remote")
	flag.BoolVar(help, "h", false, "Help")
	flag.BoolVar(verbose, "v", false, "Alias for --verbose")
	flag.BoolVar(livereload, "l", false, "Alias for --livereload")
}

func main() {

	cli.AppHelpTemplate = AppHelpTemplate
	app := cli.NewApp()
	app.Name = "corsair"
	readme, _ := Asset("Readme.md")

	app.Usage = "\n" + string(readme)
	app.Action = func(c *cli.Context) {
		println("Arrrgh me hearties")
		os.Exit(0)
	}
	app.Run(os.Args)
	os.Exit(0)
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

	log.Printf("Starting corsair in %v on port %v forwarding to %v://%v", *staticFilesDirectory, *port, destination.Scheme, destination.Host)
	log.Printf("Visit:\n\n    http://localhost:%d\n\nTo get started", *port)

<<<<<<< HEAD
	snippit := fmt.Sprintf(`<script>document.write('<script src="http://' + (location.host || 'localhost').split(':')[0] + ':%d/livereload.js?snipver=1"></' + 'script>')</script>`, *port)
	// Ignore case, look for "</body>", allow extra stuff in the tags
	expression, _ := regexp.Compile("(?i)</body[^>]+>")
	expression, _ = regexp.Compile("(?i)</body[^>]*>")

	// Set up a proxy object, and let it be chatty if needed
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	if !*verbose {
		log.SetOutput(ioutil.Discard)
		lrserver.Logger = log.New(ioutil.Discard, "[lrserver]", 0)
	}

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

	// Do we live reload?
	if *livereload {
		livereloader(*staticFilesDirectory)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	log.Print("Running!")
=======
	startServer(destination)
>>>>>>> Broke into subsections, adding proper cli parsing
}

// The text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `NAME:
   {{.Name}} - corsair is a small webserver to help write software that makes REST calls to a server, without having to run on the server

USAGE:
   {{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
HELP:
{{.Usage}}
`
