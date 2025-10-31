package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	flagPassword := flag.String("password", "password", "password which will be required by Authorization header")
	flagCacheDuration := flag.Duration("cacheDuration", 20*time.Minute, "duration of the cache response")
	flagListenAddress := flag.String("listenAddr", ":8080", "address to listen (for example: :8080 or 0.0.0.0:80")
	flagHTTPClientTimeout := flag.Duration("httpClientTimeout", 10*time.Second, "timeout for the http client which calls the external API")
	flag.Parse()

	handler := ipHandlerfunc(*flagPassword, *flagHTTPClientTimeout, *flagCacheDuration)
	mux := http.NewServeMux()
	mux.Handle("GET /", handler)
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server := &http.Server{
		Addr:    *flagListenAddress,
		Handler: mux,
	}

	go func() {
		slog.Info("listening on", slog.String("address", server.Addr))

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", slog.String("error", err.Error()))
		}
		cancel()
	}()

	<-ctx.Done()

	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func ipHandlerfunc(password string, timeout, cacheDuration time.Duration) http.Handler {
	lock := sync.Mutex{}
	cachedResp := ""
	validUntil := time.Time{}
	client := &http.Client{
		Timeout: timeout,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != password {
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
			slog.Error("failed to retrieve IP", slog.String("error", err.Error()))
			http.Error(w, "Error while retrieving IP", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("failed to read response body", slog.String("error", err.Error()))
			http.Error(w, "Error while retrieving IP", http.StatusInternalServerError)
			return
		}

		validUntil = time.Now().Add(cacheDuration)
		cachedResp = string(body)
		fmt.Fprint(w, cachedResp)
	})
}
