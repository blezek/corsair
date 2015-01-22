package main

import (
	"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"log"
	"net/url"
	"os"
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
	app.Version = "1.0.0"
	app.EnableBashCompletion = true
	app.Author = "Daniel Blezek"
	app.Email = "daniel.blezek@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "dir,d,directory",
			Value: currentWorkingDirectory,
			Usage: "Where to look for static files, defaults to current working directory",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Verbose logging",
		},
		cli.IntFlag{
			Name:  "port,p",
			Value: 8080,
			Usage: "Port to serve static files and proxy to the remote",
		},
		cli.StringFlag{
			Name:  "remote,proxy,r",
			Value: "http://localhost:80",
			Usage: "Proxy destination",
		},
		cli.BoolFlag{
			Name:  "livereload,l",
			Usage: "Support livereload of the pages",
		},
	}

	app.Action = func(c *cli.Context) {
		// do we have a valid directory?
		if _, err := os.Stat(c.String("directory")); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Directory %v does not exist\n", c.String("directory"))
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		// Can we parse the URL?
		var destination *url.URL
		var err error
		destination, err = url.Parse(c.String("proxy"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse %s\n  Error: %v\n", c.String("proxy"), err)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		// Do we live reload?
		if *livereload {
			livereloader(c.String("directory"))
		}

		log.Printf("Starting corsair in %v on port %v forwarding to %v://%v", c.String("directory"), c.Int("port"), destination.Scheme, destination.Host)
		log.Printf("Visit:\n\n    http://localhost:%d\n\nTo get started", c.Int("port"))

		startServer(*staticFilesDirectory, destination, c.Int("port"))

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


	startServer(destination)
	log.Printf("Starting corsair in %v on port %v forwarding to %v://%v", *staticFilesDirectory, *port, destination.Scheme, destination.Host)
	log.Printf("Visit:\n\n    http://localhost:%d\n\nTo get started", *port)

	startServer(*staticFilesDirectory, destination, *port)
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
