package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	FlagPassword          = flag.String("password", "password", "password which will be required by Authorization header")
	FlagCacheDuration     = flag.Duration("cacheDuration", 20*time.Minute, "duration of the cache response")
	FlagListenAddress     = flag.String("listenAddr", ":8080", "address to listen (for example: :8080 or 0.0.0.0:80")
	FlagHTTPClientTimeout = flag.Duration("httpClientTimeout", 10*time.Second, "timeout for the http client which calls the external API")
)

func main() {
	flag.Parse()
	lock := sync.Mutex{}
	cachedResp := ""
	validUntil := time.Time{}
	client := http.Client{
		Timeout: *FlagHTTPClientTimeout,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != *FlagPassword {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		lock.Lock()
		defer lock.Unlock()
		if validUntil.After(time.Now()) && cachedResp != "" {
			fmt.Fprint(w, cachedResp)
			return
		}

		resp, err := client.Get("https://api.ipify.org")
		if err != nil {
			log.Println(err)
			http.Error(w, "Error while retrieving IP", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error while retrieving IP", http.StatusInternalServerError)
			return
		}
		validUntil = time.Now().Add(*FlagCacheDuration)
		cachedResp = string(body)
		fmt.Fprint(w, cachedResp)
	})
	log.Println("Listening on", *FlagListenAddress)
	http.ListenAndServe(*FlagListenAddress, nil)
}
