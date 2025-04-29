package models

import (
	"time"
)

// FinancialAccount представляет собой финансовый счет клиента (бывший Account)
type FinancialAccount struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	ClientID     int       `json:"clientId"`
	Title        string    `json:"title"`
	Funds        float64   `json:"funds"`
	OpenedDate   time.Time `json:"openedDate"`
	
	PaymentCards []PaymentCard `gorm:"foreignKey:FinancialAccountID" json:"paymentCards"`
}
