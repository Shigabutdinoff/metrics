package middleware

import (
	"net/http"
	"strings"
)

// EnsureContentTypeIsTextPlain пропускает только запросы с text/plain.
func EnsureContentTypeIsTextPlain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(w, "Допустим только text/plain Content-Type !", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}
