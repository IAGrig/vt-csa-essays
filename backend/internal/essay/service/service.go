package service

import (
	"github.com/IAGrig/vt-csa-essays/internal/essay"
	"github.com/IAGrig/vt-csa-essays/internal/essay/store"
)

type UserSevice interface {
	Add(essay.EssayRequest) (essay.Essay, error)
	GetByAuthorName(string) (essay.Essay, error)
	RemoveByAuthorName(string) (essay.Essay, error)
}

type service struct {
	store store.EssayStore
}

func New(store store.EssayStore) UserSevice {
	return &service{store}
}

func (service *service) Add(request essay.EssayRequest) (essay.Essay, error) {
	return service.store.Add(request)
}

func (service *service) GetByAuthorName(authorname string) (essay.Essay, error) {
	return service.store.GetByAuthorName(authorname)
}

func (service *service) RemoveByAuthorName(authorname string) (essay.Essay, error) {
	return service.store.RemoveByAuthorName(authorname)
}
