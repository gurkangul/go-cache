package cache

import (
	"net/http"
	"testing"
)

var (
	cache = New(&Options{CheckTime: 1, IsLog: false})
)

func TestSet(t *testing.T) {
	resp, err := http.Get("http://localhost:3030/set?key=foo&value=bar")

	if err != nil {
		t.Errorf("%v", resp)
	}
}

func TestGet(t *testing.T) {
	_, err := http.Get("http://localhost:3030/get?key=foo")

	if err != nil {
		t.Errorf("No record %v", "foo")
	}
}
