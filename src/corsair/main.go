package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"github.com/skratchdot/open-golang/open"
)

// Package variables
var (
	verbose               = false
	timeoutInMilliseconds = 300
	silent                = false
	logger                = logging.MustGetLogger("corsair")
)

func init() {
	// Configure our logging
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{module} ▶ %{level:.5s} %{id:03x}%{color:reset} %{message}")
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)

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
			Usage: "Where to look for static files, defaults to current working directory",
		},
		cli.BoolFlag{
			Name:  "silent",
			Usage: "Don't print anything at all",
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
		cli.BoolFlag{
			Name:  "open,o",
			Usage: "Open web page in browser",
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
	app.Commands = []cli.Command{
		{
			Name:   "proxy",
			Usage:  "Run a whitelist proxy",
			Action: whitelist,
		},
	}
	app.Action = func(c *cli.Context) {
		// Set some variables
		if c.Bool("verbose") {
			logging.SetLevel(logging.DEBUG, "corsair")
		}
		if c.Bool("silent") {
			logging.SetLevel(logging.CRITICAL, "corsair")
		}
		port := c.Int("port")

		directory, _ := os.Getwd()
		if c.String("directory") != "" {
			directory = c.String("directory")
		}

		// do we have a valid directory?
		if _, err := os.Stat(directory); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Directory %v does not exist\n", directory)
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
			livereloader(directory)
		}

		logger.Info("Starting corsair in %v on port %v forwarding to %v://%v\n\nVisit:\n\n    http://localhost:%d\n\nTo get started\n\n", c.String("directory"), port, destination.Scheme, destination.Host, port)

		if c.Bool("open") {
			// Open the page
			open.Start(fmt.Sprintf("http://localhost:%d", port))
		}

		startServer(directory, destination, port)
	}

	app.Run(os.Args)

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
