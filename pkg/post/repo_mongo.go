package post

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/greatjudge/redditclone/pkg/comment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostMongoDBRepository struct {
	posts CollectionHelper
}

func NewMongoDBRepo(collecion *mongo.Collection) *PostMongoDBRepository {
	mc := &MongoCollection{
		Coll: collecion,
	}
	return &PostMongoDBRepository{
		posts: mc,
	}
}

func (repo *PostMongoDBRepository) GetAll() ([]Post, error) {
	posts := []Post{}
	c, err := repo.posts.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("fail to get posts %w", err)
	}
	err = c.All(context.Background(), &posts)
	if err != nil {
		return nil, fmt.Errorf("fail to get all posts %w", err)
	}
	return posts, nil
}

func (repo *PostMongoDBRepository) getPost(id string) (Post, error) {
	post := Post{}
	filter := bson.M{"_id": id}
	err := repo.posts.FindOne(context.Background(), filter).Decode(&post)
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return Post{}, ErrNoPost
	case err != nil:
		return Post{}, fmt.Errorf("fail to FindOne with id:%v, %w", id, err)
	}
	return post, nil
}

func (repo *PostMongoDBRepository) GetByID(id string) (Post, error) {
	post, err := repo.getPost(id)
	if err != nil {
		return post, err
	}
	post.Views += 1
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"views": post.Views}}
	_, err = repo.posts.UpdateOne(context.Background(), filter, update)
	return post, err
}

func (repo *PostMongoDBRepository) GetByCategory(category string) ([]Post, error) {
	posts := make([]Post, 0)
	c, err := repo.posts.Find(context.Background(), bson.M{"category": category})
	if err != nil {
		return nil, fmt.Errorf(`fail to find posts by caterory "%v" %w`, category, err)
	}
	err = c.All(context.Background(), &posts)
	if err != nil {
		return nil, fmt.Errorf("fail to get all posts %w", err)
	}
	return posts, nil
}

func (repo *PostMongoDBRepository) Add(post Post) (Post, error) {
	post.ID = uuid.NewString()
	post.Created = CreationTime()
	_, err := repo.posts.InsertOne(context.Background(), post)
	if err != nil {
		return Post{}, fmt.Errorf("fail to insert post: %w", err)
	}
	return post, nil
}

func (repo *PostMongoDBRepository) AddComment(id string, comm comment.Comment) (Post, error) {
	comm.ID = uuid.NewString()
	comm.Created = CreationTime()

	filter := bson.M{"_id": id}
	update := bson.M{"$push": bson.M{"comments": comm}}
	result, err := repo.posts.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return Post{}, err
	}
	if result.ModifiedCount == 0 {
		return Post{}, ErrNoPost
	}
	return repo.getPost(id)

}

func (repo *PostMongoDBRepository) DeleteComment(postID string, commentID string, userID string) (Post, error) {
	filter := bson.M{"_id": postID}
	pullFilter := bson.M{"comments": bson.M{"$and": bson.A{
		bson.M{"id": commentID},
		bson.M{"author.id": userID},
	}}}
	update := bson.M{"$pull": pullFilter}

	result, err := repo.posts.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return Post{}, err
	}
	switch {
	case result.MatchedCount == 0:
		return Post{}, ErrNoPost
	case result.ModifiedCount == 0:
		return Post{}, comment.ErrNoComment
	}
	return repo.getPost(postID)
}

func (repo *PostMongoDBRepository) Upvote(postID string, userID string) (Post, error) {
	post, err := repo.getPost(postID)
	if err != nil {
		return Post{}, err
	}
	post.Upvote(userID)
	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err = repo.posts.UpdateOne(context.Background(), filter, update)
	return post, err
}

func (repo *PostMongoDBRepository) Downvote(postID string, userID string) (Post, error) {
	post, err := repo.getPost(postID)
	if err != nil {
		return Post{}, err
	}
	post.Downvote(userID)
	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err = repo.posts.UpdateOne(context.Background(), filter, update)
	return post, err
}

func (repo *PostMongoDBRepository) Unvote(postID string, userID string) (Post, error) {
	post, err := repo.getPost(postID)
	if err != nil {
		return Post{}, err
	}
	post.Unvote(userID)
	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}
	_, err = repo.posts.UpdateOne(context.Background(), filter, update)
	return post, err
}

func (repo *PostMongoDBRepository) Delete(postID string, userID string) error {
	filter := bson.M{"$and": bson.A{
		bson.M{"_id": postID},
		bson.M{"author.id": userID},
	}}
	_, err := repo.posts.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostMongoDBRepository) GetUserPosts(username string) ([]Post, error) {
	posts := make([]Post, 0)
	c, err := repo.posts.Find(context.Background(), bson.M{"author.username": username})
	if err != nil {
		return nil, fmt.Errorf(`fail to find posts by author "%v", %w`, username, err)
	}
	err = c.All(context.Background(), &posts)
	if err != nil {
		return nil, fmt.Errorf("fail to get all posts %w", err)
	}
	return posts, nil
}
