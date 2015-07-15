package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
)

type Item struct {
	Id                  int64     `json:"id"`
	URL                 string    `json:"url"`
	URLPattern          string    `json:"url_pattern"`
	AccessCount         int       `json:"access_count"`
	LastAccess          time.Time `json:"last_access"`
	LastAccessUnix      int64     `json:"last_access_unix"`
	LastAccessHumanized string    `json:"last_access_humanized"`
	IsAllowed           bool      `json:"is_allowed"`
}

type Items struct {
	Rows      []Item `json:"rows"`
	TotalRows int    `json:"total_rows"`
}

func registerRest(r *mux.Router) {
	// Get the list
	r.Path("/whitelist").Methods("GET").HandlerFunc(getItems)
	// Put a new item
	r.Path("/whitelist").Methods("POST").HandlerFunc(newItem)
	// Delete an item
	r.Path("/whitelist/{id}").Methods("DELETE").HandlerFunc(deleteItem)
	// Get misses
	r.Path("/blacklist").Methods("GET").HandlerFunc(getBlacklist)
	r.Path("/blacklist/{id}").Methods("DELETE").HandlerFunc(deleteBlacklist)
}

func newItem(w http.ResponseWriter, request *http.Request) {
	var item Item
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	statement, err := db.Prepare("insert or ignore into whitelist (url) values ( ? )")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()
	result, err := statement.Exec(item.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item.Id, err = result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(item)
}

func getItems(w http.ResponseWriter, request *http.Request) {

	// Get the queries
	values := request.URL.Query()
	logger.Info("Values to the query %v", values)

	// Do the select
	statement, err := db.Prepare("select id, url from whitelist")
	if err != nil {
		logger.Error("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()
	rows, err := statement.Query()
	if err != nil {
		logger.Error("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items := Items{Rows: make([]Item, 0, 10)}
	for rows.Next() {
		item := Item{}
		rows.Scan(&item.Id, &item.URL)
		if err != nil {
			logger.Error("Error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items.Rows = append(items.Rows, item)
	}
	json.NewEncoder(w).Encode(items)

}

func deleteItem(w http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]

	// Do the select
	statement, err := db.Prepare("delete from cache where whitelist_id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Do the select
	statement, err = db.Prepare("delete from whitelist where id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getBlacklist(w http.ResponseWriter, request *http.Request) {

	// Get the queries
	values := request.URL.Query()
	logger.Info("Values to the query %v", values)

	// Do the select
	statement, err := db.Prepare("select id, url, access_count, last_access, is_allowed from cache where is_allowed = 0 order by last_access")
	if err != nil {
		logger.Error("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()
	rows, err := statement.Query()
	if err != nil {
		logger.Error("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items := Items{Rows: make([]Item, 0, 10)}
	for rows.Next() {
		item := Item{}
		err = rows.Scan(&item.Id, &item.URL, &item.AccessCount, &item.LastAccess, &item.IsAllowed)
		if err != nil {
			logger.Error("Error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		item.LastAccessHumanized = humanize.Time(item.LastAccess)
		item.LastAccessUnix = item.LastAccess.Unix()
		item.URLPattern = strings.Replace(item.URL, ".", "*.", -1)
		items.Rows = append(items.Rows, item)
	}
	rows.Close()

	// Count the rows
	statement, err = db.Prepare("select count(*) from cache")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err = statement.Query()
	if err != nil {
		logger.Error("Error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()
	for rows.Next() {
		err = rows.Scan(&items.TotalRows)
		if err != nil {
			logger.Error("Error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	rows.Close()
	json.NewEncoder(w).Encode(items)

}

func deleteBlacklist(w http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]

	logger.Info("Deleting %v from cache", id)
	// Do the delete
	statement, err := db.Prepare("delete from cache where id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	result, err := statement.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("Deleted with Result: %v", result)
	count, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("Purged %v rows from cache", count)
	fmt.Fprintf(w, "Deleted %v rows", count)
}
