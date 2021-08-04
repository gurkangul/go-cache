package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

var (
	cache = New(&Options{CheckTime: 1, IsLog: false})
)

func init() {
	go cache.handleStart()
}
func TestSet(t *testing.T) {
	resp, err := http.Get("http://localhost:3030/set?key=foo&value=bar")
	if err != nil {
		fmt.Println(err)
		t.Errorf("%v", resp)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(err)
		t.Errorf("%v", resp)
	}
	fmt.Printf("%s \n", result)

}

func TestGet(t *testing.T) {
	resp, err := http.Get("http://localhost:3030/get?key=foo")
	if err != nil {
		t.Errorf("TestGet %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(err)
		t.Errorf("%v", resp)
	}
	if err != nil {
		fmt.Println(err)
		t.Errorf("%v", resp)
	}
	fmt.Printf("%s \n", result)
}
