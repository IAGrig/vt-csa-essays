package store

import "github.com/IAGrig/vt-csa-essays/internal/essay"

type EssayStore interface {
	Add(essay essay.EssayRequest) (essay.Essay, error)
	GetAllEssays() ([]essay.Essay, error)
	GetByAuthorName(username string) (essay.Essay, error)
	RemoveByAuthorName(username string) (essay.Essay, error)
}
