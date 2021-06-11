package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDB(t *testing.T) {
	// TODO: make new file
	db, err := NewDB("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	assert.Equal(t,
		[]byte("color,blue"),
		db.makeRecord("color", "blue"),
		)


	if err := db.Set("", "value"); err == nil {
		t.Fatal("setting an empty key should result in an error")
	}

	// before there is a key, there should be no value present in the index for the key/offset pair

	testCases := []struct {
		Key         string
		IsNewKey    bool
		Expected    string
		Description string
	}{
		{"day", true, "monday", "set first key"},
		{"day", false, "tuesday", "overwrite first key"},
		{"time", true, "now", "set second key"},
		{"day", false, "wednesday", "overwrite first key after setting second"},
	}
	for i, testCase := range testCases {

		if testCase.IsNewKey {
			db.index[testCase.Key]
			_, exists, err := db.Get(testCase.Key)
			if err != nil {
				t.Fatalf("IsNewKey error: Get(%s) returned %s", testCase.Key, err.Error())
			}
			if exists {
				t.Fatalf("IsNewKey error: key exists but shoulnd't. key: '%s', test case: %s",
					testCase.Key, testCase.Description)
			}
		}
		if err := db.Set(testCase.Key, testCase.Expected); err != nil {
			t.Fatalf("Set(%s) error: %s", testCase.Key, err.Error())
		}
		got, exists, err := db.Get(testCase.Key)
		if err != nil {
			t.Fatalf("Get(%s) error: %s, test case: %s", testCase.Key, err.Error(), testCase.Description)
		}
		if !exists {
			t.Fatalf("Get(%s) erroneously reports that key doesn't exist. test case: %s", testCase.Key, testCase.Description)
		}
		assert.Equal(t, testCase.Expected, got, fmt.Sprintf("test case: %s", testCase.Description))
	}
}
