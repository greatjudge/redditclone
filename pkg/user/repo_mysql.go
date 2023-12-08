package user

import (
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type UserMysqlRepository struct {
	DB     *sql.DB
	Logger *zap.SugaredLogger
}

func NewMysqlRepo(db *sql.DB, logger *zap.SugaredLogger) *UserMysqlRepository {
	return &UserMysqlRepository{
		DB:     db,
		Logger: logger,
	}
}

func MD5hashInt(val int64) string {
	idStr := strconv.FormatInt(val, 10)
	hash := md5.Sum([]byte(idStr))
	return fmt.Sprintf("%x", hash)
}

func (repo *UserMysqlRepository) Register(username, password string) (User, error) {
	result, err := repo.DB.Exec(
		`INSERT INTO users (username, password) VALUES (?, MD5(?))`,
		username,
		password,
	)
	var mysqlErr *mysql.MySQLError
	switch {
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1062:
		return User{}, ErrAlreadyExists
	case err != nil:
		repo.Logger.Error("in Register: ", err)
		return User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		repo.Logger.Error("In Register result.LastInsertId(): ", err)
		return User{}, err
	}

	user := User{
		ID:       MD5hashInt(id),
		Username: username,
	}
	return user, nil
}

func (repo *UserMysqlRepository) Authorize(username, password string) (User, error) {
	user := &User{}
	err := repo.DB.
		QueryRow(
			"SELECT MD5(id), username FROM users WHERE username = ? AND password = MD5(?)",
			username,
			password,
		).
		Scan(&user.ID, &user.Username)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return User{}, ErrBadUserPass
	case err != nil:
		repo.Logger.Error("in Authorize: ", err)
		return User{}, err
	}
	return *user, nil
}

func (repo *UserMysqlRepository) GetByID(userID string) (User, error) {
	user := &User{}
	err := repo.DB.
		QueryRow("SELECT MD5(id), username FROM users WHERE MD5(id) = ?", userID).
		Scan(&user.ID, &user.Username)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return User{}, ErrNoUser
	case err != nil:
		repo.Logger.Error("in GetByID: ", err)
		return User{}, err
	}
	return *user, nil
}
