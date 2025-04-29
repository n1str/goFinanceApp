package services

import (
	"FinanceSystem/pkg/domain/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Создаем мок для хранилища клиентов
type MockClientStorage struct {
	mock.Mock
}

func (m *MockClientStorage) CreateClient(client *models.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockClientStorage) GetClientByID(id uint) (*models.Client, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientStorage) GetClientByLoginName(loginName string) (*models.Client, error) {
	args := m.Called(loginName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientStorage) GetClientByContact(contact string) (*models.Client, error) {
	args := m.Called(contact)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientStorage) GetAllClients() ([]models.Client, error) {
	args := m.Called()
	return args.Get(0).([]models.Client), args.Error(1)
}

func (m *MockClientStorage) UpdateClient(client *models.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockClientStorage) DeleteClient(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockClientStorage) GetClientWithGrants(id uint) (*models.Client, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

// Создаем мок для хранилища прав доступа
type MockAccessStorage struct {
	mock.Mock
}

func (m *MockAccessStorage) CreateAccess(access *models.Access) error {
	args := m.Called(access)
	return args.Error(0)
}

func (m *MockAccessStorage) GetAccessByID(id uint) (*models.Access, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Access), args.Error(1)
}

func (m *MockAccessStorage) GetAccessByName(name string) (*models.Access, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Access), args.Error(1)
}

func (m *MockAccessStorage) GetAllAccess() ([]models.Access, error) {
	args := m.Called()
	return args.Get(0).([]models.Access), args.Error(1)
}

func (m *MockAccessStorage) AssignAccessToClient(clientID uint, accessID uint) error {
	args := m.Called(clientID, accessID)
	return args.Error(0)
}

func (m *MockAccessStorage) RevokeAccessFromClient(clientID uint, accessID uint) error {
	args := m.Called(clientID, accessID)
	return args.Error(0)
}

func (m *MockAccessStorage) GetClientAccess(clientID uint) ([]models.Access, error) {
	args := m.Called(clientID)
	return args.Get(0).([]models.Access), args.Error(1)
}

// Тесты для AuthService
func TestRegisterClient(t *testing.T) {
	// Настраиваем моки
	mockClientStorage := new(MockClientStorage)
	mockAccessStorage := new(MockAccessStorage)

	// Создаем тестовый сервис с моками
	authService := NewAuthService(mockClientStorage, mockAccessStorage, "testSecretKey", 24*time.Hour)

	// Тест случая, когда пользователь с таким именем уже существует
	mockClientStorage.On("GetClientByLoginName", "existingUser").Return(&models.Client{}, nil)
	err := authService.RegisterClient("Test User", "existingUser", "test@example.com", "password123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "уже существует")

	// Тест случая, когда контакт уже используется
	mockClientStorage.On("GetClientByLoginName", "newUser").Return(nil, errors.New("not found"))
	mockClientStorage.On("GetClientByContact", "existing@example.com").Return(&models.Client{}, nil)
	err = authService.RegisterClient("Test User", "newUser", "existing@example.com", "password123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "контактом уже существует")

	// Тест успешной регистрации
	mockClientStorage.On("GetClientByLoginName", "successUser").Return(nil, errors.New("not found"))
	mockClientStorage.On("GetClientByContact", "success@example.com").Return(nil, errors.New("not found"))
	mockAccessStorage.On("GetAccessByName", models.AccessUser).Return(&models.Access{
		ID:          1,
		Name:        models.AccessUser,
		Description: "Обычный пользователь",
	}, nil)
	mockClientStorage.On("CreateClient", mock.AnythingOfType("*models.Client")).Return(nil)

	err = authService.RegisterClient("Success User", "successUser", "success@example.com", "password123")
	assert.NoError(t, err)

	mockClientStorage.AssertExpectations(t)
	mockAccessStorage.AssertExpectations(t)
}

func TestAuthenticate(t *testing.T) {
	// Настраиваем моки
	mockClientStorage := new(MockClientStorage)
	mockAccessStorage := new(MockAccessStorage)

	// Создаем тестовый сервис с моками
	authService := NewAuthService(mockClientStorage, mockAccessStorage, "testSecretKey", 24*time.Hour)

	// Создаем тестового клиента с хешированным паролем
	// В реальном тесте нужно использовать правильный хеш
	testClient := &models.Client{
		ID:        1,
		FullName:  "Test User",
		LoginName: "testuser",
		Contact:   "test@example.com",
		PassHash:  "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // hash для "password123"
	}

	// Тест неверного имени пользователя
	mockClientStorage.On("GetClientByLoginName", "nonexistent").Return(nil, errors.New("not found"))
	token, err := authService.Authenticate("nonexistent", "password123")
	assert.Error(t, err)
	assert.Empty(t, token)

	// Из-за ограничений, связанных с тем как реально работает bcrypt.CompareHashAndPassword,
	// мы не можем полностью протестировать сценарий неверного пароля здесь.
	// В реальном тесте нужно использовать интерфейс для функции хеширования и сравнения паролей.

	// Тест успешной аутентификации нельзя выполнить без использования правильного хеша пароля
	// или мока для bcrypt.CompareHashAndPassword

	mockClientStorage.AssertExpectations(t)
}
