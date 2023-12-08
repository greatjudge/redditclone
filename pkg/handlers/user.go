package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/greatjudge/redditclone/pkg/sending"
	"github.com/greatjudge/redditclone/pkg/session"
	"github.com/greatjudge/redditclone/pkg/user"

	"go.uber.org/zap"
)

var (
	ErrValidation  = errors.New("validation error")
	ErrContentType = errors.New("bad content type")
)

type UserHandler struct {
	Logger   *zap.SugaredLogger
	UserRepo user.UserRepo
	Sessions session.SessionsManager
}

type LoginForm struct {
	Username string `json:"username" valid:"required,matches(^[a-zA-Z0-9_]+$)"`
	Password string `json:"password" valid:"required,length(8|255)"`
}

type LoginAnswer struct {
	Token string `json:"token"`
}

func Validate(lf *LoginForm) []string {
	_, err := govalidator.ValidateStruct(lf)
	valErrs := make([]string, 0)
	if allErrs, ok := err.(govalidator.Errors); ok {
		for _, fld := range allErrs.Errors() {
			valErrs = append(valErrs, fld.Error())
		}
	}
	return valErrs
}

func handleValidationErrors(validationErrs []string, w http.ResponseWriter) error {
	errMes, err := json.Marshal(validationErrs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("fail to marshall validationErrs: %w", err)
	}
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write(errMes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("fail to write errMes: %w", err)
	}
	return ErrValidation
}

func handleUserErrors(err error, w http.ResponseWriter) {
	switch {
	case errors.Is(err, user.ErrNoUser):
		sending.SendJSONMessage(w, "user not found", http.StatusNotFound)
	case errors.Is(err, user.ErrBadUserPass):
		sending.SendJSONMessage(w, "bad username or password", http.StatusBadRequest)
	case errors.Is(err, user.ErrBadPass):
		sending.SendJSONMessage(w, "invalid password", http.StatusBadRequest)
	case errors.Is(err, user.ErrAlreadyExists):
		sending.SendJSONMessage(w, "user already exists", http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func LoginFormFromBody(w http.ResponseWriter, r *http.Request) (LoginForm, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return LoginForm{}, fmt.Errorf("fail to ReadAll r.Body: %w", err)
	}

	lf := &LoginForm{}
	err = json.Unmarshal(body, lf)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return LoginForm{}, fmt.Errorf("fail to unmarshall loginForm : %w", err)
	}

	validationErrs := Validate(lf)
	if len(validationErrs) != 0 {
		err := handleValidationErrors(validationErrs, w)
		return LoginForm{}, fmt.Errorf("validation errors: %w", err)
	}
	return *lf, nil
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	lf, err := LoginFormFromBody(w, r)
	if err != nil {
		h.Logger.Errorf("Login: error from LoginFormFromBody: %v", err.Error())
		return
	}

	u, err := h.UserRepo.Authorize(lf.Username, lf.Password)
	if err != nil {
		handleUserErrors(err, w)
		return
	}

	token, err := h.Sessions.Create(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Logger.Infof("created session for %v", u.ID)
	sending.JSONMarshalAndSend(w, LoginAnswer{Token: token})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	lf, err := LoginFormFromBody(w, r)
	if err != nil {
		h.Logger.Errorf("Register: error from LoginFormFromBody: %v", err.Error())
		return
	}

	u, err := h.UserRepo.Register(lf.Username, lf.Password)
	if err != nil {
		handleUserErrors(err, w)
		return
	}

	token, err := h.Sessions.Create(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Logger.Infof("user id=%v, username=%v registered", u.ID, u.Username)
	w.WriteHeader(http.StatusCreated)
	sending.JSONMarshalAndSend(w, LoginAnswer{Token: token})
}
