package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

// PaymentCardService представляет интерфейс сервиса платежных карт
type PaymentCardService interface {
	CreatePaymentCard(accountID uint, holderName string) (*models.PaymentCard, string, error)
	GetPaymentCardByID(id int) (*models.PaymentCard, error)
	GetPaymentCardByNumber(cardNumber string) (*models.PaymentCard, error)
	GetPaymentCardsByAccountID(accountID uint) ([]models.PaymentCard, error)
	GetPaymentCardsByClientID(clientID uint) ([]models.PaymentCard, error)
	BlockPaymentCard(cardID int) error
	UnblockPaymentCard(cardID int) error
	ValidatePaymentCard(cardNumber, cvv string, expirationDate time.Time) (bool, error)
}

// PaymentCardServiceImpl реализует функциональность сервиса платежных карт
type PaymentCardServiceImpl struct {
	cardStorage    storage.PaymentCardStorage
	accountStorage storage.FinancialAccountStorage
	encryptionKey  string
	saltBytes      []byte
}

// NewPaymentCardService создаёт новый сервис платежных карт
func NewPaymentCardService(
	cardStorage storage.PaymentCardStorage,
	accountStorage storage.FinancialAccountStorage,
	encryptionKey string,
	saltBytes []byte,
) PaymentCardService {
	return &PaymentCardServiceImpl{
		cardStorage:    cardStorage,
		accountStorage: accountStorage,
		encryptionKey:  encryptionKey,
		saltBytes:      saltBytes,
	}
}

// generateCardNumber генерирует номер платежной карты
func (s *PaymentCardServiceImpl) generateCardNumber() (string, error) {
	// Генерируем 16 цифр номера карты
	var cardNumber string
	for i := 0; i < 4; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			return "", err
		}
		cardNumber += fmt.Sprintf("%04d", n)
	}
	return cardNumber, nil
}

// generateCVV генерирует CVV код карты
func (s *PaymentCardServiceImpl) generateCVV() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", n), nil
}

// encryptCVV шифрует CVV код для безопасного хранения
func (s *PaymentCardServiceImpl) encryptCVV(cvv string) string {
	// Простая демонстрационная функция шифрования
	// В реальном проекте здесь должен быть использован криптографически стойкий метод шифрования
	data := []byte(cvv + s.encryptionKey)
	return base64.StdEncoding.EncodeToString(data)
}

// decryptCVV расшифровывает CVV код из зашифрованного представления
func (s *PaymentCardServiceImpl) decryptCVV(encryptedCVV string) (string, error) {
	// Расшифровка Base64-закодированных данных
	decodedData, err := base64.StdEncoding.DecodeString(encryptedCVV)
	if err != nil {
		return "", errors.New("ошибка расшифровки CVV")
	}
	
	// Отделение CVV от ключа шифрования
	rawCVV := strings.TrimSuffix(string(decodedData), s.encryptionKey)
	return rawCVV, nil
}

// CreatePaymentCard создаёт новую платежную карту
func (s *PaymentCardServiceImpl) CreatePaymentCard(accountID uint, holderName string) (*models.PaymentCard, string, error) {
	// Проверяем существование счета
	_, err := s.accountStorage.GetFinancialAccountByID(accountID)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось найти счет: %v", err)
	}

	// Генерируем номер карты
	cardNumber, err := s.generateCardNumber()
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при генерации номера карты: %v", err)
	}

	// Генерируем CVV
	cvv, err := s.generateCVV()
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при генерации CVV: %v", err)
	}

	// Шифруем CVV
	encryptedCVV := s.encryptCVV(cvv)

	// Устанавливаем срок действия на 5 лет с текущей даты
	expirationDate := time.Now().AddDate(5, 0, 0)

	// Создаем платежную карту
	card := &models.PaymentCard{
		CardNumber:         cardNumber,
		CardholderName:     holderName,
		CVV:                encryptedCVV,
		ExpirationDate:     expirationDate,
		Status:             "active",
		FinancialAccountID: int(accountID),
	}

	// Сохраняем в БД
	err = s.cardStorage.CreatePaymentCard(card)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при сохранении карты: %v", err)
	}

	// Важно: возвращаем оригинальный незашифрованный CVV код вместе с объектом карты
	return card, cvv, nil
}

