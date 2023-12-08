package user

import (
	"strconv"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type UserMemoryRepository struct {
	username2User map[string]*User
	mu            *sync.RWMutex
	lastID        int
}

func NewMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		username2User: make(map[string]*User),
		mu:            &sync.RWMutex{},
		lastID:        0,
	}
}

func (repo *UserMemoryRepository) Register(username, password string) (User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.username2User[username]
	if ok {
		return User{}, ErrAlreadyExists
	}

	passhashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return User{}, err
	}

	repo.lastID += 1
	user := NewUser(strconv.Itoa(repo.lastID), username, string(passhashed))
	repo.username2User[username] = &user
	return user, nil
}

func (repo *UserMemoryRepository) Authorize(username, password string) (User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	user, ok := repo.username2User[username]
	if !ok {
		return User{}, ErrNoUser
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
	if err != nil {
		return User{}, ErrBadPass
	}
	return *user, nil
}

func (repo *UserMemoryRepository) GetByID(userID string) (*User, error) {
	for _, user := range repo.username2User {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, ErrNoUser
}
