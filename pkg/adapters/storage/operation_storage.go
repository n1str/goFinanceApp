package storage

import (
	"FinanceSystem/pkg/domain/models"
	"gorm.io/gorm"
	"time"
)

// OperationStorage определяет интерфейс для работы с хранилищем финансовых операций
type OperationStorage interface {
	CreateOperation(operation *models.Operation) error
	GetOperationByID(id uint) (*models.Operation, error)
	GetOperationsByAccountID(accountID uint) ([]models.Operation, error)
	GetOperationsByPeriod(startDate, endDate time.Time) ([]models.Operation, error)
	GetAllOperations() ([]models.Operation, error)
}

// OperationStorageImpl реализует функциональность хранилища финансовых операций
type OperationStorageImpl struct {
	*BaseRepository
}

// NewOperationStorage создаёт новое хранилище финансовых операций
func NewOperationStorage(db *gorm.DB) OperationStorage {
	return &OperationStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateOperation создаёт новую финансовую операцию
func (r *OperationStorageImpl) CreateOperation(operation *models.Operation) error {
	return r.db.Create(operation).Error
}

// GetOperationByID находит финансовую операцию по ID
func (r *OperationStorageImpl) GetOperationByID(id uint) (*models.Operation, error) {
	var operation models.Operation
	err := r.db.First(&operation, id).Error
	if err != nil {
		return nil, err
	}
	return &operation, nil
}

// GetOperationsByAccountID находит все финансовые операции по ID счета
func (r *OperationStorageImpl) GetOperationsByAccountID(accountID uint) ([]models.Operation, error) {
	var operations []models.Operation
	err := r.db.Where("source_account_id = ? OR target_account_id = ?", accountID, accountID).Find(&operations).Error
	return operations, err
}

// GetOperationsByPeriod находит все финансовые операции за указанный период
func (r *OperationStorageImpl) GetOperationsByPeriod(startDate, endDate time.Time) ([]models.Operation, error) {
	var operations []models.Operation
	err := r.db.Where("executed_at BETWEEN ? AND ?", startDate, endDate).Find(&operations).Error
	return operations, err
}

// GetAllOperations возвращает все финансовые операции
func (r *OperationStorageImpl) GetAllOperations() ([]models.Operation, error) {
	var operations []models.Operation
	err := r.db.Find(&operations).Error
	return operations, err
}
