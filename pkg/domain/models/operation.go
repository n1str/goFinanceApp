package models

import (
	"time"
)

// OperationType определяет тип финансовой операции
type OperationType string

const (
	OperationDeposit   OperationType = "deposit"   // Пополнение
	OperationWithdraw  OperationType = "withdraw"  // Снятие
	OperationTransfer  OperationType = "transfer"  // Перевод
	OperationLoanIssue OperationType = "loanIssue" // Оформление кредита
)

// Operation представляет финансовую операцию между счетами (бывшая Transaction)
type Operation struct {
	ID              int           `json:"id" gorm:"primaryKey"`
	OperationType   OperationType `json:"operationType" gorm:"type:varchar(20)"`
	SourceAccountID int           `json:"sourceAccountId,omitempty"`
	TargetAccountID int           `json:"targetAccountId"`
	Sum             float64       `json:"sum"`
	Details         string        `json:"details"`
	ExecutedAt      time.Time     `json:"executedAt"`
	Result          string        `json:"result" gorm:"type:varchar(20)"`

	SourceAccount FinancialAccount `gorm:"foreignKey:SourceAccountID" json:"sourceAccount,omitempty"`
	TargetAccount FinancialAccount `gorm:"foreignKey:TargetAccountID" json:"targetAccount"`
}
