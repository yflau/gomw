package realip

import (
	"context"
	"net/http"

	"github.com/Tomasen/realip"
)

// Key to use when setting the request ID.
type ctxKeyRealIP int

// RealIPKey is the key that holds th Real IP in a request context.
const RealIPKey ctxKeyRealIP = 0

// RealIP is a middleware that sets a http.Request's Context Variable `RealIP` to the results
// of parsing either the X-Forwarded-For header or the X-Real-IP header (in that
// order).
//
// This middleware should be inserted fairly early in the middleware stack to
// ensure that subsequent layers (e.g., request loggers) which examine the
// RemoteAddr will see the intended value.
//
// You should only use this middleware if you can trust the headers passed to
// you (in particular, the two headers this middleware uses), for example
// because you have placed a reverse proxy like HAProxy or nginx in front of
// Goji. If your reverse proxies are configured to pass along arbitrary header
// values from the client, or if you use this middleware without a reverse
// proxy, malicious clients will be able to make you very sad (or, depending on
// how you're using RemoteAddr, vulnerable to an attack of some sort).
func RealIP(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rip := realip.FromRequest(r)
		ctx := context.WithValue(r.Context(), RealIPKey, rip)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// GetRealIP returns real IP from the given context if one is present.
// Returns the empty string if a real IP cannot be found.
func GetRealIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if rip, ok := ctx.Value(RealIPKey).(string); ok {
		return rip
	}
	return ""
}

