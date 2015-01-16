package main

import (
	"github.com/go-fsnotify/fsnotify"
	"log"
)

func livereloader(directory string) {

	log.Println("Starting livereloader @ %v", directory)
	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("Modifed file ", event.Name)
			case err := <-watcher.Errors:
				log.Println("Error: ", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}

	// Read forever
	<-done

}
