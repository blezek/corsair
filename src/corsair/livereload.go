package main

import (
	"bufio"
	"flag"
	"github.com/dblezek/lrserver"
	"github.com/go-fsnotify/fsnotify"
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
func requestLiveReload(fileName string) {
	interval := time.Duration(*debounceTimeout) * time.Millisecond
	time.Sleep(interval)
	if time.Now().After(lastReloadRequest.Add(interval)) {
		// now request the reload
		logger.Debug("Finally sending request")
		lrserver.Reload(fileName)
	}
}

func livereloader(directory string) {

	logger.Debug("Starting livereloader @ %v\n", directory)
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
				logger.Debug("Read %v\n", f)

				go requestLiveReload(f)
			}
		}()
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				logger.Debug("Modifed file ", event.Name)

				// Record the time of this reload request, and queue it up
				lastReloadRequest = time.Now()
				go requestLiveReload(event.Name)
			case err := <-watcher.Errors:
				logger.Error("Error: ", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		logger.Fatal(err)
	}

}
