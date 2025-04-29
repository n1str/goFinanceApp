package storage

import (
	"FinanceSystem/pkg/domain/models"
	"gorm.io/gorm"
)

// AccessStorage определяет интерфейс для работы с хранилищем прав доступа
type AccessStorage interface {
	CreateAccess(access *models.Access) error
	GetAccessByID(id uint) (*models.Access, error)
	GetAccessByName(name string) (*models.Access, error)
	GetAllAccess() ([]models.Access, error)
	AssignAccessToClient(clientID uint, accessID uint) error
	RevokeAccessFromClient(clientID uint, accessID uint) error
	GetClientAccess(clientID uint) ([]models.Access, error)
}

// AccessStorageImpl реализует функциональность хранилища прав доступа
type AccessStorageImpl struct {
	*BaseRepository
}

// NewAccessStorage создаёт новое хранилище прав доступа
func NewAccessStorage(db *gorm.DB) AccessStorage {
	return &AccessStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateAccess создаёт новое право доступа
func (r *AccessStorageImpl) CreateAccess(access *models.Access) error {
	return r.db.Create(access).Error
}

// GetAccessByID находит право доступа по ID
func (r *AccessStorageImpl) GetAccessByID(id uint) (*models.Access, error) {
	var access models.Access
	err := r.db.First(&access, id).Error
	if err != nil {
		return nil, err
	}
	return &access, nil
}

// GetAccessByName находит право доступа по имени
func (r *AccessStorageImpl) GetAccessByName(name string) (*models.Access, error) {
	var access models.Access
	err := r.db.Where("name = ?", name).First(&access).Error
	if err != nil {
		return nil, err
	}
	return &access, nil
}

// GetAllAccess возвращает все права доступа
func (r *AccessStorageImpl) GetAllAccess() ([]models.Access, error) {
	var accesses []models.Access
	err := r.db.Find(&accesses).Error
	return accesses, err
}

// AssignAccessToClient назначает право доступа клиенту
func (r *AccessStorageImpl) AssignAccessToClient(clientID uint, accessID uint) error {
	return r.db.Exec("INSERT INTO client_access (client_id, access_id) VALUES (?, ?)", clientID, accessID).Error
}

// RevokeAccessFromClient отзывает право доступа у клиента
func (r *AccessStorageImpl) RevokeAccessFromClient(clientID uint, accessID uint) error {
	return r.db.Exec("DELETE FROM client_access WHERE client_id = ? AND access_id = ?", clientID, accessID).Error
}

// GetClientAccess возвращает все права доступа клиента
func (r *AccessStorageImpl) GetClientAccess(clientID uint) ([]models.Access, error) {
	var accesses []models.Access
	err := r.db.Raw(`
		SELECT a.* FROM accesses a 
		JOIN client_access ca ON a.id = ca.access_id 
		WHERE ca.client_id = ?
	`, clientID).Scan(&accesses).Error
	return accesses, err
}
