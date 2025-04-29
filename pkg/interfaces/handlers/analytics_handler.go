package handlers

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/presentation/dto"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler представляет обработчик запросов для аналитики
type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
}

// NewAnalyticsHandler создает новый обработчик аналитики
func NewAnalyticsHandler(analyticsService services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetClientSummary обрабатывает запрос на получение общей информации о финансах клиента
func (h *AnalyticsHandler) GetClientSummary(c *gin.Context) {
	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем общую информацию
	summary, err := h.analyticsService.GetClientSummary(clientID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetPeriodReport обрабатывает запрос на получение отчета за период
func (h *AnalyticsHandler) GetPeriodReport(c *gin.Context) {
	var request dto.PeriodReportRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
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

	// Получаем отчет за период
	report, err := h.analyticsService.GetPeriodReport(clientID.(uint), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetSpendingAnalytics обрабатывает запрос на получение анализа расходов
func (h *AnalyticsHandler) GetSpendingAnalytics(c *gin.Context) {
	// Получаем период из запроса
	period := c.DefaultQuery("period", "month")

	// Проверяем, что период указан корректно
	validPeriods := map[string]bool{
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
	}

	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный период. Допустимые значения: week, month, quarter, year"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем анализ расходов
	analytics, err := h.analyticsService.GetSpendingAnalytics(clientID.(uint), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetFinancialForecast обрабатывает запрос на получение финансового прогноза
func (h *AnalyticsHandler) GetFinancialForecast(c *gin.Context) {
	// Получаем количество месяцев для прогноза
	monthsStr := c.DefaultQuery("months", "12")
	months, err := strconv.Atoi(monthsStr)
	if err != nil || months <= 0 || months > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверное количество месяцев. Допустимый диапазон: 1-60"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем финансовый прогноз
	forecast, err := h.analyticsService.GenerateFinancialForecast(clientID.(uint), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecast)
}

// GetAccountsAnalytics обрабатывает запрос на получение аналитики счетов клиента
func (h *AnalyticsHandler) GetAccountsAnalytics(c *gin.Context) {
	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем общую информацию о счетах
	summary, err := h.analyticsService.GetClientSummary(clientID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts_summary": summary,
		"date": time.Now().Format("2006-01-02"),
	})
}
