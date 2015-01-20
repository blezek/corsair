package main

import (
	"github.com/dblezek/lrserver"
	"github.com/go-fsnotify/fsnotify"
	"log"
	"time"
)

var lastReloadRequest time.Time = time.Now()

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

	log.Printf("Starting livereloader @ %v\n", directory)
	watcher, err := fsnotify.NewWatcher()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("Modifed file ", event.Name)
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
