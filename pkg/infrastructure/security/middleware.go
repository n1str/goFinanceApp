package security

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware создает middleware для проверки JWT токена
func AuthMiddleware(clientStorage storage.ClientStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка запроса
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует токен авторизации"})
			c.Abort()
			return
		}

		// Обычно токен передается в формате "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			return
		}

		// Проверяем токен
		claims, err := ValidateToken(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен: " + err.Error()})
			c.Abort()
			return
		}

		// Проверяем существование клиента
		client, err := clientStorage.GetClientByID(claims.ClientID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Клиент не найден"})
			c.Abort()
			return
		}

		// Сохраняем данные клиента в контексте для использования в обработчиках
		c.Set("clientID", claims.ClientID)
		c.Set("loginName", claims.Login)
		c.Set("client", client)

		c.Next()
	}
}

// AdminMiddleware создает middleware для проверки административных прав
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем клиента из контекста
		clientInterface, exists := c.Get("client")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
			c.Abort()
			return
		}

		// Проверяем права администратора
		client, ok := clientInterface.(*models.Client)
		if !ok || !client.IsAdministrator() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Требуются права администратора"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggerMiddleware создает middleware для логирования запросов
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Время начала обработки запроса
		startTime := time.Now()

		// Продолжаем выполнение запроса
		c.Next()

		// После обработки запроса
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// Форматируем и записываем лог
		logrus.WithFields(logrus.Fields{
			"статус":        statusCode,
			"длительность":  latencyTime,
			"IP":            clientIP,
			"метод":         reqMethod,
			"URI":           reqUri,
			"время_запроса": startTime.Format("2006-01-02 15:04:05"),
		}).Info("HTTP запрос")
	}
}

// RateLimitMiddleware создает middleware для ограничения частоты запросов
func RateLimitMiddleware(limit int, duration time.Duration) gin.HandlerFunc {
	// Карта для хранения времени последнего запроса для каждого IP
	lastRequestTime := make(map[string]time.Time)
	// Карта для хранения количества запросов за период для каждого IP
	requestCount := make(map[string]int)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Проверяем, не превышен ли лимит запросов
		if count, exists := requestCount[clientIP]; exists {
			// Если прошло больше времени чем duration, сбрасываем счетчик
			if lastTime, ok := lastRequestTime[clientIP]; ok && now.Sub(lastTime) > duration {
				requestCount[clientIP] = 1
				lastRequestTime[clientIP] = now
			} else if count >= limit {
				// Если лимит превышен, отклоняем запрос
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": fmt.Sprintf("Превышен лимит запросов. Пожалуйста, подождите %v", duration),
				})
				c.Abort()
				return
			} else {
				// Увеличиваем счетчик запросов
				requestCount[clientIP]++
				lastRequestTime[clientIP] = now
			}
		} else {
			// Первый запрос от этого IP
			requestCount[clientIP] = 1
			lastRequestTime[clientIP] = now
		}

		c.Next()
	}
}

// CorsMiddleware создает middleware для обработки CORS
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
