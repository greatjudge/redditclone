package comment

import (
	"errors"

	"github.com/greatjudge/redditclone/pkg/user"
)

var (
	ErrNoComment = errors.New("no comment found")
)

type Comment struct {
	Created string    `json:"created" bson:"created"`
	Author  user.User `json:"author" bson:"author"`
	ID      string    `json:"id" bson:"id"`
	Body    string    `json:"body" bson:"body"`
}

type CommentForm struct {
	Comment string `json:"comment"`
}
