package handlers

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/presentation/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler представляет обработчик запросов аутентификации
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обрабатывает запрос на регистрацию нового клиента
func (h *AuthHandler) Register(c *gin.Context) {
	var registerRequest dto.RegisterRequest

	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверка данных
	if registerRequest.FullName == "" || registerRequest.LoginName == "" || 
	   registerRequest.Contact == "" || registerRequest.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Все поля обязательны для заполнения"})
		return
	}

	// Регистрация клиента
	err := h.authService.RegisterClient(
		registerRequest.FullName,
		registerRequest.LoginName,
		registerRequest.Contact,
		registerRequest.Password,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Регистрация успешно завершена"})
}

// Login обрабатывает запрос на аутентификацию клиента
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest dto.LoginRequest

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Проверка данных
	if loginRequest.LoginName == "" || loginRequest.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Имя пользователя и пароль обязательны"})
		return
	}

	// Аутентификация
	token, err := h.authService.Authenticate(loginRequest.LoginName, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Получение информации о клиенте
	client, err := h.authService.GetClientInfo(loginRequest.LoginName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении данных клиента"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Вход выполнен успешно",
		"token":   token,
		"client": gin.H{
			"id":        client.ID,
			"fullName":  client.FullName,
			"loginName": client.LoginName,
			"contact":   client.Contact,
		},
	})
}

// CheckAuthStatus проверяет статус аутентификации клиента
func (h *AuthHandler) CheckAuthStatus(c *gin.Context) {
	// Получаем ID клиента из контекста (установлено middleware)
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	// Получаем имя входа из контекста
	loginName, exists := c.Get("loginName")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	// Получаем информацию о клиенте
	client, err := h.authService.GetClientInfo(loginName.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении данных клиента"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"client": gin.H{
			"id":        clientID,
			"fullName":  client.FullName,
			"loginName": client.LoginName,
			"contact":   client.Contact,
		},
	})
}

// GetUserProfile возвращает профиль текущего клиента
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	// Получаем имя входа из контекста
	loginName, exists := c.Get("loginName")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем информацию о клиенте
	client, err := h.authService.GetClientInfo(loginName.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении данных клиента"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        client.ID,
		"fullName":  client.FullName,
		"loginName": client.LoginName,
		"contact":   client.Contact,
	})
}
