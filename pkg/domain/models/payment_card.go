package models

import (
	"time"
)

// PaymentCard представляет платежную карту (бывшая Card)
type PaymentCard struct {
	ID                 int       `json:"id" gorm:"primaryKey"`
	CardNumber         string    `json:"cardNumber" gorm:"unique;not null"`
	CardNumberHMAC     string    `json:"-" gorm:"unique;index;not null"`  // HMAC для проверки целостности и поиска
	CardholderName     string    `json:"cardholderName" gorm:"not null"`
	ExpirationDate     time.Time `json:"expirationDate" gorm:"not null"`
	EncryptedExpDate   string    `json:"-" gorm:"not null"`               // PGP-зашифрованная дата в формате MM/YY
	CVV                string    `json:"-" gorm:"not null"`               // Хеш CVV (bcrypt)
	Status             string    `json:"status" gorm:"not null;default:'active'"`
	FinancialAccountID int       `json:"financialAccountId" gorm:"not null"`
}
