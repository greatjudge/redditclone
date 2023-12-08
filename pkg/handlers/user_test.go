package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greatjudge/redditclone/pkg/sending"
	"github.com/greatjudge/redditclone/pkg/session"
	"github.com/greatjudge/redditclone/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockReader struct {
	mock.Mock
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func TestLoginFormFromBodyWithBodyError(t *testing.T) {
	mockR := mockReader{}
	mockR.On("Read", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("error reading"))

	r := httptest.NewRequest("GET", "/", &mockR)
	w := httptest.NewRecorder()
	expected := "error reading"
	expectedCode := http.StatusBadRequest
	_, err := LoginFormFromBody(w, r)
	if err == nil || expected != err.Error() {
		t.Errorf(`wrong result: got "%v"\n expected: "%v"`, err, expected)
	}

	if w.Code != expectedCode {
		t.Errorf("wrong StatusCode: got %d, expected %d",
			w.Code, expectedCode)
	}
}

func TestLoginFormUnmarshallError(t *testing.T) {
	reader := strings.NewReader(".s/dw{")
	r := httptest.NewRequest("GET", "http://example", reader)

	w := httptest.NewRecorder()
	expected := "invalid character '.' looking for beginning of value"
	expectedCode := http.StatusBadRequest
	_, err := LoginFormFromBody(w, r)
	if err == nil || expected != err.Error() {
		t.Errorf(`wrong result: got "%v", expected: "%v"`, err, expected)
	}

	if w.Code != expectedCode {
		t.Errorf("wrong StatusCode: got %d, expected %d",
			w.Code, expectedCode)
	}
}

func TestLoginFormValidationError(t *testing.T) {
	lf := LoginForm{
		Username: "usernam+-q1@#",
		Password: "dsdd",
	}
	lfBytes, err := json.Marshal(lf)
	if err != nil {
		t.Errorf("marshall err: %v", err.Error())
	}
	reader := bytes.NewReader(lfBytes)
	r := httptest.NewRequest("GET", "http://example", reader)

	w := httptest.NewRecorder()
	expected := ErrValidation
	expectedCode := http.StatusBadRequest
	_, err = LoginFormFromBody(w, r)
	if err == nil || expected != err {
		t.Errorf(`wrong result: got "%v", expected: "%v"`, err, expected)
	}

	if w.Code != expectedCode {
		t.Errorf("wrong StatusCode: got %d, expected %d",
			w.Code, expectedCode)
	}
}

func TestHandleValidationErrors(t *testing.T) {
	inner := ResponseWriterInner{}
	mockW := mockResponseWriter{inner: &inner}
	mockW.On("Write", mock.AnythingOfType("[]uint8")).Return(0, fmt.Errorf("some error"))
	err := handleValidationErrors([]string{"one", "two"}, &mockW)
	assert.NotNil(t, err)
	if mockW.inner.Code != http.StatusInternalServerError {
		t.Errorf(
			"bad response code, expected %d, got %d",
			http.StatusInternalServerError,
			mockW.inner.Code,
		)
		return
	}
}

const (
	LOGIN    = "Login"
	REGISTER = "Register"
)

func CheckBadLoginFormInLoginRegister(t *testing.T, method string) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := user.NewMockUserRepo(ctrl)
	sessMock := session.NewMockSessionsManager(ctrl)

	password := "1212"
	usr := user.NewUser("id", "l+/-=*", password)

	service := &UserHandler{
		Logger:   zap.NewNop().Sugar(),
		UserRepo: st,
		Sessions: sessMock,
	}

	lf := LoginForm{
		Username: usr.Username,
		Password: password,
	}
	lfBytes, err := json.Marshal(lf)
	if err != nil {
		t.Errorf("marshall err: %v", err.Error())
	}
	reader := bytes.NewReader(lfBytes)

	req := httptest.NewRequest("POST", "/", reader)
	w := httptest.NewRecorder()

	if method == LOGIN {
		service.Login(w, req)
	} else {
		service.Register(w, req)
	}
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code: %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestLoginBadLoginForm(t *testing.T) {
	CheckBadLoginFormInLoginRegister(t, LOGIN)
}

func TestRegisterBadLoginForm(t *testing.T) {
	CheckBadLoginFormInLoginRegister(t, REGISTER)
}

type TestCaseLogin struct {
	StatusCode  int
	ExpectToken bool
	RepoErr     error
	CreateErr   error
	ErrMessage  string
	Method      string // Login or Register
}

