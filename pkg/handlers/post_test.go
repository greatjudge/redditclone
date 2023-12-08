package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/golang/mock/gomock"
	"github.com/greatjudge/redditclone/pkg/comment"
	"github.com/greatjudge/redditclone/pkg/post"
	"github.com/greatjudge/redditclone/pkg/session"
	"github.com/greatjudge/redditclone/pkg/user"
	"github.com/greatjudge/redditclone/pkg/vote"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type ResponseWriterInner struct {
	Code int
}

type mockResponseWriter struct {
	mock.Mock
	inner *ResponseWriterInner
}

func (m *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.inner.Code = statusCode
}

func TestJSONMarshalAndSend(t *testing.T) {
	t.Run("error Marshall", func(t *testing.T) {
		w := httptest.NewRecorder()
		JSONMarshalAndSend(w, make(chan int))
		if w.Code != http.StatusInternalServerError {
			t.Errorf(
				"bad response code, expected %d, got %d",
				http.StatusOK,
				w.Code,
			)
			return
		}
	})

	t.Run("error write", func(t *testing.T) {
		inner := ResponseWriterInner{}
		mockW := mockResponseWriter{inner: &inner}
		mockW.On("Write", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("some error"))
		JSONMarshalAndSend(&mockW, map[string]string{"key": "val"})
		if mockW.inner.Code != http.StatusInternalServerError {
			t.Errorf(
				"bad response code, expected %d, got %d",
				http.StatusInternalServerError,
				mockW.inner.Code,
			)
			return
		}
	})

	t.Run("normal", func(t *testing.T) {
		w := httptest.NewRecorder()
		obj := map[string]string{"key": "val"}
		JSONMarshalAndSend(w, obj)
		if w.Code != http.StatusOK {
			t.Errorf(
				"bad response code, expected %d, got %d",
				http.StatusOK,
				w.Code,
			)
			return
		}

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("cant read body: %v", err)
		}
		writed := make(map[string]string, 0)
		err = json.Unmarshal(body, &writed)
		if err != nil {
			t.Errorf("cant unmarshall %v", body)
			return
		}
		if !reflect.DeepEqual(obj, writed) {
			t.Errorf("writed wrong obj, exected: %v, got %v", obj, writed)
		}
	})
}

var Posts []post.Post = []post.Post{
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
				Created: post.CreationTime(),
				Author: user.User{
					ID:       "1",
					Username: "username",
				},
				ID:   "1",
				Body: "New Comment",
			},
		},
		Created:          post.CreationTime(),
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
				Created: post.CreationTime(),
				Author: user.User{
					ID:       "2",
					Username: "username2",
				},
				ID:   "2",
				Body: "New Comment 2",
			},
		},
		Created:          post.CreationTime(),
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
				Created: post.CreationTime(),
				Author: user.User{
					ID:       "1",
					Username: "username",
				},
				ID:   "3",
				Body: "New Comment",
			},
		},
		Created:          post.CreationTime(),
		UpvotePercentage: 100,
		Score:            1,
	},
}

type TestCaseList struct {
	ReturnPosts []post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
}

func CheckTestList(t *testing.T, tc TestCaseList) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().GetAll().Return(tc.ReturnPosts, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	service.List(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	writedPosts := make([]post.Post, 0)
	err = json.Unmarshal(body, &writedPosts)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(tc.ReturnPosts, writedPosts) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPosts,
		)
	}
}

func TestList(t *testing.T) {
	cases := []TestCaseList{
		{
			ReturnPosts: Posts,
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
		},
		{
			ReturnPosts: make([]post.Post, 0),
			ReturnError: fmt.Errorf("Some Error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "error",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckTestList(t, tc)
		})
	}
}

type TestCaseListByCategory struct {
	ReturnPosts []post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
	Category    string
}

func CheckListByCategory(t *testing.T, tc TestCaseListByCategory) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().GetByCategory(tc.Category).Return(tc.ReturnPosts, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	req := httptest.NewRequest("GET", "/", nil)

	vars := map[string]string{
		"CATEGORY_NAME": tc.Category,
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	service.ListByCategory(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	writedPosts := make([]post.Post, 0)
	err = json.Unmarshal(body, &writedPosts)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(tc.ReturnPosts, writedPosts) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPosts,
		)
	}
}

