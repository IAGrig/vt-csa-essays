package db

import "github.com/IAGrig/vt-csa-essays/internal/models"

type EssayStore interface {
	Add(essay models.EssayRequest) (models.Essay, error)
	GetByAuthorName(username string) (models.Essay, error)
	RemoveByAuthorName(username string) (models.Essay, error)
}