// GetPaymentCardByID получает платежную карту по ID
func (s *PaymentCardServiceImpl) GetPaymentCardByID(id int) (*models.PaymentCard, error) {
	card, err := s.cardStorage.GetPaymentCardByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карту по ID: %v", err)
	}
	return card, nil
}

// GetPaymentCardByNumber получает платежную карту по номеру
func (s *PaymentCardServiceImpl) GetPaymentCardByNumber(cardNumber string) (*models.PaymentCard, error) {
	card, err := s.cardStorage.GetPaymentCardByNumber(cardNumber)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карту по номеру: %v", err)
	}
	return card, nil
}

// GetPaymentCardsByAccountID получает все платежные карты по ID счета
func (s *PaymentCardServiceImpl) GetPaymentCardsByAccountID(accountID uint) ([]models.PaymentCard, error) {
	cards, err := s.cardStorage.GetPaymentCardsByAccountID(accountID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карты по ID счета: %v", err)
	}
	return cards, nil
}

// GetPaymentCardsByClientID получает все платежные карты по ID клиента
func (s *PaymentCardServiceImpl) GetPaymentCardsByClientID(clientID uint) ([]models.PaymentCard, error) {
	cards, err := s.cardStorage.GetPaymentCardsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карты по ID клиента: %v", err)
	}
	return cards, nil
}

// BlockPaymentCard блокирует платежную карту
func (s *PaymentCardServiceImpl) BlockPaymentCard(cardID int) error {
	card, err := s.cardStorage.GetPaymentCardByID(uint(cardID))
	if err != nil {
		return fmt.Errorf("не удалось получить карту: %v", err)
	}

	card.Status = "blocked"
	if err := s.cardStorage.UpdatePaymentCard(card); err != nil {
		return fmt.Errorf("не удалось заблокировать карту: %v", err)
	}

	return nil
}

// UnblockPaymentCard разблокирует платежную карту
func (s *PaymentCardServiceImpl) UnblockPaymentCard(cardID int) error {
	card, err := s.cardStorage.GetPaymentCardByID(uint(cardID))
	if err != nil {
		return fmt.Errorf("не удалось получить карту: %v", err)
	}

	card.Status = "active"
	if err := s.cardStorage.UpdatePaymentCard(card); err != nil {
		return fmt.Errorf("не удалось разблокировать карту: %v", err)
	}

	return nil
}

// ValidatePaymentCard проверяет валидность платежной карты
func (s *PaymentCardServiceImpl) ValidatePaymentCard(cardNumber, cvv string, expirationDate time.Time) (bool, error) {
	card, err := s.cardStorage.GetPaymentCardByNumber(cardNumber)
	if err != nil {
		return false, errors.New("карта не найдена")
	}

	if card.Status != "active" {
		return false, errors.New("карта заблокирована")
	}

	if card.ExpirationDate.Before(time.Now()) {
		return false, errors.New("срок действия карты истек")
	}

	if card.ExpirationDate.Year() != expirationDate.Year() || card.ExpirationDate.Month() != expirationDate.Month() {
		return false, errors.New("неверный срок действия карты")
	}

	// Упрощенная проверка CVV - просто сравниваем с расшифрованным значением
	// Для тестирования в разработке это позволит использовать статические CVV коды
	decodedData, err := base64.StdEncoding.DecodeString(card.CVV)
	if err != nil {
		return false, errors.New("ошибка проверки CVV")
	}

	// Извлекаем чистый CVV без ключа шифрования
	decodedCVV := strings.TrimSuffix(string(decodedData), s.encryptionKey)
	
	// Сравниваем с введенным CVV
	if decodedCVV != cvv {
		// Для отладки в логах
		log.Printf("CVV не совпадает: ожидается %s, получено %s", decodedCVV, cvv)
		return false, errors.New("неверный CVV")
	}

	return true, nil
}