func CheckLoginTest(t *testing.T, tc TestCaseLogin) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := user.NewMockUserRepo(ctrl)
	sessMock := session.NewMockSessionsManager(ctrl)

	password := "password"
	usr := user.NewUser("id", "username", password)
	token := "kjkljHKJHKJhkjjk213"

	if tc.Method == LOGIN {
		st.EXPECT().Authorize(usr.Username, password).Return(usr, tc.RepoErr)
	} else {
		st.EXPECT().Register(usr.Username, password).Return(usr, tc.RepoErr)
	}

	if tc.RepoErr == nil {
		sessMock.EXPECT().Create(usr).Return(token, tc.CreateErr)
	}

	service := &UserHandler{
		Logger:   zap.NewNop().Sugar(),
		UserRepo: st,
		Sessions: sessMock,
	}

	lf := LoginForm{
		Username: usr.Username,
		Password: password,
	}
	lfBytes, err := json.Marshal(lf)
	if err != nil {
		t.Errorf("marshall err: %v", err.Error())
	}
	reader := bytes.NewReader(lfBytes)

	req := httptest.NewRequest("POST", "/", reader)
	w := httptest.NewRecorder()

	if tc.Method == LOGIN {
		service.Login(w, req)
	} else {
		service.Register(w, req)
	}

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cant read body: %v", err)
	}

	if tc.ExpectToken {
		la := LoginAnswer{}
		err = json.Unmarshal(body, &la)
		if err != nil {
			t.Errorf("cant unmarshall %v", body)
		}

		if token != la.Token {
			t.Errorf("incorrect token, expect: %v, got: %v", la.Token, token)
		}
	}

	if resp.StatusCode != tc.StatusCode {
		t.Errorf("expected status code: %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	if tc.ErrMessage != "" {
		mes := sending.Message{}
		err = json.Unmarshal(body, &mes)
		if err != nil {
			t.Errorf("cant unmarshall message froms %v", body)
		}

		if mes.Message != tc.ErrMessage {
			t.Errorf(`expected message: "%v", got: "%v"`, tc.ErrMessage, mes.Message)
		}
	}
}

func TestLogin(t *testing.T) {
	cases := []TestCaseLogin{
		{
			StatusCode:  http.StatusOK,
			ExpectToken: true,
			RepoErr:     nil,
			CreateErr:   nil,
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusNotFound,
			ExpectToken: false,
			RepoErr:     user.ErrNoUser,
			CreateErr:   nil,
			ErrMessage:  "user not found",
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrBadUserPass,
			CreateErr:   nil,
			ErrMessage:  "bad username or password",
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrBadPass,
			CreateErr:   nil,
			ErrMessage:  "invalid password",
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrAlreadyExists,
			CreateErr:   nil,
			ErrMessage:  "user already exists",
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusInternalServerError,
			ExpectToken: false,
			RepoErr:     fmt.Errorf("some error"),
			CreateErr:   nil,
			Method:      LOGIN,
		},
		{
			StatusCode:  http.StatusInternalServerError,
			ExpectToken: false,
			RepoErr:     nil,
			CreateErr:   fmt.Errorf("some error"),
			Method:      LOGIN,
		},
	}
	for _, tc := range cases {
		CheckLoginTest(t, tc)
	}
}

func TestRegister(t *testing.T) {
	cases := []TestCaseLogin{
		{
			StatusCode:  http.StatusCreated,
			ExpectToken: true,
			RepoErr:     nil,
			CreateErr:   nil,
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusNotFound,
			ExpectToken: false,
			RepoErr:     user.ErrNoUser,
			CreateErr:   nil,
			ErrMessage:  "user not found",
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrBadUserPass,
			CreateErr:   nil,
			ErrMessage:  "bad username or password",
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrBadPass,
			CreateErr:   nil,
			ErrMessage:  "invalid password",
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusBadRequest,
			ExpectToken: false,
			RepoErr:     user.ErrAlreadyExists,
			CreateErr:   nil,
			ErrMessage:  "user already exists",
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusInternalServerError,
			ExpectToken: false,
			RepoErr:     fmt.Errorf("some error"),
			CreateErr:   nil,
			Method:      REGISTER,
		},
		{
			StatusCode:  http.StatusInternalServerError,
			ExpectToken: false,
			RepoErr:     nil,
			CreateErr:   fmt.Errorf("some error"),
			Method:      REGISTER,
		},
	}
	for _, tc := range cases {
		CheckLoginTest(t, tc)
	}
}
