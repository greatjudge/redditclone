package session

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"database/sql"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/greatjudge/redditclone/pkg/user"
	"go.uber.org/zap"
)

const MysqlDatetimeFormat = "2006-01-02 15:04:05"

type Sessions struct {
	Token      string
	UserID     string
	Expiration time.Time
}

type SessionsManagerMySQL struct {
	DB       *sql.DB
	UserRepo user.UserRepo
	secret   []byte
	Logger   *zap.SugaredLogger
}

func NewSessionsManagerMySQL(db *sql.DB, userRepo user.UserRepo, secret []byte, logger *zap.SugaredLogger) *SessionsManagerMySQL {
	return &SessionsManagerMySQL{
		DB:       db,
		UserRepo: userRepo,
		secret:   secret,
		Logger:   logger,
	}
}

func (sm *SessionsManagerMySQL) Create(user user.User) (string, error) {
	tokenExpireDate := time.Now().Add(TokenExpirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"iat":  time.Now().Unix(),
		"exp":  tokenExpireDate.Unix(),
	})
	tokenString, err := token.SignedString(sm.secret)
	if err != nil {
		return "", err
	}
	_, err = sm.DB.Exec(
		"INSERT IGNORE INTO sessions (`token`, `user_id`, `expiration`) VALUES (?, ?, ?)",
		tokenString,
		user.ID,
		tokenExpireDate.Format(MysqlDatetimeFormat),
	)
	if err != nil {
		sm.Logger.Error("in SessionsManagerMySQL Create: ", err)
		return "", ErrNoAuth
	}
	return tokenString, nil
}

func (sm *SessionsManagerMySQL) Check(r *http.Request) (Session, error) {
	inToken := r.Header.Get("Authorization")
	token, hasPref := strings.CutPrefix(inToken, "Bearer ")
	if !hasPref || token == "" {
		return Session{}, ErrNoAuth
	}
	return sm.Get(token)
}

func (sm *SessionsManagerMySQL) Get(tokenString string) (Session, error) {
	hashSecretGetter := getHashSecretGetter(sm.secret)
	token, err := jwt.Parse(tokenString, hashSecretGetter)
	if err != nil || !token.Valid {
		return Session{}, ErrBadToken
	}

	session, timeVal := &Sessions{}, ""
	err = sm.DB.
		QueryRow("SELECT token, user_id, expiration FROM sessions WHERE token = ?", tokenString).
		Scan(&session.Token, &session.UserID, &timeVal)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Session{}, ErrNoAuth
	case err != nil:
		sm.Logger.Error("in Get session: ", err)
		return Session{}, err
	}

	session.Expiration, err = time.Parse(MysqlDatetimeFormat, timeVal)
	if err != nil {
		sm.Logger.Error("in Get session, parse timeVal: ", err)
		return Session{}, err
	}

	if time.Now().After(session.Expiration) {
		_, err = sm.DB.Exec("DELETE FROM sessions WHERE token = ?", tokenString)
		if err != nil {
			sm.Logger.Error("in DELETE session: ", err)
		}
		return Session{}, ErrNoAuth
	}

	user, err := sm.UserRepo.GetByID(session.UserID)
	if err != nil {
		sm.Logger.Errorf("In SessionsManagerMySQL Get: user ID=%v not found", session.UserID)
		return Session{}, ErrNoAuth
	}

	return NewSession(session.Token, user), nil
}
