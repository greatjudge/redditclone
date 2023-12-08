package middleware

import (
	"net/http"

	"github.com/greatjudge/redditclone/pkg/sending"
	"github.com/greatjudge/redditclone/pkg/session"
)

func Auth(sm session.SessionsManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := sm.Check(r)
		if err != nil {
			sending.SendJSONMessage(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := session.ContextWithSession(r.Context(), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
