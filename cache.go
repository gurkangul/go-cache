package cache

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	defaultExp   = 10
	WriteTimeout = 1 * time.Second
)

type Options struct {
	CheckTime int
}
type Store struct {
	mu        sync.Mutex
	checkTime int
	kv        map[string]*Value
}
type Value struct {
	Expire int64
	Value  string
}

func (s *Store) Set(key string, value string, exp int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, isOk := s.kv[key]
	if isOk == true {
		return fmt.Sprintf("%s already added", key), false
	}

	now := time.Now()
	secs := now.Unix()
	if exp == 0 {
		exp = defaultExp
	}
	s.kv[key] = &Value{Expire: secs + exp, Value: value}
	return "success", true
}

func (s *Store) Get(key string) (*Value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, isOk := s.kv[key]
	if isOk {
		return s.kv[key], true
	}
	return nil, false
}

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
func (s *Store) HandleStart() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3030"
	}

	os.Stderr.WriteString("Listen port : " + port)
	router := http.NewServeMux()
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: WriteTimeout + 10*time.Millisecond, //10ms Redundant time
		IdleTimeout:  15 * time.Second,
	}
	router.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		ctx, _ := context.WithTimeout(context.Background(), WriteTimeout)
		worker, cancel := context.WithCancel(context.Background())
		var buffer string
		go func() {
			// do something
			time.Sleep(2 * time.Second)
			buffer = "ready all response\n"
			//do another
			time.Sleep(2 * time.Second)
			cancel()
			fmt.Printf("worker finish\n")
		}()
		select {
		case <-ctx.Done():
			//add more friendly tips
			w.WriteHeader(http.StatusInternalServerError)
			return
		case <-worker.Done():
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
			message, isOk := s.Set(key, value, 20)
			if isOk == false {
				fmt.Fprintf(w, message)

			}
			fmt.Println(buffer)
			fmt.Fprintf(w, "Key :%s , Value :%s", key, value)
			return
		}

	})
	router.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		key, ok := req.URL.Query()["key"]

		if !ok || len(key[0]) < 1 {
			fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
			return
		}
		searchKey := string(key[0])
		foundValue, _ := s.Get(searchKey)
		if foundValue == nil {
			fmt.Fprintf(w, "%s", "Found nothing")
			return
		}
		fmt.Printf("%#v \n", foundValue)
		fmt.Printf("%d \n", len(s.kv))
		fmt.Printf("%#v", s)

		fmt.Fprintf(w, "%s", foundValue.Value)

	})

	server.ListenAndServe()
}

func (s *Store) CheckExpired() {
	ticker := time.NewTicker(time.Second * time.Duration(s.checkTime))
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			for k, v := range s.kv {
				fmt.Println(k, v, t)
				if t.Unix() > v.Expire {
					delete(s.kv, k)
				}
			}
		}
	}
}

func New(opt *Options) *Store {
	store := make(map[string]*Value)
	return &Store{kv: store, checkTime: opt.CheckTime}
}
