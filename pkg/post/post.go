package post

import (
	"errors"
	"time"

	"github.com/greatjudge/redditclone/pkg/comment"
	"github.com/greatjudge/redditclone/pkg/user"
	"github.com/greatjudge/redditclone/pkg/vote"
)

const (
	TEXT = "text"
)

var (
	ErrNoPost            = errors.New("no post found")
	ErrPostAlreadyExists = errors.New("post already exists")
	ErrNoAccess          = errors.New("no access")
)

type Post struct {
	ID               string            `json:"id" bson:"_id"`
	Title            string            `json:"title" bson:"title"`
	Views            int               `json:"views" bson:"views"`
	Type             string            `json:"type" bson:"type" valid:"required, in(text|link)"`
	URL              string            `json:"url,omitempty" bson:"url,omitempty" valid:"url"`
	Text             string            `json:"text,omitempty" bson:"text,omitempty"`
	Author           user.User         `json:"author" bson:"author"`
	Category         string            `json:"category" bson:"category"`
	Votes            []vote.Vote       `json:"votes" bson:"votes"`
	Comments         []comment.Comment `json:"comments" bson:"comments"`
	Created          string            `json:"created" bson:"created"`
	UpvotePercentage int               `json:"upvotePercentage" bson:"upvotePercentage"`
	Score            int               `json:"score" bson:"score"`
}

func (p *Post) SyncUpvotePercentage() {
	if len(p.Votes) != 0 {
		upvoteCnt := (p.Score + len(p.Votes)) / 2
		p.UpvotePercentage = upvoteCnt * 100 / len(p.Votes)
	} else {
		p.UpvotePercentage = 0
	}
}

func InitPost(post *Post, usr user.User) {
	post.Author = usr
	post.Comments = make([]comment.Comment, 0)
	post.Score = 1
	if post.Type == TEXT {
		post.URL = ""
	} else {
		post.Text = ""
	}
	post.Votes = []vote.Vote{
		{
			UserID: usr.ID,
			Value:  1,
		},
	}
	post.SyncUpvotePercentage()
}

func (p Post) findUserVote(userID string) (*vote.Vote, int) {
	for i, vote := range p.Votes {
		if vote.UserID == userID {
			return &p.Votes[i], i
		}
	}
	return nil, -1
}

func (p *Post) changeVotes(userID string, desiredValue int) {
	v, _ := p.findUserVote(userID)
	switch {
	case v == nil:
		{
			v = &vote.Vote{
				UserID: userID,
				Value:  desiredValue,
			}
			p.Votes = append(p.Votes, *v)
		}
	case v.Value == desiredValue:
		return
	default:
		p.Score -= v.Value
		v.Value = desiredValue
	}
	p.Score += v.Value
	p.SyncUpvotePercentage()
}

func (p *Post) Upvote(userID string) {
	p.changeVotes(userID, 1)
}

func (p *Post) Downvote(userID string) {
	p.changeVotes(userID, -1)
}

func (p *Post) Unvote(userID string) {
	v, i := p.findUserVote(userID)
	if v == nil {
		return
	}
	p.Score -= v.Value
	p.Votes[i] = p.Votes[len(p.Votes)-1]
	p.Votes = p.Votes[:len(p.Votes)-1]
	p.SyncUpvotePercentage()
}

//go:generate mockgen -source=post.go -destination=repo_mock.go -package=post PostRepo
type PostRepo interface {
	GetAll() ([]Post, error)
	GetByID(id string) (Post, error)
	GetByCategory(category string) ([]Post, error)
	Add(post Post) (Post, error)
	AddComment(id string, comm comment.Comment) (Post, error)
	DeleteComment(postID string, commentID string, userID string) (Post, error)
	Upvote(postID string, userID string) (Post, error)
	Downvote(postID string, userID string) (Post, error)
	Unvote(postID string, userID string) (Post, error)
	Delete(postID string, userID string) error
	GetUserPosts(username string) ([]Post, error)
}

func CreationTime() string {
	return time.Now().Format("2006-01-02T15:04:05.999Z")
}