func TestListByCategory(t *testing.T) {
	cases := []TestCaseListByCategory{
		{
			ReturnPosts: Posts,
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			Category:    "music",
		},
		{
			ReturnPosts: make([]post.Post, 0),
			ReturnError: fmt.Errorf("Some Error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "error",
			Category:    "some",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckListByCategory(t, tc)
		})
	}
}

type TestCaseGetByID struct {
	ReturnPost  post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
	PostID      string
}

func CheckGetByID(t *testing.T, tc TestCaseGetByID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().GetByID(tc.PostID).Return(tc.ReturnPost, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	req := httptest.NewRequest("GET", "/", nil)
	vars := map[string]string{
		"POST_ID": tc.PostID,
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	service.GetByID(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	writedPost := post.Post{}
	err = json.Unmarshal(body, &writedPost)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(tc.ReturnPost, writedPost) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPost,
		)
	}
}

func TestGetByID(t *testing.T) {
	cases := []TestCaseGetByID{
		{
			ReturnPost:  Posts[0],
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			PostID:      "1",
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: fmt.Errorf("Some Error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "error",
			PostID:      "0",
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "error",
			PostID:      "0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckGetByID(t, tc)
		})
	}
}

func CheckSessionError(t *testing.T, method string) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()

	switch method {
	case "Add":
		service.Add(w, req)
	case "AddComment":
		service.AddComment(w, req)
	case "DeleteComment":
		service.DeleteComment(w, req)
	case "Upvote":
		service.Upvote(w, req)
	case "Downvote":
		service.Downvote(w, req)
	case "Unvote":
		service.Unvote(w, req)
	case "Delete":
		service.Delete(w, req)
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
		return
	}
}

type TestCaseAdd struct {
	ReturnPost  post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
}

func CheckTestAdd(t *testing.T, tc TestCaseAdd) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := tc.ReturnPost
	post.InitPost(&p, p.Author)

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().Add(p).Return(p, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  p.Author,
	}

	postBytes, err := json.Marshal(p)
	if err != nil {
		t.Errorf("marshall err: %v", err.Error())
	}
	reader := bytes.NewReader(postBytes)
	req := httptest.NewRequest("POST", "/", reader)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.Add(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	writedPost := post.Post{}
	err = json.Unmarshal(body, &writedPost)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(p, writedPost) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPost,
		)
	}
}

func TestAdd(t *testing.T) {
	cases := []TestCaseAdd{
		{
			ReturnPost:  Posts[0],
			ReturnError: nil,
			StatusCode:  http.StatusCreated,
			CaseName:    "normal",
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusBadRequest,
			CaseName:    "some err",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckTestAdd(t, tc)
		})
	}
}

func TestAddSessionError(t *testing.T) {
	CheckSessionError(t, "Add")
}

func TestAddReadBodyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  user.User{},
	}

	mockR := mockReader{}
	mockR.On("Read", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("error reading"))

	req := httptest.NewRequest("POST", "/", &mockR)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.Add(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
		return
	}
}

func TestAddReadBodyUnmarshallErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  user.User{},
	}

	reader := strings.NewReader(".s/dw{")
	req := httptest.NewRequest("POST", "/", reader)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.Add(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusBadRequest,
			w.Code,
		)
		return
	}
}

type TestCaseAddComment struct {
	ReturnPost  post.Post
	CommForm    comment.CommentForm
	ReturnError error
	StatusCode  int
	CaseName    string
}

func TestAddCommentReadBodyErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  user.User{},
	}

	mockR := mockReader{}
	mockR.On("Read", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("error reading"))

	req := httptest.NewRequest("POST", "/", &mockR)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
		return
	}
}

func TestAddCommentUnmarshallError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  user.User{},
	}

	reader := strings.NewReader(".s/dw{")
	req := httptest.NewRequest("POST", "/", reader)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.AddComment(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusBadRequest,
			w.Code,
		)
		return
	}
}

func TestAddCommentSessionError(t *testing.T) {
	CheckSessionError(t, "AddComment")
}

func CheckAddComment(t *testing.T, tc TestCaseAddComment) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := tc.ReturnPost
	post.InitPost(&p, p.Author)

	commForm := tc.CommForm

	comm := comment.Comment{
		Author: p.Author,
		Body:   commForm.Comment,
	}

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().AddComment(p.ID, comm).Return(p, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  p.Author,
	}

	commFormBytes, err := json.Marshal(commForm)
	if err != nil {
		t.Errorf("marshall err: %v", err.Error())
	}
	reader := bytes.NewReader(commFormBytes)
	req := httptest.NewRequest("POST", "/", reader)
	vars := map[string]string{
		"POST_ID": p.ID,
	}
	req = mux.SetURLVars(req, vars)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.AddComment(w, req)

	resp := w.Result()

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusCreated,
			tc.StatusCode,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	writedPost := post.Post{}
	err = json.Unmarshal(body, &writedPost)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}

	if !reflect.DeepEqual(p, writedPost) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPost,
		)
	}
}

