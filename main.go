package cache

import (
	"fmt"
	"net/http"
)

var cache = New(&Options{CheckTime: 1})

func init() {

}

func main() {
	go cache.CheckExpired()

}

func setStore(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
		return
	}
	values, ok := req.URL.Query()["value"]
	if !ok || len(values[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'value' is missing")
		return
	}

	key := keys[0]
	value := values[0]
	cache.Set(key, value, 20)
	fmt.Fprintf(w, "Key :%s , Value :%s", key, value)
}

func getStore(w http.ResponseWriter, req *http.Request) {
	key, ok := req.URL.Query()["key"]

	if !ok || len(key[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
		return
	}
	searchKey := string(key[0])
	foundValue, _ := cache.Get(searchKey)
	if foundValue == nil {
		fmt.Fprintf(w, "%s", "Found nothing")
		return
	}
	fmt.Printf("%#v \n", foundValue)
	fmt.Printf("%d \n", len(cache.kv))
	fmt.Printf("%#v", cache)
	fmt.Fprintf(w, "%s", foundValue)

}
