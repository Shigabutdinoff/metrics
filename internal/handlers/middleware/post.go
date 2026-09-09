package middleware

import "net/http"

// EnsureMethodIsPost пропускает только POST-запросы.
func EnsureMethodIsPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Допустимы только POST-запросы!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
