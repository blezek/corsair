package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/elazarl/goproxy"
	lru "github.com/hashicorp/golang-lru"
	"github.com/olekukonko/tablewriter"
)

var (
	siteRegex    []*regexp.Regexp
	cache        *lru.Cache
	lock         sync.Mutex
	lastReadTime time.Time
	db           *bolt.DB
)

type WhitelistEntry struct {
	Domain      string
	LastAccess  time.Time
	AccessCount int64
	IsAllowed   bool
}

func whitelist(c *cli.Context) {
	if len(c.Args()) < 1 {
		log.Fatalf("Whitelist requires a file of hostnames to match")
	}
	configureLogging(c)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = c.GlobalBool("verbose")

	filename := c.Args().First()
	siteRegex = make([]*regexp.Regexp, 0)
	cache, _ = lru.New(1000)

	var err error
	db, err = bolt.Open("whitelist.db", 0600, nil)
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createBuckets()
	clearCache()
	go func() {
		for {
			content, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Print("Could not open file: %v -- %v", filename, err)
			}
			lines := strings.Split(string(content), "\n")
			lock.Lock()
			siteRegex = siteRegex[0:0]
			lastReadTime = time.Now()
			for idx, line := range lines {
				r, err := regexp.Compile(line)
				if err != nil {
					log.Print("Could not compile %v on line %v: %v", line, idx, err)
				}
				if line != "" {
					siteRegex = append(siteRegex, r)
					insertRegex(line)
				}
			}
			lock.Unlock()

			// Check the list
			// for _, host := range cache.Keys() {
			// 	if h, found := host.(string); found {
			// 		cache.Add(h, checkHost(h))
			// 	}
			// }
			time.Sleep(10 * time.Second)
		}
	}()

	var isBlacklist goproxy.ReqConditionFunc = func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		// log.Printf("Got request for %v", req.URL)
		entry := checkHost(req.Host)
		return !entry.IsAllowed
	}

	proxy.OnRequest(isBlacklist).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusForbidden,
				printCache(r))
		})

	proxy.OnRequest(isBlacklist).HandleConnect(goproxy.AlwaysReject)

	p := fmt.Sprintf(":%d", c.Int("control"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", reportCache)
	go http.ListenAndServe(p, mux)
	logger.Info("Started server on %v", p)
	p = fmt.Sprintf(":%d", c.Int("port"))
	log.Fatal(http.ListenAndServe(p, proxy))
}

func printCache(r *http.Request) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("The requested URL (%v) is not in the whitelist\n", r.URL))
	buffer.WriteString(fmt.Sprintf("\nThe accepted domains are (last read %v):\n", lastReadTime))
	for _, exp := range siteRegex {
		buffer.WriteString(fmt.Sprintf("\t%v\n", exp))
	}
	buffer.WriteString(fmt.Sprintf("\nThe cache contains the following entries:\n"))

	table := tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{"Site", "Count", "Date", "IsAllowed"})
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entry WhitelistEntry
			buffer := bytes.NewBuffer(v)
			gob.NewDecoder(buffer).Decode(&entry)
			table.Append([]string{entry.Domain, fmt.Sprintf("%v", entry.AccessCount), entry.LastAccess.Format(time.RubyDate), fmt.Sprintf("%v", entry.IsAllowed)})
		}
		return nil
	})
	// buffer.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\n", entry.Domain, entry.AccessCount, entry.LastAccess, entry.IsAllowed))
	table.Render()
	return buffer.String()

}

func reportCache(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, printCache(req))
}

func checkHost(host string) WhitelistEntry {
	lock.Lock()
	defer lock.Unlock()
	entry := WhitelistEntry{Domain: host, LastAccess: time.Now(), AccessCount: 0, IsAllowed: false}

	// Check using bolt
	return checkAndCache(entry)

	// for _, re := range siteRegex {
	// 	if re.MatchString(host) {
	// 		entry.IsAllowed = false
	// 		return entry
	// 	}
	// }
	// return entry

}
