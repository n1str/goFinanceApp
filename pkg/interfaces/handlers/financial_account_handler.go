package handlers

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/domain/models"
	"FinanceSystem/pkg/presentation/dto"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// FinancialAccountHandler представляет обработчик запросов для финансовых счетов
type FinancialAccountHandler struct {
	accountService services.FinancialAccountService
}

// NewFinancialAccountHandler создает новый обработчик финансовых счетов
func NewFinancialAccountHandler(accountService services.FinancialAccountService) *FinancialAccountHandler {
	return &FinancialAccountHandler{
		accountService: accountService,
	}
}

// CreateAccount обрабатывает запрос на создание нового финансового счета
func (h *FinancialAccountHandler) CreateAccount(c *gin.Context) {
	var request dto.CreateAccountRequest

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

	// Создаем новый счет
	account := &models.FinancialAccount{
		Title:      request.Title,
		Funds:      request.InitialFunds,
		OpenedDate: time.Now(),
	}

	if err := h.accountService.CreateFinancialAccount(account, clientID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Счет успешно создан",
		"account": account,
	})
}

// GetAllAccounts обрабатывает запрос на получение всех счетов
func (h *FinancialAccountHandler) GetAllAccounts(c *gin.Context) {
	// Проверяем права администратора (должны быть установлены middleware)
	accounts, err := h.accountService.GetAllFinancialAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetMyAccounts обрабатывает запрос на получение счетов текущего клиента
func (h *FinancialAccountHandler) GetMyAccounts(c *gin.Context) {
	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	accounts, err := h.accountService.GetFinancialAccountsByClientID(clientID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccountByID обрабатывает запрос на получение счета по ID
func (h *FinancialAccountHandler) GetAccountByID(c *gin.Context) {
	// Получаем ID счета из пути
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID счета"})
		return
	}

	account, err := h.accountService.GetFinancialAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет текущему клиенту
	if uint(account.ClientID) != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	c.JSON(http.StatusOK, account)
}

// AddFunds обрабатывает запрос на пополнение счета
func (h *FinancialAccountHandler) AddFunds(c *gin.Context) {
	var request dto.FundsOperationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Получаем ID счета из пути
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID счета"})
		return
	}

	// Проверяем права на операцию со счетом
	account, err := h.accountService.GetFinancialAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет текущему клиенту
	if uint(account.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Пополняем счет
	if err := h.accountService.AddFunds(uint(accountID), request.Amount, request.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Средства успешно зачислены"})
}

// WithdrawFunds обрабатывает запрос на снятие средств со счета
func (h *FinancialAccountHandler) WithdrawFunds(c *gin.Context) {
	var request dto.FundsOperationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Получаем ID счета из пути
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID счета"})
		return
	}

	// Проверяем права на операцию со счетом
	account, err := h.accountService.GetFinancialAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет текущему клиенту
	if uint(account.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Снимаем средства со счета
	if err := h.accountService.WithdrawFunds(uint(accountID), request.Amount, request.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Средства успешно сняты"})
}

// TransferFunds обрабатывает запрос на перевод средств между счетами
func (h *FinancialAccountHandler) TransferFunds(c *gin.Context) {
	var request dto.TransferRequest

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

	// Проверяем права на операцию с исходным счетом
	fromAccount, err := h.accountService.GetFinancialAccountByID(request.SourceAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Исходный счет не найден"})
		return
	}

	// Проверяем, принадлежит ли исходный счет текущему клиенту
	if uint(fromAccount.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Проверяем существование целевого счета
	_, err = h.accountService.GetFinancialAccountByID(request.TargetAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Целевой счет не найден"})
		return
	}

	// Выполняем перевод
	if err := h.accountService.TransferFunds(request.SourceAccountID, request.TargetAccountID, request.Amount, request.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Перевод успешно выполнен"})
}

// GetAccountOperations обрабатывает запрос на получение операций по счету
func (h *FinancialAccountHandler) GetAccountOperations(c *gin.Context) {
	// Получаем ID счета из пути
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID счета"})
		return
	}

	// Проверяем права на просмотр операций по счету
	account, err := h.accountService.GetFinancialAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет текущему клиенту
	if uint(account.ClientID) != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	// Получаем операции по счету
	operations, err := h.accountService.GetAccountOperations(uint(accountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, operations)
}
