package main

import (
	"database/sql"
	"regexp"
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
    url text primary key
  );

  create table if not exists cache (
	  url text primary key,
    last_access timestamp,
    access_count int,
    is_allowed int
  );

  `
	_, err = db.Exec(create)
	if err != nil {
		logger.Fatalf("Failed to create db:", err)
		return
	}
}

func insertRegex(r string) {
	statement, err := db.Prepare("insert or ignore into whitelist values (?)")
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
	logger.Info("Checking host %v", host)
	var err error
	entry := WhitelistEntry{Domain: host, LastAccess: time.Now(), AccessCount: 0, IsAllowed: false}
	statement, _ := db.Prepare("select url, last_access, access_count, is_allowed from cache where url = ?")
	defer statement.Close()

	rows, _ := statement.Query(entry.Domain)
	rows.Close()

	if rows.Next() {
		if err = rows.Scan(&entry.Domain, &entry.LastAccess, &entry.AccessCount, &entry.IsAllowed); err != nil {
			logger.Error("Unable to scan results:", err)
		}
		logger.Info("Found a cache entry %v", entry)
	} else {
		statement, err := db.Prepare("insert into cache (url, last_access, access_count, is_allowed ) values (?,?,?,?)")
		if err != nil {
			logger.Error("Unable to prepare:", err)
		}
		defer statement.Close()
		statement.Exec(entry.Domain, entry.LastAccess, entry.AccessCount, entry.IsAllowed)

		// Need to loop over our regex's and check
		statement, _ = db.Prepare("select url from whitelist")
		whitelistRows, _ := statement.Query()
		var line string
		for whitelistRows.Next() {
			whitelistRows.Scan(&line)
			logger.Info("Looking at URL entry %v", line)
			re, _ := regexp.Compile(line)
			if re.MatchString(host) {
				entry.IsAllowed = true
				break
			}
		}
		whitelistRows.Close()
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