func TestAddComment(t *testing.T) {
	cases := []TestCaseAddComment{
		{
			ReturnPost: Posts[0],
			CommForm: comment.CommentForm{
				Comment: "some comment",
			},
			ReturnError: nil,
			StatusCode:  http.StatusCreated,
			CaseName:    "normal",
		},
		{
			ReturnPost:  post.Post{},
			CommForm:    comment.CommentForm{},
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
		},
		{
			ReturnPost:  post.Post{},
			CommForm:    comment.CommentForm{},
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckAddComment(t, tc)
		})
	}
}

type TestCaseDeleteComment struct {
	ReturnPost  post.Post
	CommID      string
	ReturnError error
	StatusCode  int
	CaseName    string
}

func TestDeleteCommentSessionError(t *testing.T) {
	CheckSessionError(t, "DeleteComment")
}

func CheckDeleteComment(t *testing.T, tc TestCaseDeleteComment) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := tc.ReturnPost
	commID := tc.CommID

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().DeleteComment(p.ID, commID, p.Author.ID).Return(p, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  p.Author,
	}

	req := httptest.NewRequest("POST", "/", nil)
	vars := map[string]string{
		"POST_ID":    p.ID,
		"COMMENT_ID": commID,
	}
	req = mux.SetURLVars(req, vars)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp := w.Result()

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	writedPost := post.Post{}
	err = json.Unmarshal(body, &writedPost)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(p, writedPost) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPost,
		)
	}
}

func TestDeleteComment(t *testing.T) {
	cases := []TestCaseDeleteComment{
		{
			ReturnPost:  Posts[0],
			CommID:      "1",
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
		},
		{
			ReturnPost:  post.Post{},
			CommID:      "1",
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
		},
		{
			ReturnPost:  post.Post{},
			CommID:      "1",
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
		},
		{
			ReturnPost:  post.Post{},
			CommID:      "1",
			ReturnError: comment.ErrNoComment,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no comment err",
		},
		{
			ReturnPost:  post.Post{},
			CommID:      "1",
			ReturnError: post.ErrNoAccess,
			StatusCode:  http.StatusForbidden,
			CaseName:    "no access err",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckDeleteComment(t, tc)
		})
	}
}

type TestCaseVote struct {
	ReturnPost  post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
	VoteMethod  string // Upvote, Downvote, Unvote
}

const (
	UPVOTE   = "Upvote"
	DOWNVOTE = "Downvote"
	UNVOTE   = "Unvote"
)

func CheckVote(t *testing.T, tc TestCaseVote) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := tc.ReturnPost

	st := post.NewMockPostRepo(ctrl)
	switch tc.VoteMethod {
	case UPVOTE:
		st.EXPECT().Upvote(p.ID, p.Author.ID).Return(p, tc.ReturnError)
	case DOWNVOTE:
		st.EXPECT().Downvote(p.ID, p.Author.ID).Return(p, tc.ReturnError)
	case UNVOTE:
		st.EXPECT().Unvote(p.ID, p.Author.ID).Return(p, tc.ReturnError)
	}

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User:  p.Author,
	}

	req := httptest.NewRequest("POST", "/", nil)
	vars := map[string]string{
		"POST_ID": p.ID,
	}
	req = mux.SetURLVars(req, vars)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	switch tc.VoteMethod {
	case UPVOTE:
		service.Upvote(w, req)
	case DOWNVOTE:
		service.Downvote(w, req)
	case UNVOTE:
		service.Unvote(w, req)
	}

	resp := w.Result()

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	writedPost := post.Post{}
	err = json.Unmarshal(body, &writedPost)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(p, writedPost) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPost,
		)
	}
}

func TestUpvote(t *testing.T) {
	cases := []TestCaseVote{
		{
			ReturnPost:  Posts[0],
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			VoteMethod:  UPVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
			VoteMethod:  UPVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
			VoteMethod:  UPVOTE,
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckVote(t, tc)
		})
	}
}

