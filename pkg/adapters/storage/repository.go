package storage

import (
	"gorm.io/gorm"
)

// Repository представляет базовый интерфейс для всех репозиториев
type Repository interface {
	GetDB() *gorm.DB
}

// BaseRepository представляет базовую реализацию для всех репозиториев
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository создает новый базовый репозиторий
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetDB возвращает экземпляр соединения с базой данных
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}
