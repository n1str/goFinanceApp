package security

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT ключи и настройки
const (
	// Секретный ключ для подписи JWT токенов (в реальном приложении должен храниться безопасно)
	JWTSecretKey = "fM7NVJqxErP3GYzH5tLW9FdZ2cRbTgKj"
	
	// Время жизни токена - 24 часа
	JWTExpirationTime = 24 * time.Hour
)

// TokenClaims представляет собой структуру данных для JWT токена
type TokenClaims struct {
	ClientID uint   `json:"clientId"`
	Login    string `json:"login"`
	jwt.StandardClaims
}

// GenerateToken создает новый JWT токен для клиента
func GenerateToken(clientID uint, login string) (string, error) {
	expirationTime := time.Now().Add(JWTExpirationTime)
	claims := &TokenClaims{
		ClientID: clientID,
		Login:    login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken проверяет JWT токен и возвращает данные из него
func ValidateToken(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем использование подходящего алгоритма подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неподдерживаемый метод подписи токена")
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("недействительный токен")
	}

	return claims, nil
}

// RefreshToken обновляет JWT токен, если он скоро истечет
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Если токен истекает менее чем через 12 часов, обновляем его
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 12*time.Hour {
		return GenerateToken(claims.ClientID, claims.Login)
	}

	// Иначе возвращаем текущий токен
	return tokenString, nil
}
