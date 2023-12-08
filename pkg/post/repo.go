package post

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/greatjudge/redditclone/pkg/comment"
)

type PostMemoryRepository struct {
	id2Post map[string]*Post
	mu      *sync.RWMutex
}

func NewMemoryRepo() *PostMemoryRepository {
	return &PostMemoryRepository{
		id2Post: make(map[string]*Post),
		mu:      &sync.RWMutex{},
	}
}

func (repo *PostMemoryRepository) GetAll() ([]Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	posts := make([]Post, 0, len(repo.id2Post))
	for _, post := range repo.id2Post {
		posts = append(posts, *post)
	}
	return posts, nil
}

func (repo *PostMemoryRepository) GetByID(id string) (Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	post, ok := repo.id2Post[id]
	if !ok {
		return Post{}, ErrNoPost
	}
	post.Views += 1
	return *post, nil
}

func (repo *PostMemoryRepository) GetByCategory(category string) ([]Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	posts := make([]Post, 0, len(repo.id2Post))
	for _, post := range repo.id2Post {
		if post.Category == category {
			posts = append(posts, *post)
		}
	}
	return posts, nil
}

func (repo *PostMemoryRepository) Add(post Post) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	uid, err := uuid.NewUUID()
	if err != nil {
		return Post{}, fmt.Errorf("in post add: %w", err)
	}
	post.ID = uid.String()
	post.Created = CreationTime()
	repo.id2Post[post.ID] = &post
	return post, nil
}

func (repo *PostMemoryRepository) AddComment(postID string, comm comment.Comment) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return Post{}, ErrNoPost
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		return Post{}, fmt.Errorf("in post add comment: %w", err)
	}

	comm.ID = uid.String()
	comm.Created = CreationTime()
	post.Comments = append(post.Comments, comm)
	return *post, nil
}

func (repo *PostMemoryRepository) DeleteComment(postID string, commentID string, userID string) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return Post{}, ErrNoPost
	}

	commIdx := -1
	for i, comm := range post.Comments {
		if comm.ID == commentID {
			commIdx = i
			break
		}
	}
	if commIdx == -1 {
		return Post{}, comment.ErrNoComment
	}

	comm := post.Comments[commIdx]
	if comm.Author.ID != userID {
		return Post{}, ErrNoAccess
	}
	post.Comments[commIdx] = post.Comments[len(post.Comments)-1]
	post.Comments = post.Comments[:len(post.Comments)-1]
	return *post, nil
}

func (repo *PostMemoryRepository) Upvote(postID string, userID string) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return Post{}, ErrNoPost
	}
	post.Upvote(userID)
	return *post, nil
}

func (repo *PostMemoryRepository) Downvote(postID string, userID string) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return Post{}, ErrNoPost
	}
	post.Downvote(userID)
	return *post, nil
}

func (repo *PostMemoryRepository) Unvote(postID string, userID string) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return Post{}, ErrNoPost
	}
	post.Unvote(userID)
	return *post, nil
}

func (repo *PostMemoryRepository) Delete(postID string, userID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	post, ok := repo.id2Post[postID]
	if !ok {
		return ErrNoPost
	}
	if post.Author.ID != userID {
		return ErrNoAccess
	}
	delete(repo.id2Post, postID)
	return nil
}

func (repo *PostMemoryRepository) GetUserPosts(username string) ([]Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	posts := make([]Post, 0)
	for _, p := range repo.id2Post {
		if p.Author.Username == username {
			posts = append(posts, *p)
		}
	}
	return posts, nil
}
