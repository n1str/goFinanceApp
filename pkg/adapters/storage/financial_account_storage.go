package storage

import (
	"FinanceSystem/pkg/domain/models"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// FinancialAccountStorage определяет интерфейс для работы с хранилищем финансовых счетов
type FinancialAccountStorage interface {
	CreateFinancialAccount(account *models.FinancialAccount, clientID uint) error
	GetFinancialAccountByID(id uint) (*models.FinancialAccount, error)
	GetFinancialAccountsByClientID(clientID uint) ([]models.FinancialAccount, error)
	GetAllFinancialAccounts() ([]models.FinancialAccount, error)
	UpdateFinancialAccount(account *models.FinancialAccount) error
	DeleteFinancialAccount(id uint) error
	GetFinancialAccountWithCards(id uint) (*models.FinancialAccount, error)
}

// FinancialAccountStorageImpl реализует функциональность хранилища финансовых счетов
type FinancialAccountStorageImpl struct {
	*BaseRepository
}

// NewFinancialAccountStorage создаёт новое хранилище финансовых счетов
func NewFinancialAccountStorage(db *gorm.DB) FinancialAccountStorage {
	return &FinancialAccountStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateFinancialAccount создаёт новый финансовый счет
func (r *FinancialAccountStorageImpl) CreateFinancialAccount(account *models.FinancialAccount, clientID uint) error {
	account.ClientID = int(clientID)
	return r.db.Create(account).Error
}

// GetFinancialAccountByID находит финансовый счет по ID
func (r *FinancialAccountStorageImpl) GetFinancialAccountByID(id uint) (*models.FinancialAccount, error) {
	var account models.FinancialAccount
	err := r.db.First(&account, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("финансовый счет с ID %d не найден", id)
		}
		return nil, err
	}
	return &account, nil
}

// GetFinancialAccountsByClientID находит все финансовые счета клиента
func (r *FinancialAccountStorageImpl) GetFinancialAccountsByClientID(clientID uint) ([]models.FinancialAccount, error) {
	var accounts []models.FinancialAccount
	// Используем Preload для загрузки связанных карт
	err := r.db.Where("client_id = ?", clientID).Preload("PaymentCards").Find(&accounts).Error
	return accounts, err
}

// GetAllFinancialAccounts возвращает все финансовые счета
func (r *FinancialAccountStorageImpl) GetAllFinancialAccounts() ([]models.FinancialAccount, error) {
	var accounts []models.FinancialAccount
	err := r.db.Find(&accounts).Error
	return accounts, err
}

// UpdateFinancialAccount обновляет данные финансового счета
func (r *FinancialAccountStorageImpl) UpdateFinancialAccount(account *models.FinancialAccount) error {
	return r.db.Save(account).Error
}

// DeleteFinancialAccount удаляет финансовый счет
func (r *FinancialAccountStorageImpl) DeleteFinancialAccount(id uint) error {
	return r.db.Delete(&models.FinancialAccount{}, id).Error
}

// GetFinancialAccountWithCards возвращает финансовый счет с привязанными картами
func (r *FinancialAccountStorageImpl) GetFinancialAccountWithCards(id uint) (*models.FinancialAccount, error) {
	var account models.FinancialAccount
	err := r.db.Preload("PaymentCards").First(&account, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("финансовый счет с ID %d не найден", id)
		}
		return nil, err
	}
	return &account, nil
}
