package main

import (
	"bufio"
	"flag"
	"github.com/dblezek/lrserver"
	"github.com/go-fsnotify/fsnotify"
	"log"
	"os"
	"time"
)

var lastReloadRequest time.Time = time.Now()
var readInput = flag.Bool("read", false, "Read standard input to get file changes, helpful when using fswatch")
var readNull = flag.Bool("null", false, "Read looking for nulls (default is new line)")

var debounceTimeout = flag.Int("timeout", 300, "Time in milliseconds to wait for repeated file change events.  Typically programs like fswatch generate 3 file events per file change, it is good to wait a few milliseconds before sending the livereload request to prevent multiple, unwanted reloads.")

// Take a request, sleep on it.  If we were the last request, then
// tell the lrserver to reload, otherwise just go away.
// Only interval seconds after the last request will the lrserver actually be
// requested to reload
func request(fileName string, interval time.Duration) {
	time.Sleep(interval)
	if time.Now().After(lastReloadRequest.Add(interval)) {
		// now request the reload
		log.Println("Finally sending request")
		lrserver.Reload(fileName)
	}
}

func livereloader(directory string) {

	if verbose {
		log.Printf("Starting livereloader @ %v\n", directory)
	}
	watcher, err := fsnotify.NewWatcher()

	if *readInput {
		var terminationCharacter byte = byte('\n')
		if *readNull {
			terminationCharacter = 0
		}

		go func() {
			// Read from stdin and request a reload
			reader := bufio.NewReader(os.Stdin)
			for {
				f, _ := reader.ReadString(terminationCharacter)
				if verbose {
					log.Printf("Read %v\n", f)
				}
				go request(f, 300*time.Millisecond)
			}
		}()
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if verbose {
					log.Println("Modifed file ", event.Name)
				}
				// Record the time of this reload request, and queue it up
				lastReloadRequest = time.Now()
				go request(event.Name, 300*time.Millisecond)
			case err := <-watcher.Errors:
				log.Println("Error: ", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}

}
