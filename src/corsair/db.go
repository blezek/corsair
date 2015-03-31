package main

import (
	"bytes"
	"encoding/gob"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
)

func createBuckets() {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("user"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("whitelist"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("cache"))
		if err != nil {
			return err
		}
		return nil
	})
}

func insertRegex(r string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("whitelist"))
		logger.Debug("Putting %v into whitelist regexp", r)
		b.Put([]byte(r), []byte(r))
		return nil
	})
}

func clearCache() {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cache"))
		b.ForEach(func(k, v []byte) error {
			logger.Debug("Deleting %v from cache", string(k))
			return b.Delete(k)
		})
		return nil
	})
}

func checkAndCache(entry WhitelistEntry) WhitelistEntry {
	// Do we alredy have the entry?
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("cache"))
		val := bucket.Get([]byte(entry.Domain))
		if val != nil {
			buffer := bytes.NewBuffer(val)
			gob.NewDecoder(buffer).Decode(&entry)
			logger.Debug("Found cached entry: %v", entry)
		} else {
			// Is it a valid domain?
			wlBucket := tx.Bucket([]byte("whitelist"))
			entry.IsAllowed = false
			wlBucket.ForEach(func(k, v []byte) error {
				logger.Debug("Checking host %v against %v ", entry.Domain, string(k))
				r, err := regexp.Compile(string(k))
				if err == nil {
					entry.IsAllowed = entry.IsAllowed || r.MatchString(entry.Domain)
					if entry.IsAllowed {
						return nil
					}
				}
				return nil
			})
			logger.Debug("Inserting new cache entry %v", entry)
		}

		// insert it
		entry.AccessCount = entry.AccessCount + 1
		entry.LastAccess = time.Now()
		buffer := new(bytes.Buffer)
		gob.NewEncoder(buffer).Encode(entry)
		bucket.Put([]byte(entry.Domain), buffer.Bytes())

		return nil
	})
	return entry
}
