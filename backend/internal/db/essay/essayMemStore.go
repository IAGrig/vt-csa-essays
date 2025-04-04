package essaystore

import (
	"errors"
	"time"

	"github.com/IAGrig/vt-csa-essays/internal/models"
)

var (
	DublicateErr     = errors.New("essay already exists")
	EssayNotFoundErr = errors.New("essay not found")
)

// !Must be used only for test purposes!
type EssayMemStore struct {
	list map[string]models.Essay
}

// Creates an instance of EssayMemStore, !must be used only for test purposes!
func NewEssayMemStore() *EssayMemStore {
	list := make(map[string]models.Essay)

	return &EssayMemStore{
		list,
	}
}

func (store *EssayMemStore) Add(request models.EssayRequest) (models.Essay, error) {
	if _, err := store.GetByAuthorName(request.Author); err != EssayNotFoundErr {
		return models.Essay{}, DublicateErr
	}

	essay := models.Essay{
		ID:        len(store.list) + 1,
		Content:   request.Content,
		Author:    request.Author,
		CreatedAt: time.Now(),
	}

	store.list[request.Author] = essay
	return essay, nil
}

func (store *EssayMemStore) GetByAuthorName(username string) (models.Essay, error) {
	if essay, ok := store.list[username]; ok {
		return essay, nil
	}
	return models.Essay{}, EssayNotFoundErr
}

func (store *EssayMemStore) RemoveByAuthorName(username string) (models.Essay, error) {
	essay, err := store.GetByAuthorName(username)
	if err != nil {
		return models.Essay{}, DublicateErr
	}

	delete(store.list, essay.Author)

	return essay, nil
}
