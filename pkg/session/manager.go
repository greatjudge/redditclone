package session

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/greatjudge/redditclone/pkg/user"
	"go.uber.org/zap"
)

type SessionsManagerMemory struct {
	token2Session map[string]*Session
	mu            *sync.RWMutex
	secret        []byte
	lastID        int
	Logger        *zap.SugaredLogger
}

func NewSessionsManagerMemory(secret []byte, logger *zap.SugaredLogger) *SessionsManagerMemory {
	return &SessionsManagerMemory{
		token2Session: make(map[string]*Session, 10),
		mu:            &sync.RWMutex{},
		secret:        secret,
		lastID:        0,
		Logger:        logger,
	}
}

func (sm *SessionsManagerMemory) Create(user user.User) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(TokenExpirationTime).Unix(),
	})
	tokenString, err := token.SignedString(sm.secret)
	if err != nil {
		return "", fmt.Errorf("fail from SignedString: %w", err)
	}
	sm.lastID += 1
	session := NewSession(strconv.Itoa(sm.lastID), user)
	sm.token2Session[tokenString] = &session
	return tokenString, nil
}

func (sm *SessionsManagerMemory) Check(r *http.Request) (*Session, error) {
	inToken := r.Header.Get("Authorization")
	token, hasPref := strings.CutPrefix(inToken, "Bearer ")
	if !hasPref || token == "" {
		return nil, ErrNoAuth
	}
	return sm.Get(token)
}

func getExpirationTime(token jwt.Token) (time.Time, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, fmt.Errorf("can`t convert %v to jwt.MapClaims", token.Claims)
	}
	expirationTime, ok := claims["exp"]
	if !ok {
		return time.Time{}, fmt.Errorf("exp not in claims")
	}
	expTime := int64(expirationTime.(float64))
	return time.Unix(expTime, 0), nil
}

func (sm *SessionsManagerMemory) Get(tokenString string) (*Session, error) {
	hashSecretGetter := getHashSecretGetter(sm.secret)
	token, err := jwt.Parse(tokenString, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, ErrBadToken
	}

	exp, err := getExpirationTime(*token)
	if err != nil {
		sm.Logger.Error("in SessionsManagerMySQL.Get: ", err)
		return nil, ErrBadToken
	}
	tokenIsExpired := time.Now().After(exp)

	sm.mu.Lock()
	session, ok := sm.token2Session[tokenString]
	sm.mu.Unlock()

	if !ok || tokenIsExpired {
		return nil, ErrNoAuth
	}
	return session, nil
}
