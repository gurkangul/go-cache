package cache

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	fileLocation = "./log/"
	fileMode     = 777
	defaultExp   = 10
	WriteTimeout = 3 * time.Second
)

var (
	currentPort        = "3030"
	isLog              = false
	isHandle           = false
	defaultWriteSecond = 5
	timeout            = time.Duration(2 * time.Second)
)

type Options struct {
	CheckTime int
	Port      string
	IsHandle  bool
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
		WriteTimeout: WriteTimeout + 10*time.Millisecond, //10ms Redundant time
		IdleTimeout:  15 * time.Second,
	}
	router.HandleFunc("/set", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), WriteTimeout)
		worker, cancel := context.WithCancel(context.Background())
		go func() {
			// do something
			keys, ok := req.URL.Query()["key"]
			if !ok || len(keys[0]) < 1 {
				fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
				cancel()
				return
			}
			values, ok := req.URL.Query()["value"]
			if !ok || len(values[0]) < 1 {
				fmt.Fprintf(w, "%s", "Url Param 'value' is missing")
				cancel()
				return
			}

			key := keys[0]
			value := values[0]
			message, isOk := s.set(key, value, 20)
			if !isOk {
				fmt.Fprintf(w, "%s", message)
				cancel()
				return
			}
			fmt.Fprintf(w, "Key :%s , Value :%s", key, value)
			cancel()
		}()
		select {
		case <-ctx.Done():
			//add more friendly tips
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		case <-worker.Done():
			return
		}

	})
	router.HandleFunc("/get", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method Not Allowed")
			return
		}
		key, ok := req.URL.Query()["key"]

		if !ok || len(key[0]) < 1 {
			fmt.Fprintf(w, "%s", "Url Param 'key' is missing")
			return
		}
		searchKey := string(key[0])
		foundValue, _ := s.get(searchKey)
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

func (s *store) checkExpired() {
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

func (s *store) Run() {
	fmt.Println(isHandle, isLog)
	if isHandle {
		go s.handleStart()
	}

	if isLog {
		go s.writeToFile()
	}

	go s.checkExpired()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
}

func New(opt *Options) *store {
	if opt.Port != "" {
		currentPort = opt.Port
	}
	if opt.IsHandle != isHandle {
		isHandle = opt.IsHandle

	}
	if opt.IsLog != isLog {
		isLog = opt.IsLog
	}

	if opt.WriteTime != 0 {
		defaultWriteSecond = opt.WriteTime
	}

	s := make(map[string]*value)
	return &store{kv: s, checkTime: opt.CheckTime}
}

func (s *store) writeToFile() {
	ticker := time.NewTicker(time.Second * time.Duration(defaultWriteSecond))
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			s.mu.Lock()
			timestamp := fmt.Sprintf("%v", time.Now().Unix())
			logFile, err := os.OpenFile(fileLocation+timestamp+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, fileMode)
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

			for k, v := range s.kv {
				fmt.Println(k, v, t)
				if !v.Writed {
					kv := fmt.Sprintf("%s --> %s ", k, v.Value)
					_, err = logFile.WriteString(kv + "\n")
					if err != nil {
						fmt.Println("writing log failed")
					}
				}
			}
			s.mu.Unlock()
		}
	}

}
