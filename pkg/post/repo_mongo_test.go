package post

import (
	context "context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/greatjudge/redditclone/pkg/comment"
	"github.com/greatjudge/redditclone/pkg/user"
	"github.com/greatjudge/redditclone/pkg/vote"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var Posts []Post = []Post{
	{
		ID:    "1",
		Title: "title1",
		Views: 2,
		Type:  "text",
		URL:   "",
		Text:  "kkfkdlfkdf",
		Author: user.User{
			ID:       "1",
			Username: "username",
		},
		Category: "music",
		Votes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		Comments: []comment.Comment{
			{
				Created: CreationTime(),
				Author: user.User{
					ID:       "1",
					Username: "username",
				},
				ID:   "1",
				Body: "New Comment",
			},
		},
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	},
	{
		ID:    "2",
		Title: "title2",
		Views: 2,
		Type:  "text",
		URL:   "",
		Text:  "kkfkdlfkdf",
		Author: user.User{
			ID:       "2",
			Username: "username2",
		},
		Category: "programming",
		Votes: []vote.Vote{
			{
				UserID: "2",
				Value:  1,
			},
		},
		Comments: []comment.Comment{
			{
				Created: CreationTime(),
				Author: user.User{
					ID:       "2",
					Username: "username2",
				},
				ID:   "2",
				Body: "New Comment 2",
			},
		},
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	},
	{
		ID:    "3",
		Title: "title3",
		Views: 2,
		Type:  "url",
		URL:   `http://exmaple.com`,
		Text:  "kkfkdlfkdf",
		Author: user.User{
			ID:       "1",
			Username: "username",
		},
		Category: "music",
		Votes: []vote.Vote{
			{
				UserID: "1",
				Value:  1,
			},
		},
		Comments: []comment.Comment{
			{
				Created: CreationTime(),
				Author: user.User{
					ID:       "1",
					Username: "username",
				},
				ID:   "3",
				Body: "New Comment",
			},
		},
		Created:          CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	},
}

func Post2bsonD(p Post) (bson.D, error) {
	data, err := bson.Marshal(p)
	if err != nil {
		return bson.D{}, err
	}
	bsonD := bson.D{}
	err = bson.Unmarshal(data, &bsonD)
	return bsonD, err
}

func BsonedPosts(posts []Post) ([]bson.D, error) {
	bsoned := make([]bson.D, len(posts))
	for i, p := range posts {
		bsonedPost, err := Post2bsonD(p)
		if err != nil {
			return nil, err
		}
		bsoned[i] = bsonedPost
	}
	return bsoned, nil
}

type SomeStruct struct {
	Name string
	Age  int
	Jobs []string
}

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	t.Run("some error", func(t *testing.T) {
		mockColl.EXPECT().Find(context.Background(), bson.D{}).Return(nil, fmt.Errorf("error"))
		_, err := repo.GetAll()
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		toReturn := make([]interface{}, len(Posts))
		for i, p := range Posts {
			toReturn[i] = p
		}
		cursor, err := mongo.NewCursorFromDocuments(toReturn, nil, nil)
		if err != nil {
			t.Errorf("create cursor err")
			return
		}

		mockColl.EXPECT().Find(context.Background(), bson.D{}).Return(cursor, nil)
		returned, err := repo.GetAll()
		assert.Nil(t, err)
		assert.Equal(t, Posts, returned)
	})

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	bsoned, err := Post2bsonD(Posts[0])
	if err != nil {
		t.Errorf("cant cast post to bson: %v", err.Error())
		return
	}

	mt.Run("error all", func(mt *mtest.T) {
		repo := NewMongoDBRepo(mt.Coll)
		cursorResposes := []primitive.D{
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bsoned),
			{},
			mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch),
		}
		mt.AddMockResponses(cursorResposes...)
		_, err = repo.GetAll()
		assert.NotNil(t, err)
		assert.Equal(t, "fail to get all posts command failed", err.Error())
	})
}

type TestCaseGetByID struct {
	Post        Post
	Error       error
	ReturnError error
	CaseName    string
}

func CheckGetByID(t *testing.T, tc TestCaseGetByID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}
	filter := bson.M{"_id": tc.Post.ID}

	singleResponse := mongo.NewSingleResultFromDocument(tc.Post, tc.Error, nil)
	mockColl.EXPECT().FindOne(context.Background(), filter).Return(singleResponse)

	if tc.Error == nil {
		tc.Post.Views += 1
		update := bson.M{"$set": bson.M{"views": tc.Post.Views}}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update)
	}

	post, err := repo.GetByID(tc.Post.ID)

	switch {
	case tc.Error != nil && tc.ReturnError != nil:
		assert.Equal(t, tc.ReturnError, err)
		return
	case tc.Error != nil:
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, tc.Post, post)
}

