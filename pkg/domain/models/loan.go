package models

import (
	"time"
)

// LoanStatus определяет статус займа
type LoanStatus string

const (
	LoanStatusActive    LoanStatus = "active"    // Активный заем
	LoanStatusCompleted LoanStatus = "completed" // Погашен
	LoanStatusDelayed   LoanStatus = "delayed"   // Просрочен
	LoanStatusRevoked   LoanStatus = "revoked"   // Отменен
)

// PaymentStatus определяет статус платежа по кредиту
type PaymentStatus string

const (
	PaymentStatusScheduled PaymentStatus = "scheduled" // Запланирован
	PaymentStatusCompleted PaymentStatus = "completed" // Оплачен
	PaymentStatusDelayed   PaymentStatus = "delayed"   // Просрочен
)

// Loan представляет кредитный договор (бывший Credit)
type Loan struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	ClientID       uint       `json:"clientId"`
	AccountID      uint       `json:"accountId"`
	Principal      float64    `json:"principal"`
	InterestRate   float64    `json:"interestRate"`
	DurationMonths int        `json:"durationMonths"`
	MonthlyPayment float64    `json:"monthlyPayment"`
	Status         LoanStatus `json:"status" gorm:"type:varchar(20)"`
	IssueDate      time.Time  `json:"issueDate"`
	MaturityDate   time.Time  `json:"maturityDate"`
	Purpose        string     `json:"purpose"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	Client  Client           `gorm:"foreignKey:ClientID" json:"client"`
	Account FinancialAccount `gorm:"foreignKey:AccountID" json:"account"`
}

// PaymentPlan представляет график платежей по кредиту (бывший PaymentSchedule)
type PaymentPlan struct {
	ID              uint          `json:"id" gorm:"primaryKey"`
	LoanID          uint          `json:"loanId"`
	InstallmentNum  int           `json:"installmentNum"`
	DueDate         time.Time     `json:"dueDate"`
	Total           float64       `json:"total"`
	InterestPortion float64       `json:"interestPortion"`
	PrincipalPortion float64      `json:"principalPortion"`
	TotalPayment    float64       `json:"totalPayment"`
	Status          PaymentStatus `json:"status" gorm:"type:varchar(20)"`
	PaidDate        *time.Time    `json:"paidDate,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`

	Loan Loan `gorm:"foreignKey:LoanID" json:"loan"`
}
