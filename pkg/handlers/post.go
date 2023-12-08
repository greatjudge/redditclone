package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/greatjudge/redditclone/pkg/comment"
	"github.com/greatjudge/redditclone/pkg/post"
	"github.com/greatjudge/redditclone/pkg/sending"
	"github.com/greatjudge/redditclone/pkg/session"
	"go.uber.org/zap"
)

func JSONMarshalAndSend(w http.ResponseWriter, obj any) {
	serialized, err := json.Marshal(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(serialized)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type PostHandler struct {
	Logger   *zap.SugaredLogger
	PostRepo post.PostRepo
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	elems, err := h.PostRepo.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JSONMarshalAndSend(w, elems)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["CATEGORY_NAME"]
	elems, err := h.PostRepo.GetByCategory(category)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Logger.Infof("get posts by category %v", category)
	JSONMarshalAndSend(w, elems)
}

func handlePostRepoErrors(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, post.ErrNoPost):
		sending.SendJSONMessage(w, "invalid post id", http.StatusNotFound)
	case errors.Is(err, post.ErrNoAccess):
		sending.SendJSONMessage(w, "no access", http.StatusForbidden)
	case errors.Is(err, comment.ErrNoComment):
		sending.SendJSONMessage(w, "invalid comment id", http.StatusNotFound)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, err := h.PostRepo.GetByID(vars["POST_ID"])
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("get post %v", p.ID)
	JSONMarshalAndSend(w, p)
}

func (h *PostHandler) Add(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p := post.Post{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	post.InitPost(&p, sess.User)

	p, err = h.PostRepo.Add(p)
	if err != nil {
		h.Logger.Error("fail add post: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Logger.Infof("add post %v", p.ID)
	w.WriteHeader(http.StatusCreated)
	JSONMarshalAndSend(w, p)
}

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	commForm := &comment.CommentForm{}
	err = json.Unmarshal(body, commForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comm := comment.Comment{
		Author: sess.User,
		Body:   commForm.Comment,
	}

	post, err := h.PostRepo.AddComment(vars["POST_ID"], comm)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("add comment to post %v by %v", post.ID, sess.User.ID)
	w.WriteHeader(http.StatusCreated)
	JSONMarshalAndSend(w, post)
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	post, err := h.PostRepo.DeleteComment(vars["POST_ID"], vars["COMMENT_ID"], sess.User.ID)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("add comment from post %v by %v", post.ID, sess.User.ID)
	JSONMarshalAndSend(w, post)
}

func (h *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	post, err := h.PostRepo.Upvote(vars["POST_ID"], sess.User.ID)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("upvote post %v by %v", post.ID, sess.User.ID)
	JSONMarshalAndSend(w, post)
}

func (h *PostHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	post, err := h.PostRepo.Downvote(vars["POST_ID"], sess.User.ID)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("downvote post %v by %v", post.ID, sess.User.ID)
	JSONMarshalAndSend(w, post)
}

func (h *PostHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	post, err := h.PostRepo.Unvote(vars["POST_ID"], sess.User.ID)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("unvote post %v by %v", post.ID, sess.User.ID)
	JSONMarshalAndSend(w, post)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	vars := mux.Vars(r)
	err = h.PostRepo.Delete(vars["POST_ID"], sess.User.ID)
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("delete post %v by %v", vars["POST_ID"], sess.User.ID)
	sending.SendJSONMessage(w, "success", http.StatusOK)
}

func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	elems, err := h.PostRepo.GetUserPosts(vars["USER_LOGIN"])
	if err != nil {
		handlePostRepoErrors(w, err)
		return
	}
	h.Logger.Infof("get posts created by %v", vars["USER_LOGIN"])
	JSONMarshalAndSend(w, elems)
}
