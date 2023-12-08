package user

import "errors"

var (
	ErrNoUser        = errors.New("no user found")
	ErrBadPass       = errors.New("invald password")
	ErrBadUserPass   = errors.New("bad username or password")
	ErrAlreadyExists = errors.New("user already exists")
)

type User struct {
	ID       string `json:"id" bson:"id"`
	Username string `json:"username" bson:"username"`
	password string
}

//go:generate mockgen -source=user.go -destination=repo_mock.go -package=user UserRepo
type UserRepo interface {
	Authorize(username, pass string) (User, error)
	Register(username, password string) (User, error)
	GetByID(userID string) (User, error)
}

func NewUser(id, username, password string) User {
	return User{
		ID:       id,
		Username: username,
		password: password,
	}
}
