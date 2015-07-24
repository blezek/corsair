package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"text/template"

	humanize "github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
)

type NewItem struct {
	Id          int64  `json:"id"`
	URL         string `json:"url"`
	Password    string `json:"password",omit_empty`
	Destination string `json:"destination",omit_empty`
}

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
	r.Path("/whitelist/post").Methods("POST").HandlerFunc(newItemPost)
	// Delete an item
	r.Path("/whitelist/{id}").Methods("DELETE").HandlerFunc(deleteItem)
	// Get misses
	r.Path("/blacklist").Methods("GET").HandlerFunc(getBlacklist)
	r.Path("/blacklist").Methods("DELETE").HandlerFunc(deleteBlacklistCache)
	r.Path("/blacklist/{id}").Methods("DELETE").HandlerFunc(deleteBlacklistItem)
}

func newItemPost(w http.ResponseWriter, request *http.Request) {
	var item NewItem
	var err error
	// Try post info
	logger.Info("Trying to parse the form")
	err = request.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		logger.Info("Parsed: %v", request.Form)
		item.Password = request.PostFormValue("password")
		item.URL = request.PostFormValue("url")
		item.Destination = request.PostFormValue("destination")
	}

	if !checkPassword(item.Password) {
		// Denied
		u, _ := url.Parse(item.Destination)
		buffer, _ := submitNewSite(u, "Incorrect password")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, buffer)
		return
	}

	item, err = createItem(item, w)
	if err != nil {
		return
	}
	logger.Info("Got password %v", item.Password)
	// Redirect to URL
	logger.Info("Redirecting to %v", item.Destination)

	deleteBlacklistMatching(item.URL)

	http.Redirect(w, request, item.Destination, http.StatusTemporaryRedirect)

}
func newItem(w http.ResponseWriter, request *http.Request) {
	var item NewItem
	var err error
	err = json.NewDecoder(request.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	item, err = createItem(item, w)
	if err != nil {
		return
	}
	json.NewEncoder(w).Encode(item)
}

func createItem(item NewItem, w http.ResponseWriter) (NewItem, error) {
	statement, err := db.Prepare("insert or ignore into whitelist (url) values ( ? )")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return item, err
	}
	defer statement.Close()
	result, err := statement.Exec(item.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return item, err
	}
	item.Id, err = result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return item, err
	}
	return item, nil
}

func getItems(w http.ResponseWriter, request *http.Request) {
	query := "select id, url from whitelist order by url "

	// page_size and start
	page_size, start, ok := getPaginationParameters(request.URL.Query())
	if ok {
		query = query + fmt.Sprintf(" limit %v offset %v ", page_size, start)
	}

	// Do the select
	statement, err := db.Prepare(query)
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
	// Count the rows
	statement, err = db.Prepare("select count(*) from whitelist")
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

	query := "select id, url, access_count, last_access, is_allowed from cache where is_allowed = 0 "

	columnMap := map[string]string{"url": "url", "access_count": "access_count", "last_access": "last_access"}
	var sort_by string = "url"
	var order string = "asc"
	a, ok := request.URL.Query()["sort_by"]
	if ok && len(a) > 0 {
		sort_by, ok = columnMap[a[0]]
		if !ok {
			sort_by = "url"
		}
	}
	a, ok = request.URL.Query()["descending"]
	if ok && len(a) > 0 {
		d, err := strconv.ParseBool(a[0])
		if err != nil && d {
			order = "desc"
		}
	}

	query = query + " order by " + sort_by + " " + order

	// page_size and start
	page_size, start, ok := getPaginationParameters(request.URL.Query())
	if ok {
		query = query + fmt.Sprintf(" limit %v offset %v ", page_size, start)
	}

	// Do the select
	statement, err := db.Prepare(query)
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
		item.URLPattern = makeURLPattern(item.URL)
		items.Rows = append(items.Rows, item)
	}
	rows.Close()

	// Count the rows
	statement, err = db.Prepare("select count(*) from cache where is_allowed = 0")
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

func deleteBlacklistMatching(url string) {
	statement, err := db.Prepare("delete from cache where url = ?")
	if err != nil {
		logger.Error("Error deleting from cache:", err)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(url)
	if err != nil {
		logger.Error("Error deleting from cache:", err)
	}

}

func deleteCache() (int64, error) {
	result, err := db.Exec("delete from cache")
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, err
}

func deleteBlacklistCache(w http.ResponseWriter, request *http.Request) {
	// Do the delete
	count, err := deleteCache()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("Purged %v rows from cache", count)
	fmt.Fprintf(w, "Deleted %v rows", count)
}

func deleteBlacklistItem(w http.ResponseWriter, request *http.Request) {
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

func makeURLPattern(url string) string {
	v := strings.Split(url, ".")
	if len(v) >= 2 {
		l := len(v)
		return v[l-2] + "*." + v[l-1]
	} else {
		return url
	}
}

func getPaginationParameters(values url.Values) (int, int, bool) {
	var page_size int
	var start int
	var err error
	page_size_array, page_size_ok := values["page_size"]
	if page_size_ok && len(page_size_array) > 0 {
		page_size, err = strconv.Atoi(page_size_array[0])
		page_size_ok = err == nil
	} else {
		page_size_ok = false
	}
	start_array, start_ok := values["start"]
	if start_ok && len(start_array) > 0 {
		start, err = strconv.Atoi(start_array[0])
		start_ok = err == nil
	} else {
		start_ok = false
	}
	return page_size, start, page_size_ok && start_ok
}

func submitNewSite(url *url.URL, message string) (string, error) {
	logger.Debug("Creating submit request for %v with message %v", url, message)
	var err error
	bodyBytes, err := Asset("post.html")
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	bootstrap_min_css, err := Asset("css/bootstrap.min.css")
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	bootstrap_theme_min_css, err := Asset("css/bootstrap-theme.min.css")
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	jumbotron_narrow_css, err := Asset("css/jumbotron-narrow.css")
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}

	data := map[string]string{
		"destination":             url.String(),
		"url":                     host,
		"bootstrap_min_css":       string(bootstrap_min_css),
		"bootstrap_theme_min_css": string(bootstrap_theme_min_css),
		"jumbotron_narrow_css":    string(jumbotron_narrow_css),
		"message":                 message,
	}
	tmpl, err := template.New("post").Parse(string(bodyBytes))
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		logger.Error("Error:%v", err)
		return "", err
	}
	return buffer.String(), nil
}
