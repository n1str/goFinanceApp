package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims представляет данные для JWT токена
type JWTClaims struct {
	ClientID uint   `json:"clientId"`
	Login    string `json:"login"`
	jwt.StandardClaims
}

// AuthService представляет интерфейс сервиса аутентификации
type AuthService interface {
	RegisterClient(fullName, loginName, contact, password string) error
	Authenticate(loginName, password string) (string, error)
	VerifyToken(tokenString string) (*JWTClaims, error)
	GetClientInfo(loginName string) (*models.Client, error)
}

// AuthServiceImpl реализует функциональность сервиса аутентификации
type AuthServiceImpl struct {
	clientStorage storage.ClientStorage
	accessStorage storage.AccessStorage
	secretKey     string
	tokenExpiry   time.Duration
}

// NewAuthService создаёт новый сервис аутентификации
func NewAuthService(clientStorage storage.ClientStorage, accessStorage storage.AccessStorage, secretKey string, tokenExpiry time.Duration) AuthService {
	return &AuthServiceImpl{
		clientStorage: clientStorage,
		accessStorage: accessStorage,
		secretKey:     secretKey,
		tokenExpiry:   tokenExpiry,
	}
}

// RegisterClient регистрирует нового клиента
func (s *AuthServiceImpl) RegisterClient(fullName, loginName, contact, password string) error {
	// Проверяем, существует ли пользователь с таким логином
	existingUser, _ := s.clientStorage.GetClientByLoginName(loginName)
	if existingUser != nil {
		return errors.New("клиент с таким именем входа уже существует")
	}

	// Проверяем, существует ли пользователь с таким контактом
	existingContact, _ := s.clientStorage.GetClientByContact(contact)
	if existingContact != nil {
		return errors.New("клиент с таким контактом уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("ошибка при хешировании пароля")
	}

	// Получаем роль пользователя
	userAccess, err := s.accessStorage.GetAccessByName(models.AccessUser)
	if err != nil {
		return errors.New("ошибка при получении прав доступа")
	}

	// Создаем нового клиента
	client := &models.Client{
		FullName:  fullName,
		LoginName: loginName,
		Contact:   contact,
		PassHash:  string(hashedPassword),
		Grants:    []models.Access{*userAccess},
	}

	if err := s.clientStorage.CreateClient(client); err != nil {
		return errors.New("ошибка при создании клиента")
	}

	return nil
}

// Authenticate аутентифицирует клиента и возвращает JWT токен
func (s *AuthServiceImpl) Authenticate(loginName, password string) (string, error) {
	// Ищем клиента по имени входа
	client, err := s.clientStorage.GetClientByLoginName(loginName)
	if err != nil {
		return "", errors.New("неверное имя входа или пароль")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(client.PassHash), []byte(password)); err != nil {
		return "", errors.New("неверное имя входа или пароль")
	}

	// Загружаем права доступа клиента - проверяем, что клиент существует с правами
	_, err = s.clientStorage.GetClientWithGrants(client.ID)
	if err != nil {
		return "", errors.New("ошибка при загрузке прав доступа")
	}

	// Создаем JWT токен
	expirationTime := time.Now().Add(s.tokenExpiry)
	claims := &JWTClaims{
		ClientID: client.ID,
		Login:    client.LoginName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", errors.New("ошибка при создании токена")
	}

	return tokenString, nil
}

// VerifyToken проверяет JWT токен
func (s *AuthServiceImpl) VerifyToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, errors.New("невалидный токен")
	}

	if !token.Valid {
		return nil, errors.New("токен недействителен")
	}

	return claims, nil
}

// GetClientInfo возвращает информацию о клиенте
func (s *AuthServiceImpl) GetClientInfo(loginName string) (*models.Client, error) {
	return s.clientStorage.GetClientByLoginName(loginName)
}
