package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"errors"
	"fmt"
)

// AccessService представляет интерфейс сервиса управления правами доступа
type AccessService interface {
	CreateAccess(name, description string) (*models.Access, error)
	GetAccessByID(id uint) (*models.Access, error)
	GetAccessByName(name string) (*models.Access, error)
	GetAllAccess() ([]models.Access, error)
	AssignAccessToClient(clientID uint, accessID uint) error
	RevokeAccessFromClient(clientID uint, accessID uint) error
	GetClientAccess(clientID uint) ([]models.Access, error)
	InitializeDefaultAccess() error
}

// AccessServiceImpl реализует функциональность сервиса управления правами доступа
type AccessServiceImpl struct {
	accessStorage storage.AccessStorage
}

// NewAccessService создаёт новый сервис управления правами доступа
func NewAccessService(accessStorage storage.AccessStorage) AccessService {
	return &AccessServiceImpl{
		accessStorage: accessStorage,
	}
}

// CreateAccess создаёт новое право доступа
func (s *AccessServiceImpl) CreateAccess(name, description string) (*models.Access, error) {
	// Проверяем, существует ли уже такое право доступа
	existingAccess, _ := s.accessStorage.GetAccessByName(name)
	if existingAccess != nil {
		return nil, errors.New("право доступа с таким именем уже существует")
	}

	access := &models.Access{
		Name:        name,
		Description: description,
	}

	if err := s.accessStorage.CreateAccess(access); err != nil {
		return nil, fmt.Errorf("ошибка при создании права доступа: %v", err)
	}

	return access, nil
}

// GetAccessByID получает право доступа по ID
func (s *AccessServiceImpl) GetAccessByID(id uint) (*models.Access, error) {
	return s.accessStorage.GetAccessByID(id)
}

// GetAccessByName получает право доступа по имени
func (s *AccessServiceImpl) GetAccessByName(name string) (*models.Access, error) {
	return s.accessStorage.GetAccessByName(name)
}

// GetAllAccess получает все права доступа
func (s *AccessServiceImpl) GetAllAccess() ([]models.Access, error) {
	return s.accessStorage.GetAllAccess()
}

// AssignAccessToClient назначает право доступа клиенту
func (s *AccessServiceImpl) AssignAccessToClient(clientID uint, accessID uint) error {
	return s.accessStorage.AssignAccessToClient(clientID, accessID)
}

// RevokeAccessFromClient отзывает право доступа у клиента
func (s *AccessServiceImpl) RevokeAccessFromClient(clientID uint, accessID uint) error {
	return s.accessStorage.RevokeAccessFromClient(clientID, accessID)
}

// GetClientAccess получает все права доступа клиента
func (s *AccessServiceImpl) GetClientAccess(clientID uint) ([]models.Access, error) {
	return s.accessStorage.GetClientAccess(clientID)
}

// InitializeDefaultAccess инициализирует стандартные права доступа
func (s *AccessServiceImpl) InitializeDefaultAccess() error {
	// Проверяем, существуют ли уже права доступа
	accesses, err := s.accessStorage.GetAllAccess()
	if err != nil {
		return fmt.Errorf("ошибка при проверке существующих прав доступа: %v", err)
	}

	if len(accesses) > 0 {
		// Права доступа уже существуют, ничего не делаем
		return nil
	}

	// Создаем стандартные права доступа
	defaultAccesses := []struct {
		Name        string
		Description string
	}{
		{models.AccessUser, "Обычный пользователь системы"},
		{models.AccessAdmin, "Администратор системы с полным доступом"},
	}

	for _, access := range defaultAccesses {
		_, err := s.CreateAccess(access.Name, access.Description)
		if err != nil {
			return fmt.Errorf("ошибка при создании права доступа %s: %v", access.Name, err)
		}
	}

	return nil
}
