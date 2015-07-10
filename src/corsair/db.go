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
  create table if not exists hit (
    url text primary key,
    access_count int,
    last_access timestamp
  );
  create table if not exists miss (
    url text primary key,
    access_count int,
    last_access timestamp
  );

  create table if not exists cache (
	  domain text primary key,
    last_access timestamp,
    access_count int,
    is_allowed boolean 
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
	var err error
	entry := WhitelistEntry{Domain: host, LastAccess: time.Now(), AccessCount: 0, IsAllowed: false}
	statement, _ := db.Prepare("select domain, last_access, access_count, is_allowed from cache where domain = ?")
	defer statement.Close()

	rows, _ := statement.Query(host)
	defer rows.Close()

	if rows.Next() {
		if err = rows.Scan(&entry.Domain, &entry.LastAccess, &entry.AccessCount, &entry.IsAllowed); err != nil {
			logger.Error("Unable to scan results:", err)
		}
	} else {
		statement, err := db.Prepare("insert into cache values (?,?,?,?)")
		if err != nil {
			logger.Error("Unable to prepare:", err)
		}
		statement.Exec(entry.Domain, entry.LastAccess, entry.AccessCount, entry.IsAllowed)
		defer statement.Close()

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
		defer whitelistRows.Close()

	}
	entry.AccessCount += 1
	entry.LastAccess = time.Now()
	statement, _ = db.Prepare("update cache set last_access = ?, access_count = ? where domain = ?")
	statement.Exec(entry.LastAccess, entry.AccessCount, entry.Domain)
	defer statement.Close()

	entry.LastAccess = time.Now()

	// Check using bolt
	return entry

	// for _, re := range siteRegex {
	// 	if re.MatchString(host) {
	// 		entry.IsAllowed = false
	// 		return entry
	// 	}
	// }
	// return entry

}
