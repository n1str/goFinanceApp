package handlers

import (
	"FinanceSystem/pkg/application/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PredictionHandler обрабатывает запросы связанные с прогнозами и аналитикой
type PredictionHandler struct {
	predictionService services.PredictionService
	loanService      services.LoanService
}

// NewPredictionHandler создает новый обработчик прогнозов
func NewPredictionHandler(predictionService services.PredictionService, loanService services.LoanService) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
		loanService:      loanService,
	}
}

// PredictBalance прогнозирует баланс клиента на указанное количество дней
func (h *PredictionHandler) PredictBalance(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("client_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат идентификатора клиента"})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат количества дней"})
		return
	}

	// Ограничение прогноза до 365 дней
	if days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "количество дней должно быть положительным числом"})
		return
	}
	if days > 365 {
		days = 365
	}

	predictions, err := h.loanService.PredictBalance(uint(clientID), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":    clientID,
		"period_days":  days,
		"predictions":  predictions,
	})
}

// GetClientDebtRatio возвращает коэффициент долговой нагрузки клиента
func (h *PredictionHandler) GetClientDebtRatio(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("client_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат идентификатора клиента"})
		return
	}

	ratio, err := h.loanService.GetClientDebtRatio(uint(clientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в проценты и округляем до двух знаков
	ratioPercent := ratio * 100

	c.JSON(http.StatusOK, gin.H{
		"client_id":     clientID,
		"debt_ratio":    ratio,
		"debt_ratio_percent": ratioPercent,
		"status":        getDebtRatioStatus(ratio),
	})
}

// getDebtRatioStatus определяет статус долговой нагрузки
func getDebtRatioStatus(ratio float64) string {
	switch {
	case ratio <= 0.2:
		return "Низкая долговая нагрузка"
	case ratio <= 0.4:
		return "Средняя долговая нагрузка"
	case ratio <= 0.6:
		return "Высокая долговая нагрузка"
	default:
		return "Критическая долговая нагрузка"
	}
}
