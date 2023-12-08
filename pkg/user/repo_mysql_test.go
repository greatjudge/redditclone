package user

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	elemID := MD5hashInt(1)

	rows := sqlmock.NewRows([]string{"id", "username"})
	expect := []User{
		{elemID, "username", ""},
	}
	for _, user := range expect {
		rows = rows.AddRow(elemID, user.Username)
	}
	query := `SELECT MD5\(id\), username FROM users WHERE MD5\(id\) = ?`

	mock.
		ExpectQuery(query).
		WithArgs(elemID).
		WillReturnRows(rows)

	repo := &UserMysqlRepository{
		DB:     db,
		Logger: zap.NewNop().Sugar(),
	}

	item, err := repo.GetByID(elemID)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	mock.
		ExpectQuery(query).
		WithArgs(elemID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(elemID)
	if err != ErrNoUser {
		t.Errorf("expected %v, got %v", ErrNoUser.Error(), err.Error())
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id"}).
		AddRow(elemID)

	mock.
		ExpectQuery(query).
		WithArgs(elemID).
		WillReturnRows(rows)

	_, err = repo.GetByID(elemID)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	username, password := "username", "password"
	query := `SELECT MD5\(id\), username FROM users WHERE username = \? AND password = MD5\(\?\)`
	id := MD5hashInt(1)
	rows := sqlmock.NewRows([]string{"id", "username"})
	expect := []User{
		{id, username, ""},
	}
	for _, user := range expect {
		rows = rows.AddRow(id, user.Username)
	}

	mock.
		ExpectQuery(query).
		WithArgs(username, password).
		WillReturnRows(rows)

	repo := &UserMysqlRepository{
		DB:     db,
		Logger: zap.NewNop().Sugar(),
	}

	item, err := repo.Authorize(username, password)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	mock.
		ExpectQuery(query).
		WithArgs(username, password).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Authorize(username, password)
	if err != ErrBadUserPass {
		t.Errorf("expected %v, got %v", ErrNoUser.Error(), err.Error())
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id"}).
		AddRow(id)

	mock.
		ExpectQuery(query).
		WithArgs(username, password).
		WillReturnRows(rows)

	_, err = repo.Authorize(username, password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewMysqlRepo(db, zap.NewNop().Sugar())

	username, password := "username", "password"
	query := `INSERT INTO users \(username, password\) VALUES \(\?, MD5\(\?\)\)`

	// ok query
	mock.
		ExpectExec(query).
		WithArgs(username, password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.Register(username, password)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	exptected := User{
		ID:       MD5hashInt(1),
		Username: username,
		password: "",
	}
	if user != exptected {
		t.Errorf("bad user: want %v, have %v", exptected, user)
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query ErrAlreadyExists
	er := &mysql.MySQLError{
		Number: 1062,
	}
	mock.
		ExpectExec(query).
		WithArgs(username, password).
		WillReturnError(er)

	_, err = repo.Register(username, password)
	if err != ErrAlreadyExists {
		t.Errorf("expected %v, got %v", ErrAlreadyExists.Error(), err.Error())
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// query error
	mock.
		ExpectExec(query).
		WithArgs(username, password).
		WillReturnError(fmt.Errorf("bad query"))

	_, err = repo.Register(username, password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// result error
	mock.
		ExpectExec(query).
		WithArgs(username, password).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("bad_result")))

	_, err = repo.Register(username, password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
