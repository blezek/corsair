package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"log"
	"net/url"
	"os"
)

// Package variables
var (
	verbose               = false
	timeoutInMilliseconds = 300
)

func main() {
	currentWorkingDirectory, _ := os.Getwd()

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
		cli.IntFlag{
			Name:  "timeout,t",
			Value: timeoutInMilliseconds,
			Usage: "debounce timeout for live reload",
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
		if c.Bool("livereload") {
			timeoutInMilliseconds = c.Int("timeout")
			livereloader(c.String("directory"))
		}

		log.Printf("Starting corsair in %v on port %v forwarding to %v://%v", c.String("directory"), c.Int("port"), destination.Scheme, destination.Host)
		log.Printf("Visit:\n\n    http://localhost:%d\n\nTo get started", c.Int("port"))

		startServer(c.Bool("verbose"), c.String("directory"), destination, c.Int("port"))
	}

	app.Run(os.Args)
<<<<<<< HEAD
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
=======
>>>>>>> More flags, cleanup

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
