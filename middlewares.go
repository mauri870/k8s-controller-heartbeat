package k8sheartbeat

import (
	"net/http"
	"strings"

	limiter "github.com/ulule/limiter/v3"
	mhttp "github.com/ulule/limiter/v3/drivers/middleware/stdlib"
)

const ErrAuthorizationFailed = "authorization failed"

func authBasicMiddleware(token string) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

			var reqToken string
			if len(auth) == 2 && auth[0] == "Basic" {
				reqToken = auth[1]
			}

			if reqToken == "" {
				// fallback to the query param "token" field
				reqToken = r.URL.Query().Get("token")
			}

			// The caller probably does not know that this resource requires auth
			if reqToken == "" {
				rw.Header().Set("WWW-Authenticate", "Basic")
				http.Error(rw, ErrAuthorizationFailed, http.StatusUnauthorized)
				return
			}

			if reqToken != token {
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(rw, r)
		})
	}
}

func rateLimitMiddleware(limit *limiter.Limiter) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return mhttp.NewMiddleware(limit).Handler(next)
	}
}
