package handlers

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/presentation/dto"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PaymentCardHandler представляет обработчик запросов для платежных карт
type PaymentCardHandler struct {
	cardService    services.PaymentCardService
	accountService services.FinancialAccountService
}

// NewPaymentCardHandler создает новый обработчик платежных карт
func NewPaymentCardHandler(
	cardService services.PaymentCardService,
	accountService services.FinancialAccountService,
) *PaymentCardHandler {
	return &PaymentCardHandler{
		cardService:    cardService,
		accountService: accountService,
	}
}

// CreateCard обрабатывает запрос на создание новой платежной карты
func (h *PaymentCardHandler) CreateCard(c *gin.Context) {
	var request dto.CreateCardRequest

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

	// Создаем новую карту
	card, cvv, err := h.cardService.CreatePaymentCard(request.AccountID, request.CardholderName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем информацию о карте вместе с CVV
	c.JSON(http.StatusCreated, gin.H{
		"message": "Карта успешно создана",
		"card": gin.H{
			"id":                 card.ID,
			"cardNumber":         card.CardNumber,
			"cardholderName":     card.CardholderName,
			"expirationDate":     card.ExpirationDate.Format("01/06"), // MM/YY
			"cvv":                cvv, // Обязательно включаем CVV в ответ
			"status":             card.Status,
			"financialAccountID": card.FinancialAccountID,
		},
	})
}

// GetCardsByAccountID обрабатывает запрос на получение всех карт привязанных к счету
func (h *PaymentCardHandler) GetCardsByAccountID(c *gin.Context) {
	// Получаем ID счета из пути
	accountIDStr := c.Param("accountId")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID счета"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет этому клиенту
	account, err := h.accountService.GetFinancialAccountByID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Счет не найден"})
		return
	}

	if uint(account.ClientID) != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	// Получаем все карты для этого счета
	cards, err := h.cardService.GetPaymentCardsByAccountID(uint(accountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Скрываем CVV в ответе
	var response []gin.H
	for _, card := range cards {
		response = append(response, gin.H{
			"id":                 card.ID,
			"cardNumber":         card.CardNumber,
			"cardholderName":     card.CardholderName,
			"expirationDate":     card.ExpirationDate.Format("01/06"), // MM/YY
			"status":             card.Status,
			"financialAccountID": card.FinancialAccountID,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetCardByID обрабатывает запрос на получение карты по ID
func (h *PaymentCardHandler) GetCardByID(c *gin.Context) {
	// Получаем ID карты из пути
	cardIDStr := c.Param("id")
	cardID, err := strconv.ParseUint(cardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID карты"})
		return
	}

	// Получаем карту
	card, err := h.cardService.GetPaymentCardByID(int(cardID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Карта не найдена"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет карты этому клиенту
	account, err := h.accountService.GetFinancialAccountByID(uint(card.FinancialAccountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Счет карты не найден"})
		return
	}

	if uint(account.ClientID) != clientID.(uint) {
		// Проверяем, является ли пользователь администратором
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			return
		}
	}

	// Скрываем CVV в ответе
	c.JSON(http.StatusOK, gin.H{
		"id":                 card.ID,
		"cardNumber":         card.CardNumber,
		"cardholderName":     card.CardholderName,
		"expirationDate":     card.ExpirationDate.Format("01/06"), // MM/YY
		"status":             card.Status,
		"financialAccountID": card.FinancialAccountID,
	})
}

// GetClientCards обрабатывает запрос на получение всех карт клиента
func (h *PaymentCardHandler) GetClientCards(c *gin.Context) {
	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Получаем все карты клиента
	cards, err := h.cardService.GetPaymentCardsByClientID(clientID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Скрываем CVV в ответе
	var response []gin.H
	for _, card := range cards {
		response = append(response, gin.H{
			"id":                 card.ID,
			"cardNumber":         card.CardNumber,
			"cardholderName":     card.CardholderName,
			"expirationDate":     card.ExpirationDate.Format("01/06"), // MM/YY
			"status":             card.Status,
			"financialAccountID": card.FinancialAccountID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cards": response,
		"total": len(response),
	})
}

// BlockCard обрабатывает запрос на блокировку карты
func (h *PaymentCardHandler) BlockCard(c *gin.Context) {
	// Получаем ID карты из пути
	cardIDStr := c.Param("id")
	cardID, err := strconv.ParseUint(cardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID карты"})
		return
	}

	// Получаем карту
	card, err := h.cardService.GetPaymentCardByID(int(cardID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Карта не найдена"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет карты этому клиенту
	account, err := h.accountService.GetFinancialAccountByID(uint(card.FinancialAccountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Счет карты не найден"})
		return
	}

	if uint(account.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Блокируем карту
	if err := h.cardService.BlockPaymentCard(int(cardID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Карта успешно заблокирована"})
}

// UnblockCard обрабатывает запрос на разблокировку карты
func (h *PaymentCardHandler) UnblockCard(c *gin.Context) {
	// Получаем ID карты из пути
	cardIDStr := c.Param("id")
	cardID, err := strconv.ParseUint(cardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID карты"})
		return
	}

	// Получаем карту
	card, err := h.cardService.GetPaymentCardByID(int(cardID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Карта не найдена"})
		return
	}

	// Получаем ID клиента из контекста
	clientID, exists := c.Get("clientID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	// Проверяем, принадлежит ли счет карты этому клиенту
	account, err := h.accountService.GetFinancialAccountByID(uint(card.FinancialAccountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Счет карты не найден"})
		return
	}

	if uint(account.ClientID) != clientID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	// Разблокируем карту
	if err := h.cardService.UnblockPaymentCard(int(cardID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Карта успешно разблокирована"})
}

// ValidateCard обрабатывает запрос на валидацию карты
func (h *PaymentCardHandler) ValidateCard(c *gin.Context) {
	var request dto.ValidateCardRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Парсинг даты истечения срока
	expirationDate, err := time.Parse("01/06", request.ExpirationDate) // MM/YY
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты истечения срока"})
		return
	}

	// Валидация карты
	valid, err := h.cardService.ValidatePaymentCard(request.CardNumber, request.CVV, expirationDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   valid,
		"message": "Карта прошла валидацию",
	})
}
