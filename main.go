package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
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
	kv map[string]string
}

func NewDB() *DB {
	return &DB{
		kv: make(map[string]string),
	}
}

func (db *DB) Set(key, value string) error {
	db.kv[key] = value
	return nil
}

func (db *DB) Get(key string) (string, bool, error) {
	val, exists := db.kv[key]
	return val, exists, nil
}

func main() {

	db := NewDB()

	r := mux.NewRouter()

	r.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, exists, err := db.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if !exists {
			http.NotFound(w, r)
		}
		w.Write([]byte(val))
	})

	r.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {

		// write the first value of the first key, ignoring the rest
		// NB: /set?s results in (k: s, v: "")

		for key, vals := range r.URL.Query() {
			for _, val := range vals {
				db.Set(key, val)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		s := fmt.Sprintf("invalid input: %s", r.URL.RawQuery)
		http.Error(w, s, http.StatusBadRequest)
	})

	http.ListenAndServe("localhost:4000", r)
}
