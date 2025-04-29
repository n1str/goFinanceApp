package models

import (
	"time"
)

// PeriodReport представляет аналитический отчет о финансах за период
type PeriodReport struct {
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	Income       float64   `json:"income"`
	Expense      float64   `json:"expense"`
	NetChange    float64   `json:"netChange"`
	LoanPayments float64   `json:"loanPayments"`
}

// AccountSummary представляет общую информацию о состоянии счета
type AccountSummary struct {
	TotalBalance      float64 `json:"totalBalance"`
	TotalAccounts     int     `json:"totalAccounts"`
	TotalCards        int     `json:"totalCards"`
	ActiveLoans       int     `json:"activeLoans"`
	TotalLoanBalance  float64 `json:"totalLoanBalance"`
	MonthlyLoanPayment float64 `json:"monthlyLoanPayment"`
}

// SpendingCategory представляет категорию трат для аналитики
type SpendingCategory struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

// SpendingAnalytics представляет анализ трат по категориям
type SpendingAnalytics struct {
	Period     string            `json:"period"`
	Categories []SpendingCategory `json:"categories"`
	Total      float64           `json:"total"`
}

// ForecastPoint представляет прогнозную точку для финансового прогноза
type ForecastPoint struct {
	Date   time.Time `json:"date"`
	Value  float64   `json:"value"`
	IsReal bool      `json:"isReal"`
}

// FinancialForecast представляет финансовый прогноз
type FinancialForecast struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	DataPoints  []ForecastPoint `json:"dataPoints"`
	Confidence  float64         `json:"confidence"`
}

// BalancePrediction представляет прогноз баланса на конкретную дату
type BalancePrediction struct {
	Date                time.Time  `json:"date"`
	InitialBalance      float64    `json:"initialBalance"`
	PredictedBalance    float64    `json:"predictedBalance"`
	PlannedIncome       float64    `json:"plannedIncome"`
	PlannedExpenses     float64    `json:"plannedExpenses"`
	LoanPayments        float64    `json:"loanPayments"`
	IsPaymentDate       bool       `json:"isPaymentDate"`
	PaymentDetails      []PaymentDetail `json:"paymentDetails,omitempty"`
}

// PaymentDetail представляет детали планового платежа
type PaymentDetail struct {
	LoanID      uint      `json:"loanId"`
	DueDate     time.Time `json:"dueDate"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
}