func TestUpvoteSessionError(t *testing.T) {
	CheckSessionError(t, UPVOTE)
}

func TestDownvote(t *testing.T) {
	cases := []TestCaseVote{
		{
			ReturnPost:  Posts[0],
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			VoteMethod:  DOWNVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
			VoteMethod:  DOWNVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
			VoteMethod:  DOWNVOTE,
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckVote(t, tc)
		})
	}
}

func TestDownvoteSessionError(t *testing.T) {
	CheckSessionError(t, DOWNVOTE)
}

func TestUnvote(t *testing.T) {
	cases := []TestCaseVote{
		{
			ReturnPost:  Posts[0],
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			VoteMethod:  UNVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
			VoteMethod:  UNVOTE,
		},
		{
			ReturnPost:  post.Post{},
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
			VoteMethod:  UNVOTE,
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckVote(t, tc)
		})
	}
}

func TestUnvoteSessionError(t *testing.T) {
	CheckSessionError(t, UNVOTE)
}

type TestCaseDelete struct {
	PostID      string
	UserID      string
	ReturnError error
	StatusCode  int
	CaseName    string
}

func CheckDelete(t *testing.T, tc TestCaseDelete) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().Delete(tc.PostID, tc.UserID).Return(tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	sess := session.Session{
		Token: "dskdkldkl",
		User: user.User{
			ID:       tc.UserID,
			Username: "username",
		},
	}

	req := httptest.NewRequest("DELETE", "/", nil)
	vars := map[string]string{
		"POST_ID": tc.PostID,
	}
	req = mux.SetURLVars(req, vars)
	ctx := session.ContextWithSession(req.Context(), sess)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	service.Delete(w, req)

	resp := w.Result()

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}
}

func TestDelete(t *testing.T) {
	cases := []TestCaseDelete{
		{
			PostID:      "1",
			UserID:      "1",
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
		},
		{
			PostID:      "1",
			UserID:      "1",
			ReturnError: fmt.Errorf("some error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "some error",
		},
		{
			PostID:      "1",
			UserID:      "1",
			ReturnError: post.ErrNoPost,
			StatusCode:  http.StatusNotFound,
			CaseName:    "no post err",
		},
		{
			PostID:      "1",
			UserID:      "1",
			ReturnError: post.ErrNoAccess,
			StatusCode:  http.StatusForbidden,
			CaseName:    "no access",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckDelete(t, tc)
		})
	}
}

func TestDeleteSessionError(t *testing.T) {
	CheckSessionError(t, "Delete")
}

type TestCaseGetUserPosts struct {
	ReturnPosts []post.Post
	ReturnError error
	StatusCode  int
	CaseName    string
	Username    string
}

func CheckGetUserPosts(t *testing.T, tc TestCaseGetUserPosts) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := post.NewMockPostRepo(ctrl)
	st.EXPECT().GetUserPosts(tc.Username).Return(tc.ReturnPosts, tc.ReturnError)

	service := PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: st,
	}

	req := httptest.NewRequest("GET", "/", nil)

	vars := map[string]string{
		"USER_LOGIN": tc.Username,
	}
	req = mux.SetURLVars(req, vars)

	w := httptest.NewRecorder()

	service.GetUserPosts(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf(
			"bad response code, expected %d, got %d",
			tc.StatusCode,
			w.Code,
		)
		return
	}

	if tc.ReturnError != nil {
		return
	}

	writedPosts := make([]post.Post, 0)
	err = json.Unmarshal(body, &writedPosts)
	if err != nil {
		t.Errorf("cant unmarshall %v", body)
		return
	}
	if !reflect.DeepEqual(tc.ReturnPosts, writedPosts) {
		t.Errorf(
			"writed wrong posts, exected: %v\n, got %v",
			Posts,
			writedPosts,
		)
	}
}

func TestGetUserPosts(t *testing.T) {
	cases := []TestCaseGetUserPosts{
		{
			ReturnPosts: Posts,
			ReturnError: nil,
			StatusCode:  http.StatusOK,
			CaseName:    "normal",
			Username:    "username1",
		},
		{
			ReturnPosts: make([]post.Post, 0),
			ReturnError: fmt.Errorf("Some Error"),
			StatusCode:  http.StatusInternalServerError,
			CaseName:    "error",
			Username:    "some",
		},
	}
	for _, tc := range cases {
		t.Run(tc.CaseName, func(t *testing.T) {
			CheckGetUserPosts(t, tc)
		})
	}
}
