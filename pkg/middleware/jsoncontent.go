package middleware

import (
	"net/http"

	"github.com/greatjudge/redditclone/pkg/sending"
)

func JSONContentCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			sending.SendJSONMessage(w, "unknown payload", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
