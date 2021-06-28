package cache

import (
	"testing"
)

var cache = New(&Options{CheckTime: 1, IsHandle: true, IsLog: true})

func TestSet(t *testing.T) {
	msg, err := cache.set("gg", "bbb", 10)

	if !err {
		t.Errorf("%s", msg)
	}
}

func TestGet(t *testing.T) {
	value, err := cache.get("gg")

	if !err {
		t.Errorf("No record %v", value)
	}
}
