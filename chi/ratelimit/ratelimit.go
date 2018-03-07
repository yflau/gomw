// Package ratelimit defines circuitbreaker which adapted from go-kit middleware for go-chi
package ratelimit

import (
	"context"
	"errors"
    "net/http"
)

// ErrLimited is returned in the request path when the rate limiter is
// triggered and the request is rejected.
var ErrLimited = errors.New("rate limit exceeded")

// Allower dictates whether or not a request is acceptable to run.
// The Limiter from "golang.org/x/time/rate" already implements this interface,
// one is able to use that in NewErroringLimiter without any modifications.
type Allower interface {
	Allow() bool
}

// NewErroringLimiter returns an chi.Middleware that acts as a rate
// limiter. Requests that would exceed the
// maximum request rate are simply rejected with an error.
func NewErroringLimiter(limit Allower) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if !limit.Allow() {
                http.Error(w, ErrLimited.Error(), http.StatusServiceUnavailable)
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// Waiter dictates how long a request must be delayed.
// The Limiter from "golang.org/x/time/rate" already implements this interface,
// one is able to use that in NewDelayingLimiter without any modifications.
type Waiter interface {
	Wait(ctx context.Context) error
}

// NewDelayingLimiter returns an chi.Middleware that acts as a
// request throttler. Requests that would
// exceed the maximum request rate are delayed via the Waiter function
//
// Demo:
//
// limit := rate.NewLimiter(rate.Every(time.Second), 100)
// r := chi.NewRouter()
// r.Use(middleware.RequestID)
// ...
// r.Group(func(r chi.Router) {
//	   r.Use(middleware.NewDelayingLimiter(limit))
//	   r.Get("/ratelimit", func(w http.ResponseWriter, r *http.Request) {
//	   	   w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	   	   w.Write(page)
//	   })
// })
//
func NewDelayingLimiter(limit Waiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if err := limit.Wait(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// AllowerFunc is an adapter that lets a function operate as if
// it implements Allower
type AllowerFunc func() bool

// Allow makes the adapter implement Allower
func (f AllowerFunc) Allow() bool {
	return f()
}

// WaiterFunc is an adapter that lets a function operate as if
// it implements Waiter
type WaiterFunc func(ctx context.Context) error

// Wait makes the adapter implement Waiter
func (f WaiterFunc) Wait(ctx context.Context) error {
	return f(ctx)
}
