package store

import (
	"errors"
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/essay"
)

var (
	DuplicateErr     = errors.New("essay already exists")
	EssayNotFoundErr = errors.New("essay not found")
)

// !Must be used only for test purposes!
type EssayMemStore struct {
	list map[string]essay.Essay
}

// Creates an instance of EssayMemStore, !must be used only for test purposes!
func NewEssayMemStore() *EssayMemStore {
	list := make(map[string]essay.Essay)

	return &EssayMemStore{
		list,
	}
}

func (store *EssayMemStore) Add(request essay.EssayRequest) (essay.Essay, error) {
	if _, err := store.GetByAuthorName(request.Author); err != EssayNotFoundErr {
		return essay.Essay{}, DuplicateErr
	}

	essay := essay.Essay{
		ID:        len(store.list) + 1,
		Content:   request.Content,
		Author:    request.Author,
		CreatedAt: time.Now(),
	}

	store.list[request.Author] = essay
	return essay, nil
}

func (store *EssayMemStore) GetByAuthorName(username string) (essay.Essay, error) {
	if essay, ok := store.list[username]; ok {
		return essay, nil
	}
	return essay.Essay{}, EssayNotFoundErr
}

func (store *EssayMemStore) RemoveByAuthorName(username string) (essay.Essay, error) {
	ess, err := store.GetByAuthorName(username)
	if err != nil {
		return essay.Essay{}, err
	}

	delete(store.list, ess.Author)

	return ess, nil
}
