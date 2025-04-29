package storage

import (
	"FinanceSystem/pkg/domain/models"
	"fmt"
	"gorm.io/gorm"
)

// PaymentCardStorage определяет интерфейс для работы с хранилищем платежных карт
type PaymentCardStorage interface {
	CreatePaymentCard(card *models.PaymentCard) error
	GetPaymentCardByID(id uint) (*models.PaymentCard, error)
	GetPaymentCardByNumber(number string) (*models.PaymentCard, error)
	GetPaymentCardByHMAC(hmac string) (*models.PaymentCard, error)
	GetPaymentCardsByAccountID(accountID uint) ([]models.PaymentCard, error)
	GetPaymentCardsByClientID(clientID uint) ([]models.PaymentCard, error)
	UpdatePaymentCard(card *models.PaymentCard) error
	DeactivatePaymentCard(id uint) error
	GetAllPaymentCards() ([]models.PaymentCard, error)
}

// PaymentCardStorageImpl реализует функциональность хранилища платежных карт
type PaymentCardStorageImpl struct {
	*BaseRepository
}

// NewPaymentCardStorage создаёт новое хранилище платежных карт
func NewPaymentCardStorage(db *gorm.DB) PaymentCardStorage {
	return &PaymentCardStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreatePaymentCard создаёт новую платежную карту
func (r *PaymentCardStorageImpl) CreatePaymentCard(card *models.PaymentCard) error {
	return r.db.Create(card).Error
}

// GetPaymentCardByID находит платежную карту по ID
func (r *PaymentCardStorageImpl) GetPaymentCardByID(id uint) (*models.PaymentCard, error) {
	var card models.PaymentCard
	err := r.db.First(&card, id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// GetPaymentCardByNumber находит платежную карту по номеру
func (r *PaymentCardStorageImpl) GetPaymentCardByNumber(number string) (*models.PaymentCard, error) {
	var card models.PaymentCard
	err := r.db.Where("card_number = ?", number).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// GetPaymentCardByHMAC находит платежную карту по HMAC номера карты
func (r *PaymentCardStorageImpl) GetPaymentCardByHMAC(hmac string) (*models.PaymentCard, error) {
	var card models.PaymentCard
	err := r.db.Where("card_number_hmac = ?", hmac).First(&card).Error
	if err != nil {
		return nil, fmt.Errorf("карта с указанным HMAC не найдена: %v", err)
	}
	return &card, nil
}

// GetPaymentCardsByAccountID находит все платежные карты по ID счета
func (r *PaymentCardStorageImpl) GetPaymentCardsByAccountID(accountID uint) ([]models.PaymentCard, error) {
	var cards []models.PaymentCard
	err := r.db.Where("financial_account_id = ?", accountID).Find(&cards).Error
	return cards, err
}

// GetPaymentCardsByClientID получает все платежные карты клиента
func (r *PaymentCardStorageImpl) GetPaymentCardsByClientID(clientID uint) ([]models.PaymentCard, error) {
	var cards []models.PaymentCard
	
	// Находим все счета клиента
	var accounts []models.FinancialAccount
	if err := r.db.Where("client_id = ?", clientID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	
	// Если счетов нет, возвращаем пустой список
	if len(accounts) == 0 {
		return []models.PaymentCard{}, nil
	}
	
	// Собираем ID всех счетов клиента
	var accountIDs []uint
	for _, account := range accounts {
		accountIDs = append(accountIDs, uint(account.ID))
	}
	
	// Находим все карты, привязанные к этим счетам
	if err := r.db.Where("financial_account_id IN ?", accountIDs).Find(&cards).Error; err != nil {
		return nil, err
	}
	
	return cards, nil
}

// UpdatePaymentCard обновляет данные платежной карты
func (r *PaymentCardStorageImpl) UpdatePaymentCard(card *models.PaymentCard) error {
	return r.db.Save(card).Error
}

// DeactivatePaymentCard деактивирует платежную карту
func (r *PaymentCardStorageImpl) DeactivatePaymentCard(id uint) error {
	return r.db.Model(&models.PaymentCard{}).Where("id = ?", id).Update("status", "inactive").Error
}

// GetAllPaymentCards возвращает все платежные карты
func (r *PaymentCardStorageImpl) GetAllPaymentCards() ([]models.PaymentCard, error) {
	var cards []models.PaymentCard
	err := r.db.Find(&cards).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении всех карт: %v", err)
	}
	return cards, nil
}
