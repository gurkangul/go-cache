package main

import (
	"fmt"
	"net/http"
)

var cache = New(1)

func main() {
	go cache.CheckExpired()
	http.HandleFunc("/set", setStore)
	http.HandleFunc("/get", getStore)
	http.ListenAndServe(":3030", nil)
}

func setStore(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["key"]

	if !ok || len(keys[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
		return
	}

	key := keys[0]

	// checkKey := cache.Get(key)
	// if checkKey != nil {
	// 	fmt.Fprintf(w, "%s already added", key)
	// 	return
	// }

	values, ok := req.URL.Query()["value"]

	if !ok || len(values[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'value' is missing")
		return
	}

	value := values[0]
	fmt.Println(key, value)
	cache.Set(key, value, 2)

	fmt.Fprintf(w, "Key :%s , Value :%s", key, value)

}

func getStore(w http.ResponseWriter, req *http.Request) {
	key, ok := req.URL.Query()["key"]

	if !ok || len(key[0]) < 1 {
		fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
		return
	}
	searchKey := string(key[0])
	foundValue := cache.Get(searchKey)
	// if foundValue == nil {
	// 	fmt.Fprintf(w, "%s", "Found nothing")
	// 	return
	// }
	fmt.Printf("%#v \n", foundValue)
	fmt.Printf("%#v", cache)
	fmt.Fprintf(w, "%s", foundValue)

}
