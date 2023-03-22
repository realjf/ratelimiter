package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/realjf/ratelimiter"
)

func main() {

	cases := map[string]struct {
		run func()
	}{
		"fixed_window": {
			run: func() {
				limit := ratelimiter.NewFixedWindowRateLimiter(5, 1*time.Second)

				runHttpServer(limit)
			},
		},
		"sliding_window": {
			run: func() {
				limit := ratelimiter.NewSlidingWindowRateLimiter(100*time.Millisecond, 1*time.Second, 5)

				runHttpServer(limit)
			},
		},
		"leaky_bucket": {
			run: func() {
				limit := ratelimiter.NewLeakyBucketRateLimiter(5, 10)

				runHttpServer(limit)
			},
		},
		"token_bucket": {
			run: func() {
				limit := ratelimiter.NewTokenBucketRateLimiter(1, 10)
				runHttpServer(limit)
			},
		},
	}
	for name, ts := range cases {
		fmt.Println(name)
		go func() {
			time.Sleep(5 * time.Second)
			_, err := exec.Command("ab", "-n", "100", "-c", "20", "http://localhost:8888/").Output()
			if err != nil {
				fmt.Printf("error: %v\n", err.Error())
			}
			// fmt.Printf("success: %s\n", string(out))
			fmt.Println("Please use Ctrl+C to continue...")
		}()

		ts.run()
	}
}

func runHttpServer(limiter ratelimiter.RateLimiter) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			w.Write([]byte("hello world"))
			log.Printf("hello")
		}
	}))

	server := &http.Server{
		Addr:         "127.0.0.1:8888",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      mux,
	}

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done

		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("Shutdown server:", err)
		}
	}()

	log.Println("Starting HTTP server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Print("Server closed under request")
		} else {
			log.Fatal("Server closed unexpected")
		}
	}
}
