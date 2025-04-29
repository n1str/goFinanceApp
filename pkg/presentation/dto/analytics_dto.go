package dto

// PeriodReportRequest представляет запрос на получение отчета за период
type PeriodReportRequest struct {
	StartDate string `json:"startDate" binding:"required"` // формат "YYYY-MM-DD"
	EndDate   string `json:"endDate" binding:"required"`   // формат "YYYY-MM-DD"
}

// CategoryResponse представляет ответ с информацией о категории трат
type CategoryResponse struct {
	Name       string  `json:"name"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

// SpendingAnalyticsResponse представляет ответ с анализом трат
type SpendingAnalyticsResponse struct {
	Period     string             `json:"period"`
	Categories []CategoryResponse `json:"categories"`
	Total      float64            `json:"total"`
}

// ForecastPointResponse представляет точку финансового прогноза
type ForecastPointResponse struct {
	Date   string  `json:"date"`
	Value  float64 `json:"value"`
	IsReal bool    `json:"isReal"`
}

// FinancialForecastResponse представляет ответ с финансовым прогнозом
type FinancialForecastResponse struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	DataPoints  []ForecastPointResponse `json:"dataPoints"`
	Confidence  float64                 `json:"confidence"`
}

// AccountSummaryResponse представляет ответ с общей информацией о счетах
type AccountSummaryResponse struct {
	TotalBalance      float64 `json:"totalBalance"`
	TotalAccounts     int     `json:"totalAccounts"`
	TotalCards        int     `json:"totalCards"`
	ActiveLoans       int     `json:"activeLoans"`
	TotalLoanBalance  float64 `json:"totalLoanBalance"`
	MonthlyLoanPayment float64 `json:"monthlyLoanPayment"`
}
