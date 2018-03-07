package accesslog

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/go-chi/chi/middleware"
	"github.com/yflau/gomw/chi/realip"
)

// NewAccessLogger record with givin logrus.Logger as go-chi middleware
func NewAccessLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			latencyMs := time.Since(start) * 1000
			logger.WithFields(logrus.Fields{
				"request": r.RequestURI,
				"method":  r.Method,
				"proto" :  r.Proto,
				"userAgent": r.UserAgent(),	
				"latencyMs" : latencyMs,
				"realip" : realip.GetRealIP(r.Context()),
				"status" : ww.Status(),
				"size" : ww.BytesWritten(),
			}).Info("completed handling request")
		}

		return http.HandlerFunc(fn)
	}
}

