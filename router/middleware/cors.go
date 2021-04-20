package middleware

import (
	"net/http"
	"os"
)

// Enables CORS for requests.
// Only the origins specified in the environment variable CORS_ORIGIN are allowed
// All methods and all headers are allowed.
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		allowedOrigins := os.Getenv("CORS_ORIGIN")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)

		next.ServeHTTP(w, r)
	})
}
