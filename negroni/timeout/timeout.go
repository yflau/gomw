// Package timeout is a middleware adapted from go-chi/middleware for negroni
package timeout

import (
	"time"
	"context"
	"net/http"
)

// Timeout is a middleware that cancels ctx after a given timeout and return
// a 504 Gateway Timeout error to the client.
//
// It's required that you select the ctx.Done() channel to check for the signal
// if the context has reached its deadline and return, otherwise the timeout
// signal will be just ignored.
//
// ie. a route/handler may look like:
//  	n := negroni.New(
//      	negroni.NewRecovery(), 
//			accesslog.NewLogursLogger(config.AccessLogger),
//			negroni.NewStatic(http.Dir("public")),
//			gzip.Gzip(gzip.DefaultCompression),
//      )
//      var m = mux.NewRouter()
//      m.Handle("/tgraph", n.With(
//      	timeout.New(30 * time.Second),
//      	negroni.Wrap(&relay.Handler{Schema: schema}),
//      ))
// 
// Note: the original code is from `https://github.com/go-chi/chi/blob/master/middleware/timeout.go`
// 
type Timeout struct {
	timeout time.Duration
}

// New return a new Timeout middleware
func New(t time.Duration) *Timeout {
	return &Timeout{timeout: t}
}

func (t *Timeout) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), t.timeout)
	defer func() {
		cancel()
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
		}
	}()

	r = r.WithContext(ctx)
	next.ServeHTTP(w, r)
}

