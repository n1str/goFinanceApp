package dto

// CreateCardRequest представляет запрос на создание новой платежной карты
type CreateCardRequest struct {
	AccountID      uint   `json:"accountId" binding:"required"`
	CardholderName string `json:"cardholderName" binding:"required"`
}

// CardResponse представляет ответ с информацией о платежной карте
type CardResponse struct {
	ID                 int    `json:"id"`
	CardNumber         string `json:"cardNumber"`
	CardholderName     string `json:"cardholderName"`
	ExpirationDate     string `json:"expirationDate"` // формат "MM/YY"
	Status             string `json:"status"`
	FinancialAccountID int    `json:"financialAccountId"`
}

// CardsListResponse представляет ответ со списком платежных карт
type CardsListResponse struct {
	Cards []CardResponse `json:"cards"`
	Total int            `json:"total"`
}

// ValidateCardRequest представляет запрос на валидацию карты
type ValidateCardRequest struct {
	CardNumber     string `json:"cardNumber" binding:"required"`
	CVV            string `json:"cvv" binding:"required"`
	ExpirationDate string `json:"expirationDate" binding:"required"` // формат "MM/YY"
}
