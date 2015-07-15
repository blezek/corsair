package main

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	logger.Info("Starting initialize")
	var err error
	db, err = sql.Open("sqlite3", "whitelist.db")
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}

	create := `
  create table if not exists whitelist (
    id integer primary key,
    url text unique
  );

  create table if not exists cache (
    id integer primary key,
	  url text unique,
    last_access timestamp,
    access_count int,
    is_allowed int,
    whitelist_id integer with NULL
  );

  `
	_, err = db.Exec(create)
	if err != nil {
		logger.Fatalf("Failed to create db:", err)
		return
	}
}

func purgeCache() {
	go func() {
		for {
			result, err := db.Exec("delete from cache where last_access < datetime('now', '-3 day' )")
			if err != nil {
				logger.Error("Failed to prepare statement :", err)
			}
			if count, err := result.RowsAffected(); err != nil && count > 0 {
				logger.Info("Purged %v rows from cache", count)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
}

func insertRegex(r string) {
	statement, err := db.Prepare("insert or ignore into whitelist (url) values (?)")
	if err != nil {
		logger.Fatalf("Failed to prepare statement :", err)
		return
	}
	defer statement.Close()
	_, err = statement.Exec(r)
	if err != nil {
		logger.Fatalf("Failed to exec statement :", err)
		return
	}
}

func checkAndCache(entry WhitelistEntry) WhitelistEntry {
	// Do we alredy have the entry?
	entry.IsAllowed = true

	// insert it
	entry.AccessCount = entry.AccessCount + 1
	entry.LastAccess = time.Now()
	return entry
}

func checkHost(host string) WhitelistEntry {
	// Clip off port, if it's there
	index := strings.Index(host, ":")
	if index > 0 {
		host = host[:index]
	}
	logger.Info("Checking host %v", host)
	var err error
	entry := WhitelistEntry{Domain: host, LastAccess: time.Now(), AccessCount: 0, IsAllowed: false, Id: 0}
	// Localhost always works
	if host == "localhost" {
		entry.IsAllowed = true
		return entry
	}

	statement, err := db.Prepare("select id, url, last_access, access_count, is_allowed from cache where url = ?")
	if err != nil {
		logger.Error("Unable to prepare statement:", err)
	}
	defer statement.Close()

	rows, err := statement.Query(host)
	if err != nil {
		logger.Error("Unable to prepare statement:", err)
	}
	if rows.Next() {
		if err = rows.Scan(&entry.Id, &entry.Domain, &entry.LastAccess, &entry.AccessCount, &entry.IsAllowed); err != nil {
			logger.Error("Unable to scan results:", err)
		}
		logger.Info("Found a cache entry %v", entry)
	} else {

		// Need to loop over our regex's and check
		statement, _ = db.Prepare("select id, url from whitelist")
		whitelistRows, _ := statement.Query()
		var line string
		var id sql.NullInt64
		for whitelistRows.Next() {
			whitelistRows.Scan(&id, &line)
			line = strings.Replace(line, ".", "\\.", -1)
			line = strings.Replace(line, "*", ".*", -1)
			logger.Info("Looking at URL entry %v", line)
			re, _ := regexp.Compile(line)
			if re.MatchString(host) {
				entry.IsAllowed = true
				entry.WhitelistId = id
				break
			}
		}
		whitelistRows.Close()

		statement, err := db.Prepare("insert into cache (url, last_access, access_count, is_allowed, whitelist_id ) values (?,?,?,?,?)")
		if err != nil {
			logger.Error("Unable to prepare:", err)
		}
		defer statement.Close()
		statement.Exec(entry.Domain, entry.LastAccess, entry.AccessCount, entry.IsAllowed, entry.WhitelistId)

	}
	rows.Close()
	logger.Info("url %v is allowed? %v", host, entry.IsAllowed)
	entry.AccessCount += 1
	entry.LastAccess = time.Now()

	transaction, err := db.Begin()
	if err != nil {
		logger.Error("Error updating cache:", err)
	}

	updateStatement, err := transaction.Prepare("update cache set last_access = ?, access_count = ?, is_allowed = ? where url = ?")
	if err != nil {
		logger.Error("Error updating cache:", err)
	}
	defer updateStatement.Close()
	var isAllowed = 0
	if entry.IsAllowed {
		isAllowed = 1
	}
	_, err = updateStatement.Exec(entry.LastAccess, entry.AccessCount, isAllowed, entry.Domain)
	if err != nil {
		logger.Error("Error updating cache:", err)
	}
	logger.Info("Finished updating cache: %v", entry)
	err = transaction.Commit()
	if err != nil {
		logger.Error("Error updating cache:%v", err)
	}

	return entry

}
