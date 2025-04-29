package handlers

import (
	"FinanceSystem/pkg/application/services"
	"net/http"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

// ExternalDataHandler представляет обработчик запросов для внешних данных
type ExternalDataHandler struct {
	externalService services.ExternalService
}

// NewExternalDataHandler создает новый обработчик внешних данных
func NewExternalDataHandler(externalService services.ExternalService) *ExternalDataHandler {
	return &ExternalDataHandler{
		externalService: externalService,
	}
}

// GetCurrentKeyRate обрабатывает запрос на получение текущей ключевой ставки
func (h *ExternalDataHandler) GetCurrentKeyRate(c *gin.Context) {
	// Получаем текущую ключевую ставку
	keyRate, err := h.externalService.GetCurrentKeyRate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, keyRate)
}

// GetKeyRateHistory обрабатывает запрос на получение истории ключевой ставки
func (h *ExternalDataHandler) GetKeyRateHistory(c *gin.Context) {
	// Парсим JSON запрос
	type HistoryRequest struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}
	
	var request HistoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Если даты не указаны, используем значения по умолчанию
	if request.StartDate == "" {
		request.StartDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	}
	if request.EndDate == "" {
		request.EndDate = time.Now().Format("2006-01-02")
	}

	// Преобразуем строки в даты
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат начальной даты"})
		return
	}

	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат конечной даты"})
		return
	}

	// Проверяем, что начальная дата меньше конечной
	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Начальная дата должна быть раньше конечной"})
		return
	}

	// Получаем историю ключевой ставки
	history, err := h.externalService.GetKeyRateHistory(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetCurrencyRate обрабатывает запрос на получение курса валюты
func (h *ExternalDataHandler) GetCurrencyRate(c *gin.Context) {
	// Получаем код валюты из параметров запроса
	currencyCode := c.Param("code")
	if currencyCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан код валюты"})
		return
	}

	// Получаем курс валюты
	rate, err := h.externalService.GetCurrencyRate(currencyCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currency": currencyCode,
		"rate":     rate,
		"date":     time.Now().Format("2006-01-02"),
	})
}

// RegisterCachedRoutes регистрирует маршруты с кэшированием
func (h *ExternalDataHandler) RegisterCachedRoutes(router *gin.RouterGroup) {
	// Создаем кэш на 1 час
	store := persistence.NewInMemoryStore(time.Hour)

	// Маршрут для получения ключевой ставки с кэшированием на 1 час
	router.GET("/current", cache.CachePage(store, time.Hour, h.GetCurrentKeyRate))
	
	// Маршрут для получения истории ключевой ставки с кэшированием на 1 день
	router.POST("/history", cache.CachePage(store, 24*time.Hour, h.GetKeyRateHistory))
	
	// Маршрут для получения курса валюты с кэшированием на 1 час
	router.GET("/currency/:code", cache.CachePage(store, time.Hour, h.GetCurrencyRate))
}
