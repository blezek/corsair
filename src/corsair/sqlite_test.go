package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQL(t *testing.T) {
	dbName := "/tmp/testing.db"
	os.Remove(dbName)
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatal("Failed to open db:", err)
		return
	}
	defer db.Close()

	create := `
     create table if not exists whitelist ( id integer not null primary key, regexp text);
     create table if not exists denied ( id integer not null primary key, url text, count integer, last_attempt date )`
	_, err = db.Exec(create)
	if err != nil {
		t.Fatal("Failed to create db:", err)
		return
	}

	statement, err := db.Prepare("insert into whitelist (regexp) values (?)")
	if err != nil {
		t.Fatal("Failed to prepare statement:", err)
		return
	}
	defer statement.Close()
	for i := 0; i < 100; i++ {
		if _, err = statement.Exec(fmt.Sprintf("site.%d", i)); err != nil {
			t.Fatal("Failed to insert statement:", err)
			return
		}
	}

	statement, err = db.Prepare("insert into denied values (?,?,?,?)")
	if err != nil {
		t.Fatal("Failed to prepare statement:", err)
		return
	}
	defer statement.Close()
	for i := 0; i < 100; i++ {
		if _, err = statement.Exec(nil, fmt.Sprintf("site.%d", i), 0, time.Now()); err != nil {
			t.Fatal("Failed to insert statement:", err)
			return
		}
	}

}
