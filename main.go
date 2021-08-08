package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

const (
	fileLocation = "./log/"
	fileMode     = 777
	defaultExp   = 60 // second
)

var (
	defaultPort        = "3030"
	isLog              = false //hasFileOutput
	defaultWriteSecond = 5
	timeout            = time.Duration(5 * time.Second)
	timeFormat         = time.RFC3339
)

// Response is http response
type Response struct {
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
}

// Options is initialize app options
type Options struct {
	CheckTime int
	Port      string
	IsLog     bool
	WriteTime int
}

//store
type store struct {
	mu        sync.Mutex
	checkTime int
	kv        map[string]*value
}

//value
type value struct {
	Expire int64
	Value  string
	Writed bool
}

//set key-value in memory
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

//get key-value from memory
func (s *store) get(key string) (*value, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, isOk := s.kv[key]
	if isOk {
		return s.kv[key], true
	}
	return nil, false
}

//handleStart
func (s *store) handleStart() {

	fmt.Println("Listen port : " + defaultPort)

	router := http.NewServeMux()
	server := &http.Server{
		Addr:           ":" + defaultPort,
		Handler:        router,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   timeout + 10*time.Millisecond, //10ms Redundant time
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 * 1024 * 1024,
	}
	router.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Println(req.Method)
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}

		resp := &Response{}
		// do something
		keys, ok := req.URL.Query()["key"]
		if !ok || len(keys[0]) < 1 {
			resp.Message = "Url Param 'key' is missing"
			resp.Success = false
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		values, ok := req.URL.Query()["value"]
		if !ok || len(values[0]) < 1 {
			resp.Message = "Url Param 'value' is missing"
			resp.Success = false
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		exp, ok := req.URL.Query()["expiration"]
		expTime := 0
		if ok {
			expTime, _ = strconv.Atoi(exp[0])
		}

		key := keys[0]
		value := values[0]
		message, ok := s.set(key, value, int64(expTime))
		if !ok {
			resp.Message = message
			resp.Success = false
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		resp = &Response{Message: "success", Result: s.kv[key], Success: true}
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println(err)
		}

	})
	router.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}
		resp := &Response{}

		key, ok := req.URL.Query()["key"]

		if !ok || len(key[0]) < 1 {
			resp.Message = "Url Param 'key' is missing"
			resp.Success = false
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		searchKey := string(key[0])
		foundValue, _ := s.get(searchKey)
		if foundValue == nil {
			resp.Message = "Found nothing"
			resp.Success = false
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		resp = &Response{Message: "success", Result: s.kv[searchKey], Success: true}

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println(err)
		}

	})

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

// check expired time key-value items
func (s *store) checkExpired(cc chan bool) {
	ticker := time.NewTicker(time.Second * time.Duration(s.checkTime))
	defer ticker.Stop()
	defer close(cc)
	for {
		select {
		case <-cc:
			log.Println("check expired close signal")
			return
		case t := <-ticker.C:
			for k, v := range s.kv {
				if t.Unix() > v.Expire {
					delete(s.kv, k)
				}
			}
		}
	}
}

// Application Run
func (s *store) Run() {
	log.Println("application running")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ww := make(chan bool)
	if isLog {
		// channel stop writing
		go s.writeToFile(ww)
	}

	cc := make(chan bool)
	go s.checkExpired(cc)

	go s.handleStart()

	select {
	case <-c:
		log.Println("interupt signal received")

	}

	if isLog {
		log.Println("stop writing...")
		ww <- false
		//wait until stop signal channel is closed (goroutine is finished)
		<-ww
		log.Println("stop writing channel closed")
	} else {
		close(ww)
		log.Println("empty channel closed")
	}

	log.Println("stop check sent")
	cc <- false
	//wait until stop signal channel is closed (goroutine is finished)
	<-cc
	log.Println("stop check channel closed")

	close(c)
	log.Println("interupt signal channel closed")
	log.Println("application stopped")
}

func New(opt *Options) *store {
	if opt.Port != "" {
		defaultPort = opt.Port
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

//writeToFile --> writing to file
func (s *store) writeToFile(ww chan bool) {

	timestamp := fmt.Sprintf("%v", time.Now().Unix())
	logFile, err := os.OpenFile(fileLocation+timestamp+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, fileMode)

	ticker := time.NewTicker(time.Second * time.Duration(defaultWriteSecond))
	defer ticker.Stop()
	defer close(ww)

	for {
		select {
		case <-ww:
			log.Println("writing close signal")
			return
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
