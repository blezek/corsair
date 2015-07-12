package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Item struct {
	Id          int64     `json:"id"`
	URL         string    `json:"url"`
	AccessCount int       `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
}

type Items struct {
	Rows []Item `json:"rows"`
}

func registerRest(r *mux.Router) {
	// Get the list
	r.Path("/whitelist").Methods("GET").HandlerFunc(getItems)
	// Put a new item
	r.Path("/whitelist").Methods("POST").HandlerFunc(newItem)
	// Delete an item
	r.Path("/whitelist/{id}").Methods("DELETE").HandlerFunc(deleteItem)
}

func newItem(w http.ResponseWriter, request *http.Request) {
	var item Item
	err := json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	statement, err := db.Prepare("insert or ignore into whitelist values ( ? )")
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
	logger.Infof("Values to the query %v", values)

	// Do the select
	statement, err := db.Prepare("select rowid, url from whitelist")
	if err != nil {
		logger.Error("Error", err)
	}
	defer statement.Close()
	rows, err := statement.Query()
	if err != nil {
		logger.Error("Error", err)
	}
	items := Items{Rows: make([]Item, 0, 10)}
	for rows.Next() {
		item := Item{}
		rows.Scan(&item.Id, &item.URL)
		if err != nil {
			logger.Error("Error", err)
		}
		items.Rows = append(items.Rows, item)
	}
	json.NewEncoder(w).Encode(items)

}
func deleteItem(w http.ResponseWriter, request *http.Request) {

	rowid := mux.Vars(request)["id"]

	// Do the select
	statement, err := db.Prepare("delete from whitelist where rowid = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(rowid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
