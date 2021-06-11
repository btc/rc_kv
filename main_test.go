package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDB(t *testing.T) {

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := NewDB(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	assert.Equal(t,
		[]byte("color,blue\n"),
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
	for _, testCase := range testCases {

		if testCase.IsNewKey {

			_, exists := db.index[testCase.Key]
			assert.False(t, exists, "index entry should not exist")

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

		_, exists := db.index[testCase.Key]
		assert.True(t, exists, "index entry should exist")

		got, exists, err := db.Get(testCase.Key)
		if err != nil {
			t.Fatalf("Get(%s) error: %s, test case: %s", testCase.Key, err.Error(), testCase.Description)
		}
		if !exists {
			t.Fatalf("Get(%s) erroneously reports that key doesn't exist. test case: %s", testCase.Key, testCase.Description)
		}
		assert.Equal(t, testCase.Expected, got, fmt.Sprintf("test case: %s", testCase.Description))
	}
	assert.Equal(t, 2, len(db.index))
}
