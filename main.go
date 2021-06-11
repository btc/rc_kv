package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

/*
Before your interview,
write a program that runs a server that is accessible
on http://localhost:4000/.

When your server receives a request
on http://localhost:4000/set?somekey=somevalue
it should store the passed key and value in memory.

When it receives a request
on http://localhost:4000/get?key=somekey
it should return the value stored at somekey.

During your interview, you will pair on saving the data to a file.
You can start with simply appending each write to the file,
and work on making it more efficient if you have time.
*/

type DB struct {
	m  sync.Mutex
	file *os.File
	index map[string]int64
}

func NewDB(filename string) (*DB, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	db := &DB{
		index: make(map[string]int64),
		file: f,
	}

	if err := db.restoreIndex(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.file.Close()
}

func (db *DB) restoreIndex() error {
	offset, err := db.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	r := bufio.NewScanner(db.file)
	for r.Scan() {
		if r.Err() != nil {
			break
		}
		record := r.Bytes()
		key, _ := db.readRecord(record)
		db.index[key] = offset
		offset += int64(len(record)) + 1
	}
	return r.Err()
}

func (db *DB) Set(key, value string) error {
	db.m.Lock()
	defer db.m.Unlock()
	if key == "" {
		return errors.New("empty key not permitted")
	}

	data := db.makeRecord(key, value)
	offset, err := db.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	if _, err := db.file.Write(data); err != nil {
		return err
	}
	db.index[key] = offset

	return nil
}

func (db *DB) makeRecord(key, value string) []byte {
	// "color,blue"
	// "<5><4>colorblue"
	return []byte(strings.Join([]string{key, value}, ",")+"\n")
}

func (db *DB) readRecord(record []byte) (string, string) {
	pair := strings.SplitN(strings.TrimSuffix(string(record), "\n"), ",", 2)

	return pair[0], pair[1]
}

func (db *DB) Get(key string) (string, bool, error) {
	db.m.Lock()
	defer db.m.Unlock()

	offset, exists := db.index[key]
	if !exists {
		return "", false, nil
	}
	_, err := db.file.Seek(offset, io.SeekStart)
	if err != nil {
		return "", false, fmt.Errorf("seek error: %w", err)
	}
	r := bufio.NewScanner(db.file)
	r.Scan()
	if r.Err() != nil {
		return "", false, fmt.Errorf("read error: %w", r.Err())
	}
	record := r.Bytes()
	_, value := db.readRecord(record)

	return value, true, nil
}

func main() {

	db, err := NewDB("production.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, exists, err := db.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(val))
	})

	r.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {

		for key, vals := range r.URL.Query() {
			for _, val := range vals {

				// write the first value of the first key, ignoring the rest
				// NB: /set?s results in (k: s, v: "")

				if err := db.Set(key, val); err != nil {
					// TODO(btc): distinguish between internal and request errors
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		s := fmt.Sprintf("invalid input: %s", r.URL.RawQuery)
		http.Error(w, s, http.StatusBadRequest)
	})

	http.ListenAndServe("localhost:4000", r)
}
