package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	fileLocation = "./log/"
	fileMode     = 777
	defaultExp   = 60
)

var (
	currentPort        = "3030"
	isLog              = false
	defaultWriteSecond = 5
	timeout            = time.Duration(2 * time.Second)
	timeFormat         = time.RFC3339
)

type Response struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
}
type Options struct {
	CheckTime int
	Port      string
	IsLog     bool
	WriteTime int
}
type store struct {
	mu        sync.Mutex
	checkTime int
	kv        map[string]*value
}
type value struct {
	Expire int64
	Value  string
	Writed bool
}

func (s *store) set(k string, v string, exp int64) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, isOk := s.kv[k]
	if isOk {
		return fmt.Sprintf("%s already added", k), false
	}

	now := time.Now()
	secs := now.Unix()
	if exp == 0 {
		exp = defaultExp
	}
	s.kv[k] = &value{Expire: secs + exp, Value: v}
	return "success", true
}

func (s *store) get(key string) (*value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, isOk := s.kv[key]
	if isOk {
		return s.kv[key], true
	}
	return nil, false
}

func (s *store) handleStart() {

	fmt.Println("Listen port : " + currentPort)

	router := http.NewServeMux()
	server := &http.Server{
		Addr:         ":" + currentPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: timeout + 10*time.Millisecond, //10ms Redundant time
		IdleTimeout:  15 * time.Second,
	}
	router.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}
		// ctxx, _ := context.WithTimeout(context.Background(), time.Second*3)
		ctx, cancel := context.WithCancel(req.Context())
		go func() {
			resp := &Response{}
			// do something
			keys, ok := req.URL.Query()["key"]
			if !ok || len(keys[0]) < 1 {
				resp.Message = "Url Param 'key' is missing"
				resp.Success = false
				json.NewEncoder(w).Encode(resp)
				cancel()
				return
			}
			values, ok := req.URL.Query()["value"]
			if !ok || len(values[0]) < 1 {
				resp.Message = "Url Param 'value' is missing"
				resp.Success = false
				json.NewEncoder(w).Encode(resp)
				cancel()
				return
			}
			exp, ok := req.URL.Query()["expiration"]
			expTime := 0
			if ok {
				expTime, _ = strconv.Atoi(exp[0])
			}

			key := keys[0]
			value := values[0]
			message, isOk := s.set(key, value, int64(expTime))
			if !isOk {
				resp.Message = message
				resp.Success = false
				json.NewEncoder(w).Encode(resp)
				cancel()
				return
			}

			resp = &Response{Message: "success", Result: s.kv[key], Success: true}
			json.NewEncoder(w).Encode(resp)
			cancel()
		}()
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * timeout):
			fmt.Println("Timeout")
			return
			// json.NewEncoder(w).Encode("sss")
		}

	})
	router.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}
		// ctx, _ := context.WithTimeout(context.Background(), timeout)
		ctx, cancel := context.WithCancel(req.Context())
		go func() {
			resp := &Response{}

			key, ok := req.URL.Query()["key"]

			if !ok || len(key[0]) < 1 {
				resp.Message = "Url Param 'key' is missing"
				resp.Success = false
				json.NewEncoder(w).Encode(resp)
				cancel()
				return
			}
			searchKey := string(key[0])
			foundValue, _ := s.get(searchKey)
			if foundValue == nil {
				resp.Message = "Found nothing"
				resp.Success = false
				json.NewEncoder(w).Encode(resp)
				cancel()
				return
			}
			resp = &Response{Message: "success", Result: s.kv[searchKey], Success: true}
			json.NewEncoder(w).Encode(resp)
			cancel()
		}()
		select {
		case <-ctx.Done():
			// w.WriteHeader(http.StatusGatewayTimeout)
			return
		case <-time.After(time.Second * timeout):
			fmt.Println("Timeout")
			return
			// json.NewEncoder(w).Encode("sss")
		}

	})

	server.ListenAndServe()
}

func (s *store) checkExpired() {
	ticker := time.NewTicker(time.Second * time.Duration(s.checkTime))
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			for k, v := range s.kv {
				if t.Unix() > v.Expire {
					delete(s.kv, k)
				}
			}
		}
	}
}

func (s *store) Run() {

	if isLog {
		go s.writeToFile()
	}

	go s.checkExpired()

	// sigs := make(chan os.Signal, 1)
	// done := make(chan bool, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// go func() {
	// 	sig := <-sigs
	// 	fmt.Println()
	// 	fmt.Println(sig)
	// 	done <- true
	// }()

	// <-done
	s.handleStart()

}

func New(opt *Options) *store {
	if opt.Port != "" {
		currentPort = opt.Port
	}

	if opt.IsLog != isLog {
		isLog = opt.IsLog
	}

	if opt.WriteTime != 0 {
		defaultWriteSecond = opt.WriteTime
	}

	checkT := 1
	if opt.CheckTime != 0 {
		checkT = opt.CheckTime
	}

	s := make(map[string]*value)
	return &store{kv: s, checkTime: checkT}
}

func (s *store) writeToFile() {
	timestamp := fmt.Sprintf("%v", time.Now().Unix())
	logFile, err := os.OpenFile(fileLocation+timestamp+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, fileMode)

	ticker := time.NewTicker(time.Second * time.Duration(defaultWriteSecond))
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			s.mu.Lock()
			if err != nil {
				fmt.Println("opening log file failed")
				return
			}
			defer func() {

				err := logFile.Close()
				if err != nil {
					fmt.Println("closing log file failed")
				}

			}()

			if len(s.kv) > 0 {
				for k, v := range s.kv {
					if !v.Writed {
						kv := fmt.Sprintf("%s --> %s = %s ", t.Format(timeFormat), k, v.Value)
						_, err = logFile.WriteString(kv + "\n")
						if err != nil {
							fmt.Println("writing log failed")
						}
						v.Writed = true
					}
				}
			}

			s.mu.Unlock()
		}
	}

}
