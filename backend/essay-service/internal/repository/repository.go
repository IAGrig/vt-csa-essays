package repository

import (
	"errors"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
)

var (
	DuplicateErr     = errors.New("essay already exists")
	EssayNotFoundErr = errors.New("essay not found")
)

type EssayRepository interface {
	Add(essay models.EssayRequest) (models.Essay, error)
	GetAllEssays() ([]models.Essay, error)
	GetByAuthorName(username string) (models.Essay, error)
	RemoveByAuthorName(username string) (models.Essay, error)
	SearchByContent(string) ([]models.Essay, error)
}
