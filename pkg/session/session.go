package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/greatjudge/redditclone/pkg/user"
)

const TokenExpirationTime = time.Hour * 24 * 7

var (
	ErrNoAuth   = errors.New("no session found")
	ErrBadToken = errors.New("bad token")
)

type Session struct {
	Token string
	User  user.User
}

func NewSession(token string, user user.User) Session {
	return Session{
		Token: token,
		User:  user,
	}
}

type sessKey string

var SessionKey sessKey = "sessionKey"

func SessionFromContext(ctx context.Context) (Session, error) {
	sess, ok := ctx.Value(SessionKey).(Session)
	if !ok {
		return Session{}, ErrNoAuth
	}
	return sess, nil
}

func ContextWithSession(ctx context.Context, sess Session) context.Context {
	return context.WithValue(ctx, SessionKey, sess)
}

func getHashSecretGetter(secret []byte) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("bad sign method")
		}
		return secret, nil
	}
}

//go:generate mockgen -source=session.go -destination=session_mock.go -package=session SessionManager
type SessionsManager interface {
	Create(user user.User) (string, error)
	Check(r *http.Request) (Session, error)
}
