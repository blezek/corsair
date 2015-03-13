package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
)

func whitelist(c *cli.Context) {
	if len(c.Args()) < 1 {
		log.Fatalf("Whitelist requires a file of hostnames to match")
	}
	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = true

	filename := c.Args().First()
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Could not open file: %v -- %v", filename, err)
	}
	lines := strings.Split(string(content), "\n")
	regex := make([]*regexp.Regexp, 0)
	for idx, line := range lines {
		r, err := regexp.Compile(line)
		if err != nil {
			log.Fatalf("Could not compile %v on line %v: %v", line, idx, err)
		}
		if line != "" {
			log.Printf("Adding %v to white list", r)
			regex = append(regex, r)
		}
	}

	var isBlacklist goproxy.ReqConditionFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		for _, re := range regex {
			if re.MatchString(req.Host) {
				return false
			}
		}
		return true
	}

	proxy.OnRequest(isBlacklist).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.Printf("Blocking request to %v", r.URL)
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusForbidden,
				"This URL is not on the whitelist.")
		})
	log.Fatal(http.ListenAndServe(":47010", proxy))
}
