package storage

import (
	"FinanceSystem/pkg/domain/models"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// ClientStorage определяет интерфейс для работы с хранилищем клиентов
type ClientStorage interface {
	CreateClient(client *models.Client) error
	GetClientByID(id uint) (*models.Client, error)
	GetClientByLoginName(loginName string) (*models.Client, error)
	GetClientByContact(contact string) (*models.Client, error)
	GetAllClients() ([]models.Client, error)
	UpdateClient(client *models.Client) error
	DeleteClient(id uint) error
	GetClientWithGrants(id uint) (*models.Client, error)
}

// ClientStorageImpl реализует функциональность хранилища клиентов
type ClientStorageImpl struct {
	*BaseRepository
}

// NewClientStorage создаёт новое хранилище клиентов
func NewClientStorage(db *gorm.DB) ClientStorage {
	return &ClientStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateClient создаёт нового клиента
func (r *ClientStorageImpl) CreateClient(client *models.Client) error {
	return r.db.Create(client).Error
}

// GetClientByID находит клиента по ID
func (r *ClientStorageImpl) GetClientByID(id uint) (*models.Client, error) {
	var client models.Client
	err := r.db.First(&client, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("клиент с ID %d не найден", id)
		}
		return nil, err
	}
	return &client, nil
}

// GetClientByLoginName находит клиента по имени входа
func (r *ClientStorageImpl) GetClientByLoginName(loginName string) (*models.Client, error) {
	var client models.Client
	err := r.db.Where("login_name = ?", loginName).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("клиент с именем входа %s не найден", loginName)
		}
		return nil, err
	}
	return &client, nil
}

// GetClientByContact находит клиента по контактной информации
func (r *ClientStorageImpl) GetClientByContact(contact string) (*models.Client, error) {
	var client models.Client
	err := r.db.Where("contact = ?", contact).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("клиент с контактом %s не найден", contact)
		}
		return nil, err
	}
	return &client, nil
}

// GetAllClients возвращает всех клиентов
func (r *ClientStorageImpl) GetAllClients() ([]models.Client, error) {
	var clients []models.Client
	err := r.db.Find(&clients).Error
	return clients, err
}

// UpdateClient обновляет данные клиента
func (r *ClientStorageImpl) UpdateClient(client *models.Client) error {
	return r.db.Save(client).Error
}

// DeleteClient удаляет клиента
func (r *ClientStorageImpl) DeleteClient(id uint) error {
	return r.db.Delete(&models.Client{}, id).Error
}

// GetClientWithGrants возвращает клиента с его разрешениями
func (r *ClientStorageImpl) GetClientWithGrants(id uint) (*models.Client, error) {
	var client models.Client
	err := r.db.Preload("Grants").First(&client, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("клиент с ID %d не найден", id)
		}
		return nil, err
	}
	return &client, nil
}
