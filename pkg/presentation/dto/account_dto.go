package dto

// CreateAccountRequest представляет запрос на создание нового финансового счета
type CreateAccountRequest struct {
	Title        string  `json:"title" binding:"required"`
	InitialFunds float64 `json:"initialFunds" binding:"min=0"`
}

// FundsOperationRequest представляет запрос на операцию с денежными средствами
type FundsOperationRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Details string  `json:"details"`
}

// TransferRequest представляет запрос на перевод средств между счетами
type TransferRequest struct {
	SourceAccountID uint    `json:"sourceAccountID" binding:"required"`
	TargetAccountID uint    `json:"targetAccountID" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Details        string  `json:"details"`
}

// AccountResponse представляет ответ с информацией о финансовом счете
type AccountResponse struct {
	ID         int     `json:"id"`
	ClientID   int     `json:"clientId"`
	Title      string  `json:"title"`
	Funds      float64 `json:"funds"`
	OpenedDate string  `json:"openedDate"`
	CardCount  int     `json:"cardCount"`
}

// AccountsListResponse представляет ответ со списком финансовых счетов
type AccountsListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
	Total    int               `json:"total"`
}
