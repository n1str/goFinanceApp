package dto

// CreateLoanRequest представляет запрос на создание нового займа
type CreateLoanRequest struct {
	AccountID      uint    `json:"accountId" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Term           int     `json:"term" binding:"required,gt=0"`
	InterestRate   float64 `json:"interestRate,omitempty"`
	Purpose        string  `json:"purpose"`
}

// MakePaymentRequest представляет запрос на совершение платежа по займу
type MakePaymentRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// CalculateLoanRequest представляет запрос на расчет параметров займа
type CalculateLoanRequest struct {
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Term           int     `json:"term" binding:"required,gt=0"`
	Purpose        string  `json:"purpose"`
}

// LoanResponse представляет ответ с информацией о займе
type LoanResponse struct {
	ID             uint    `json:"id"`
	ClientID       uint    `json:"clientId"`
	AccountID      uint    `json:"accountId"`
	Principal      float64 `json:"principal"`
	InterestRate   float64 `json:"interestRate"`
	DurationMonths int     `json:"durationMonths"`
	MonthlyPayment float64 `json:"monthlyPayment"`
	Status         string  `json:"status"`
	IssueDate      string  `json:"issueDate"`
	MaturityDate   string  `json:"maturityDate"`
	Purpose        string  `json:"purpose"`
}

// PaymentPlanResponse представляет ответ с информацией о платеже по графику
type PaymentPlanResponse struct {
	ID              uint    `json:"id"`
	LoanID          uint    `json:"loanId"`
	InstallmentNum  int     `json:"installmentNum"`
	DueDate         string  `json:"dueDate"`
	Total           float64 `json:"total"`
	InterestPortion float64 `json:"interestPortion"`
	PrincipalPortion float64 `json:"principalPortion"`
	TotalPayment    float64 `json:"totalPayment"`
	Status          string  `json:"status"`
	PaidDate        string  `json:"paidDate,omitempty"`
}

// LoanCalculationResponse представляет ответ с расчетом параметров займа
type LoanCalculationResponse struct {
	Principal      float64 `json:"principal"`
	InterestRate   float64 `json:"interestRate"`
	DurationMonths int     `json:"durationMonths"`
	MonthlyPayment float64 `json:"monthlyPayment"`
	TotalPayment   float64 `json:"totalPayment"`
	OverPayment    float64 `json:"overPayment"`
}
