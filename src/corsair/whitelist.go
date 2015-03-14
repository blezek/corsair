package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
	lru "github.com/hashicorp/golang-lru"
)

func whitelist(c *cli.Context) {
	if len(c.Args()) < 1 {
		log.Fatalf("Whitelist requires a file of hostnames to match")
	}
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = c.Bool("verbose")

	filename := c.Args().First()
	regex := make([]*regexp.Regexp, 0)
	cache, _ := lru.New(1000)
	var lock sync.Mutex

	var checkHost = func(host string) bool {
		lock.Lock()
		defer lock.Unlock()
		for _, re := range regex {
			if re.MatchString(host) {
				return false
			}
		}
		return true

	}

	go func() {
		for {
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Print("Could not open file: %v -- %v", filename, err)
			}
			lines := strings.Split(string(content), "\n")
			lock.Lock()
			regex = regex[0:0]
			for idx, line := range lines {
				r, err := regexp.Compile(line)
				if err != nil {
					log.Print("Could not compile %v on line %v: %v", line, idx, err)
				}
				if line != "" {
					regex = append(regex, r)
				}
			}
			lock.Unlock()

			// Check the list

			for _, host := range cache.Keys() {
				if h, found := host.(string); found {
					cache.Add(h, checkHost(h))
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

	var isBlacklist goproxy.ReqConditionFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		// log.Printf("Got request for %v", req.URL)
		if v, ok := cache.Get(req.Host); ok {
			return v.(bool)
		} else {
			v := checkHost(req.Host)
			cache.Add(req.Host, v)
			return v
		}
	}

	proxy.OnRequest(isBlacklist).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("The requested URL (%v) is not in the whitelist\n", r.URL))
			buffer.WriteString(fmt.Sprintf("\nThe accepted domains are:\n"))
			for _, exp := range regex {
				buffer.WriteString(fmt.Sprintf("\t%v\n", exp))
			}
			buffer.WriteString(fmt.Sprintf("\nThe cache contains the following entries:\n"))
			for _, key := range cache.Keys() {
				v, ok := cache.Get(key)
				if ok {
					if v.(bool) {
						buffer.WriteString(fmt.Sprintf("\t%v is denied\n", key))
					} else {
						buffer.WriteString(fmt.Sprintf("\t%v is allowed\n", key))
					}

				}
			}

			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusForbidden,
				buffer.String())
		})

	proxy.OnRequest(isBlacklist).HandleConnect(goproxy.AlwaysReject)
	p := fmt.Sprintf(":%d", c.Int("port"))
	log.Fatal(http.ListenAndServe(p, proxy))
}
