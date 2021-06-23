package cache

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	defaultExp = 10
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
	if isOk {
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

func New(opt *Options) *Store {
	store := make(map[string]*Value)
	return &Store{kv: store, checkTime: opt.CheckTime}
}

func HandleStart() {
	http.HandleFunc("/set", setStore)
	http.HandleFunc("/get", getStore)
	http.ListenAndServe(":3030", nil)
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