func TestGetByID(t *testing.T) {
	cases := []TestCaseGetByID{
		{
			Post:        Posts[0],
			Error:       nil,
			ReturnError: nil,
			CaseName:    "success",
		},
		{
			Post:        Post{},
			Error:       fmt.Errorf("some error"),
			ReturnError: nil,
			CaseName:    "some error",
		},
		{
			Post:        Post{},
			Error:       mongo.ErrNoDocuments,
			ReturnError: ErrNoPost,
			CaseName:    "no post err",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckGetByID(t, tc)
		})
	}
}

func TestGetByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	t.Run("some error", func(t *testing.T) {
		category := "music"
		mockColl.EXPECT().Find(context.Background(), bson.M{"category": category}).Return(nil, fmt.Errorf("error"))
		_, err := repo.GetByCategory(category)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		toReturn := make([]interface{}, len(Posts))
		for i, p := range Posts {
			toReturn[i] = p
		}
		cursor, err := mongo.NewCursorFromDocuments(toReturn, nil, nil)
		if err != nil {
			t.Errorf("create cursor err")
			return
		}

		category := "category"
		mockColl.EXPECT().Find(context.Background(), bson.M{"category": category}).Return(cursor, nil)
		returned, err := repo.GetByCategory(category)
		assert.Nil(t, err)
		assert.Equal(t, Posts, returned)
	})

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	bsoned, err := Post2bsonD(Posts[0])
	if err != nil {
		t.Errorf("cant cast post to bson: %v", err.Error())
		return
	}

	mt.Run("error all", func(mt *mtest.T) {
		repo := NewMongoDBRepo(mt.Coll)
		cursorResposes := []primitive.D{
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bsoned),
			{},
			mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch),
		}
		mt.AddMockResponses(cursorResposes...)
		category := "programming"
		_, err = repo.GetByCategory(category)
		assert.NotNil(t, err)
		assert.Equal(t, "fail to get all posts command failed", err.Error())
	})
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	t.Run("some error", func(t *testing.T) {
		post := Posts[0]
		mockColl.EXPECT().InsertOne(context.Background(), gomock.Any()).Return(nil, fmt.Errorf("error"))
		_, err := repo.Add(post)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		post := Posts[0]
		mockColl.EXPECT().InsertOne(context.Background(), gomock.Any()).Return(nil, nil)
		returned, err := repo.Add(post)
		assert.Nil(t, err)

		post.ID = returned.ID
		post.Created = returned.Created
		assert.Equal(t, post, returned)
	})
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	comm := comment.Comment{
		Author: post.Author,
		Body:   "some comment",
	}
	filter := bson.M{"_id": post.ID}

	t.Run("some error", func(t *testing.T) {
		mockColl.EXPECT().UpdateOne(context.Background(), filter, gomock.Any()).Return(nil, fmt.Errorf("error"))
		_, err := repo.AddComment(post.ID, comm)
		assert.NotNil(t, err)
	})

	t.Run("no modified", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 0,
		}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, gomock.Any()).Return(result, nil)
		_, err := repo.AddComment(post.ID, comm)
		assert.Equal(t, ErrNoPost, err)
	})

	t.Run("success", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
		}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, gomock.Any()).Return(result, nil)
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		returned, err := repo.AddComment(post.ID, comm)
		assert.Nil(t, err)
		assert.Equal(t, post, returned)
	})
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	commentID := "1"
	userID := post.Author.ID

	filter := bson.M{"_id": post.ID}
	pullFilter := bson.M{"comments": bson.M{"$and": bson.A{
		bson.M{"id": commentID},
		bson.M{"author.id": userID},
	}}}
	update := bson.M{"$pull": pullFilter}

	t.Run("some error", func(t *testing.T) {
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(nil, fmt.Errorf("error"))
		_, err := repo.DeleteComment(post.ID, commentID, userID)
		assert.NotNil(t, err)
	})

	t.Run("no modified", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
		}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)
		_, err := repo.DeleteComment(post.ID, commentID, userID)
		assert.Equal(t, ErrNoPost, err)
	})

	t.Run("no comment", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 0,
		}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)
		_, err := repo.DeleteComment(post.ID, commentID, userID)
		assert.Equal(t, comment.ErrNoComment, err)
	})

	t.Run("success", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
		}
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)

		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)

		returned, err := repo.DeleteComment(post.ID, commentID, userID)
		assert.Nil(t, err)
		assert.Equal(t, post, returned)
	})
}

func TestDBUpvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	userID := post.Author.ID
	post.Votes = []vote.Vote{
		{
			UserID: userID,
			Value:  1,
		},
	}

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}

	t.Run("err in GetByID", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(nil, fmt.Errorf("some err"), nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		_, err := repo.Upvote(post.ID, userID)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
		}
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)
		returned, err := repo.Upvote(post.ID, userID)
		assert.Nil(t, err)
		assert.Equal(t, post, returned)
	})

	t.Run("some rror", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(nil, fmt.Errorf("error"))
		_, err := repo.Upvote(post.ID, userID)
		assert.NotNil(t, err)
	})
}

func TestDBDownvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	userID := post.Author.ID
	post.Votes = []vote.Vote{
		{
			UserID: userID,
			Value:  -1,
		},
	}

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}

	t.Run("err in GetByID", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(nil, fmt.Errorf("some err"), nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		_, err := repo.Downvote(post.ID, userID)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
		}
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)
		returned, err := repo.Downvote(post.ID, userID)
		assert.Nil(t, err)
		assert.Equal(t, post, returned)
	})

	t.Run("some rror", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(nil, fmt.Errorf("error"))
		_, err := repo.Downvote(post.ID, userID)
		assert.NotNil(t, err)
	})
}

func TestDBUnvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	userID := post.Author.ID
	post.Votes = []vote.Vote{}

	filter := bson.M{"_id": post.ID}
	update := bson.M{"$set": post}

	t.Run("err in GetByID", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(nil, fmt.Errorf("some err"), nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		_, err := repo.Unvote(post.ID, userID)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		result := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
		}
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(result, nil)
		returned, err := repo.Unvote(post.ID, userID)
		assert.Nil(t, err)
		assert.Equal(t, post, returned)
	})

	t.Run("some rror", func(t *testing.T) {
		singleResponse := mongo.NewSingleResultFromDocument(post, nil, nil)
		mockColl.EXPECT().FindOne(context.Background(), bson.M{"_id": post.ID}).Return(singleResponse)
		mockColl.EXPECT().UpdateOne(context.Background(), filter, update).Return(nil, fmt.Errorf("error"))
		_, err := repo.Unvote(post.ID, userID)
		assert.NotNil(t, err)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	post := Posts[0]
	userID := post.Author.ID

	filter := bson.M{"$and": bson.A{
		bson.M{"_id": post.ID},
		bson.M{"author.id": userID},
	}}

	t.Run("some err", func(t *testing.T) {
		var r int64 = 0
		mockColl.EXPECT().DeleteOne(context.Background(), filter).Return(r, fmt.Errorf("error"))
		err := repo.Delete(post.ID, userID)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var r int64 = 1
		mockColl.EXPECT().DeleteOne(context.Background(), filter).Return(r, nil)
		err := repo.Delete(post.ID, userID)
		assert.Nil(t, err)
	})
}

func TestGetUserPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockColl := NewMockCollectionHelper(ctrl)
	repo := &PostMongoDBRepository{
		posts: mockColl,
	}

	t.Run("some error", func(t *testing.T) {
		username := "username"
		mockColl.EXPECT().Find(context.Background(), bson.M{"author.username": username}).Return(nil, fmt.Errorf("error"))
		_, err := repo.GetUserPosts(username)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		toReturn := make([]interface{}, len(Posts))
		for i, p := range Posts {
			toReturn[i] = p
		}
		cursor, err := mongo.NewCursorFromDocuments(toReturn, nil, nil)
		if err != nil {
			t.Errorf("create cursor err")
			return
		}

		username := "name"
		mockColl.EXPECT().Find(context.Background(), bson.M{"author.username": username}).Return(cursor, nil)
		returned, err := repo.GetUserPosts(username)
		assert.Nil(t, err)
		assert.Equal(t, Posts, returned)
	})

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	bsoned, err := Post2bsonD(Posts[0])
	if err != nil {
		t.Errorf("cant cast post to bson: %v", err.Error())
		return
	}

	mt.Run("error all", func(mt *mtest.T) {
		repo := NewMongoDBRepo(mt.Coll)
		cursorResposes := []primitive.D{
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bsoned),
			{},
			mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch),
		}
		mt.AddMockResponses(cursorResposes...)
		username := "newname"
		_, err = repo.GetUserPosts(username)
		assert.NotNil(t, err)
		assert.Equal(t, "fail to get all posts command failed", err.Error())
	})
}
