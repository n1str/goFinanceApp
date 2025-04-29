package handlers

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/domain/models"
	"FinanceSystem/pkg/presentation/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LoanHandler представляет обработчик запросов для займов
type LoanHandler struct {
	loanService    services.LoanService
	accountService services.FinancialAccountService
}

// NewLoanHandler создает новый обработчик займов
func NewLoanHandler(
	loanService services.LoanService,
	accountService services.FinancialAccountService,
) *LoanHandler {
	return &LoanHandler{
		loanService:    loanService,
		accountService: accountService,
	}
}

// CreateLoan обрабатывает запрос на создание нового займа
func (h *LoanHandler) CreateLoan(c *gin.Context) {
	var request dto.CreateLoanRequest

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

	// Проверяем, принадлежит ли счет этому клиенту
	account, err := h.accountService.GetFinancialAccountByID(request.AccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Счет не найден"})
		return
	}

	if uint(account.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Создаем новый займ
	loan, err := h.loanService.CreateLoan(
		clientID.(uint),
		request.AccountID,
		request.Amount,
		request.InterestRate,
		request.Term,
		request.Purpose,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Займ успешно оформлен",
		"loan":    loan,
	})
}

// GetLoansByClientID возвращает список займов клиента по ID
func (h *LoanHandler) GetLoansByClientID(c *gin.Context) {
	clientIDStr := c.Param("id")
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID клиента"})
		return
	}

	loans, err := h.loanService.GetClientLoans(uint(clientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// GetClientLoansCompat обрабатывает запрос на получение всех займов клиента
func (h *LoanHandler) GetClientLoansCompat(c *gin.Context) {
	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем все займы клиента
	loans, err := h.loanService.GetClientLoans(clientID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// GetLoanByID обрабатывает запрос на получение займа по ID
func (h *LoanHandler) GetLoanByID(c *gin.Context) {
	// Получаем ID займа из пути
	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseUint(loanIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID займа"})
		return
	}

	// Получаем займ
	loan, err := h.loanService.GetLoanByID(uint(loanID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Займ не найден"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли займ этому клиенту
	if loan.ClientID != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	c.JSON(http.StatusOK, loan)
}

// GetLoanPaymentPlan обрабатывает запрос на получение графика платежей по займу
func (h *LoanHandler) GetLoanPaymentPlan(c *gin.Context) {
	// Получаем ID займа из пути
	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseUint(loanIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID займа"})
		return
	}

	// Получаем займ и график платежей
	loan, paymentPlan, err := h.loanService.GetLoanWithPaymentPlan(uint(loanID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Займ не найден"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли займ этому клиенту
	if loan.ClientID != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	// Возвращаем только план платежей
	c.JSON(http.StatusOK, paymentPlan)
}

// MakePayment обрабатывает запрос на совершение платежа по займу
func (h *LoanHandler) MakePayment(c *gin.Context) {
	var request dto.MakePaymentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Получаем ID займа из пути
	loanIDStr := c.Param("id")
	loanID, err := strconv.ParseUint(loanIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID займа"})
		return
	}

	// Получаем займ
	loan, err := h.loanService.GetLoanByID(uint(loanID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Займ не найден"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли займ этому клиенту
	if loan.ClientID != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Совершаем платеж
	if err := h.loanService.MakePayment(uint(loanID), request.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Платеж успешно выполнен"})
}

// CalculateLoanDetails обрабатывает запрос на расчет параметров займа
func (h *LoanHandler) CalculateLoanDetails(c *gin.Context) {
	var request dto.CalculateLoanRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Рассчитываем ежемесячный платеж
	monthlyPayment := h.loanService.CalculateMonthlyPayment(
		request.Amount,
		12.0, // Используем фиксированную ставку, если она не передана
		request.Term,
	)

	// Рассчитываем общую сумму выплат
	totalPayment := monthlyPayment * float64(request.Term)
	
	// Рассчитываем переплату
	overPayment := totalPayment - request.Amount

	c.JSON(http.StatusOK, gin.H{
		"principal":      request.Amount,
		"interest_rate":  12.0, // Фиксированная ставка
		"duration":       request.Term,
		"monthly_payment": monthlyPayment,
		"total_payment":  totalPayment,
		"over_payment":   overPayment,
		"purpose":        request.Purpose,
	})
}

// UpdateAllLoanPayments обновляет ежемесячные платежи для всех займов
func (h *LoanHandler) UpdateAllLoanPayments(c *gin.Context) {
	// Вызываем метод обновления ежемесячных платежей
	err := h.loanService.UpdateAllLoanMonthlyPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ежемесячные платежи по всем займам успешно обновлены"})
}

// GetAllLoans возвращает список всех займов
func (h *LoanHandler) GetAllLoans(c *gin.Context) {
	loans, err := h.loanService.GetAllLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// GetLoansByStatus возвращает список займов по статусу
func (h *LoanHandler) GetLoansByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан статус займа"})
		return
	}

	// Преобразуем строковый статус в перечисление models.LoanStatus
	var status models.LoanStatus
	switch statusStr {
	case "active":
		status = models.LoanStatusActive
	case "completed":
		status = models.LoanStatusCompleted
	case "delayed":
		status = models.LoanStatusDelayed
	case "revoked":
		status = models.LoanStatusRevoked
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный статус займа"})
		return
	}

	loans, err := h.loanService.GetLoansByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}
