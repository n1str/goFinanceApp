package models

import (
	"time"
)

// PaymentCard представляет платежную карту (бывшая Card)
type PaymentCard struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	CardNumber         string    `json:"cardNumber" gorm:"unique;not null"`
	CardholderName     string    `json:"cardholderName" gorm:"not null"`
	ExpirationDate     time.Time `json:"expirationDate" gorm:"not null"`
	CVV                string    `json:"-" gorm:"not null"`
	Status             string    `json:"status" gorm:"not null;default:'active'"`
	FinancialAccountID int       `json:"financialAccountId" gorm:"not null"`
}
