package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultExp = 10
)

type Store struct {
	mu        sync.Mutex
	checkTime int
	kv        map[string]*Value
}
type Value struct {
	Expire int64
	Value  string
}

func (s *Store) Set(key string, value string, exp int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	secs := now.Unix()
	fmt.Println(s.kv)

	if exp == 0 {
		exp = defaultExp
	}
	s.kv[key] = &Value{Expire: secs + exp, Value: value}
	fmt.Println(key, value, "----", s)
}

func (s *Store) Get(key string) *Value {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println(key)
	return s.kv[key]
}

func New(checkTime int) *Store {
	store := make(map[string]*Value)
	return &Store{kv: store, checkTime: checkTime}
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
