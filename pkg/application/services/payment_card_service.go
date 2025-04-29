package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
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

// generateCardNumber генерирует номер платежной карты по алгоритму Луна
func (s *PaymentCardServiceImpl) generateCardNumber() (string, error) {
	// Генерируем 15 цифр номера карты (16-я будет контрольной)
	var baseNumber string
	binPrefix := "4101" // Начинаем номера карт с префикса (например, 41 для Visa)
	
	// Генерируем оставшиеся 11 цифр (4 + 11 = 15)
	for i := 0; i < 3; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10000))
		if err != nil {
			return "", err
		}
		baseNumber += fmt.Sprintf("%04d", n)
	}
	
	// Добавляем еще 1-3 случайных цифры для 15 значащих цифр
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return "", err
	}
	threeDigits := fmt.Sprintf("%03d", n)
	baseNumber = binPrefix + baseNumber + threeDigits[0:3-len(binPrefix+baseNumber)%4]
	
	// Вычисляем контрольную цифру по алгоритму Луна
	checkDigit := s.calculateLuhnCheckDigit(baseNumber)
	
	// Форматируем номер карты в группы по 4 цифры для удобочитаемости
	cardNumber := baseNumber + checkDigit
	formattedCardNumber := s.formatCardNumber(cardNumber)
	
	return strings.ReplaceAll(formattedCardNumber, " ", ""), nil
}

// calculateLuhnCheckDigit вычисляет контрольную цифру по алгоритму Луна
func (s *PaymentCardServiceImpl) calculateLuhnCheckDigit(number string) string {
	// Преобразуем строку в массив цифр
	var digits []int
	for _, r := range number {
		digit := int(r - '0')
		digits = append(digits, digit)
	}
	
	// Алгоритм Луна
	sum := 0
	for i := 0; i < len(digits); i++ {
		digit := digits[len(digits)-1-i]
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	
	// Вычисляем контрольную цифру
	checkDigit := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%d", checkDigit)
}

// validateLuhn проверяет, что номер карты соответствует алгоритму Луна
func (s *PaymentCardServiceImpl) validateLuhn(cardNumber string) bool {
	// Удаляем все нецифровые символы
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")
	
	// Проверяем длину номера карты
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}
	
	// Преобразуем строку в массив цифр
	var digits []int
	for _, r := range cardNumber {
		if r < '0' || r > '9' {
			return false // Если есть нецифровые символы, номер недействителен
		}
		digit := int(r - '0')
		digits = append(digits, digit)
	}
	
	// Обратный порядок для алгоритма Луна
	checkDigit := digits[len(digits)-1]
	digits = digits[:len(digits)-1]
	
	// Алгоритм Луна
	sum := 0
	for i := 0; i < len(digits); i++ {
		digit := digits[len(digits)-1-i]
		if i%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	
	// Проверяем контрольную цифру
	return (sum+checkDigit)%10 == 0
}

// formatCardNumber форматирует номер карты в группы по 4 цифры
func (s *PaymentCardServiceImpl) formatCardNumber(cardNumber string) string {
	var formatted string
	for i := 0; i < len(cardNumber); i += 4 {
		end := i + 4
		if end > len(cardNumber) {
			end = len(cardNumber)
		}
		formatted += cardNumber[i:end]
		if end < len(cardNumber) {
			formatted += " "
		}
	}
	return formatted
}

// generateCVV генерирует CVV код карты
func (s *PaymentCardServiceImpl) generateCVV() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", n), nil
}

// hashCVV хеширует CVV код для безопасного хранения с использованием bcrypt
func (s *PaymentCardServiceImpl) hashCVV(cvv string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyCVV проверяет соответствие CVV его хешу
func (s *PaymentCardServiceImpl) verifyCVV(cvv, hashedCVV string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCVV), []byte(cvv))
	return err == nil
}

// encryptWithPGP шифрует данные с использованием PGP
func (s *PaymentCardServiceImpl) encryptWithPGP(data string) (string, error) {
	// Создаем сущность для шифрования
	entity, err := s.createPGPEntity()
	if err != nil {
		return "", err
	}
	
	// Создаем буфер для зашифрованных данных
	buf := new(bytes.Buffer)
	
	// Создаем armored writer
	armoredWriter, err := armor.Encode(buf, "PGP MESSAGE", nil)
	if err != nil {
		return "", err
	}
	
	// Создаем шифратор
	pgpWriter, err := openpgp.Encrypt(armoredWriter, []*openpgp.Entity{entity}, nil, nil, nil)
	if err != nil {
		armoredWriter.Close()
		return "", err
	}
	
	// Записываем данные для шифрования
	_, err = pgpWriter.Write([]byte(data))
	if err != nil {
		pgpWriter.Close()
		armoredWriter.Close()
		return "", err
	}
	
	// Закрываем писатели
	pgpWriter.Close()
	armoredWriter.Close()
	
	// Возвращаем зашифрованный текст
	return buf.String(), nil
}

// decryptWithPGP расшифровывает данные PGP
func (s *PaymentCardServiceImpl) decryptWithPGP(encryptedData string) (string, error) {
	// Создаем сущность для расшифровки
	entity, err := s.createPGPEntity()
	if err != nil {
		return "", err
	}
	
	// Создаем список сущностей
	entityList := openpgp.EntityList{entity}
	
	// Декодируем armored данные
	block, err := armor.Decode(strings.NewReader(encryptedData))
	if err != nil {
		return "", err
	}
	
	// Создаем объект сообщения
	md, err := openpgp.ReadMessage(block.Body, entityList, nil, nil)
	if err != nil {
		return "", err
	}
	
	// Читаем расшифрованные данные
	decryptedBytes := new(bytes.Buffer)
	_, err = decryptedBytes.ReadFrom(md.UnverifiedBody)
	if err != nil {
		return "", err
	}
	
	return decryptedBytes.String(), nil
}

// createPGPEntity создает PGP сущность из ключа шифрования
func (s *PaymentCardServiceImpl) createPGPEntity() (*openpgp.Entity, error) {
	// Создаем имя для ключа
	name := packet.NewUserId("Finance System", "Card Encryption Key", "cards@financeapp.local")
	
	// Создаем RSA ключевую пару
	entity, err := openpgp.NewEntity("Finance System", "Card Encryption", "cards@financeapp.local", nil)
	if err != nil {
		return nil, err
	}
	
	// Добавляем идентификатор
	entity.Identities[name.Id] = &openpgp.Identity{
		Name:   name.Name,
		UserId: name,
	}
	
	// Генерируем ключ из пароля
	entity.PrivateKey.Encrypted = false
	return entity, nil
}

// computeHMAC вычисляет HMAC для данных
func (s *PaymentCardServiceImpl) computeHMAC(data string) string {
	h := hmac.New(sha256.New, s.saltBytes)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// verifyHMAC проверяет HMAC
func (s *PaymentCardServiceImpl) verifyHMAC(data, expectedHMAC string) bool {
	actualHMAC := s.computeHMAC(data)
	return hmac.Equal([]byte(actualHMAC), []byte(expectedHMAC))
}

// CreatePaymentCard создаёт новую платежную карту
func (s *PaymentCardServiceImpl) CreatePaymentCard(accountID uint, holderName string) (*models.PaymentCard, string, error) {
	// Проверяем существование счета
	_, err := s.accountStorage.GetFinancialAccountByID(accountID)
	if err != nil {
		return nil, "", fmt.Errorf("не удалось найти счет: %v", err)
	}

	// Генерируем номер карты по алгоритму Луна
	cardNumber, err := s.generateCardNumber()
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при генерации номера карты: %v", err)
	}
	
	// Проверяем, что номер соответствует алгоритму Луна
	if !s.validateLuhn(cardNumber) {
		return nil, "", fmt.Errorf("сгенерированный номер карты не соответствует алгоритму Луна")
	}

	// Генерируем CVV
	cvv, err := s.generateCVV()
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при генерации CVV: %v", err)
	}
	
	// Хешируем CVV с помощью bcrypt
	hashedCVV, err := s.hashCVV(cvv)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при хешировании CVV: %v", err)
	}
	
	// Шифруем номер карты с помощью PGP
	encryptedCardNumber, err := s.encryptWithPGP(cardNumber)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при шифровании номера карты: %v", err)
	}
	
	// Вычисляем HMAC для номера карты
	cardNumberHMAC := s.computeHMAC(cardNumber)

	// Устанавливаем срок действия на 5 лет с текущей даты
	expirationDate := time.Now().AddDate(5, 0, 0)
	
	// Шифруем срок действия карты с помощью PGP
	expirationDateStr := expirationDate.Format("01/06") // MM/YY формат
	encryptedExpirationDate, err := s.encryptWithPGP(expirationDateStr)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при шифровании срока действия карты: %v", err)
	}

	// Создаем платежную карту
	card := &models.PaymentCard{
		CardNumber:         encryptedCardNumber,
		CardNumberHMAC:     cardNumberHMAC,     // Сохраняем HMAC для проверки целостности
		CardholderName:     holderName,
		CVV:                hashedCVV,          // Используем хеш CVV
		ExpirationDate:     expirationDate,
		EncryptedExpDate:   encryptedExpirationDate, // Шифруем срок действия
		Status:             "active",
		FinancialAccountID: int(accountID),
	}

	// Сохраняем в БД
	err = s.cardStorage.CreatePaymentCard(card)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка при сохранении карты: %v", err)
	}
	
	// Для отображения клиенту возвращаем карту с расшифрованным номером
	displayCard := *card
	displayCard.CardNumber = cardNumber

	// Возвращаем оригинальный незашифрованный CVV код вместе с объектом карты
	return &displayCard, cvv, nil
}

// GetPaymentCardByID получает платежную карту по ID
func (s *PaymentCardServiceImpl) GetPaymentCardByID(id int) (*models.PaymentCard, error) {
	card, err := s.cardStorage.GetPaymentCardByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карту по ID: %v", err)
	}
	
	// Расшифровываем номер карты и срок действия для отображения
	decryptedCardNumber, err := s.decryptWithPGP(card.CardNumber)
	if err == nil {
		card.CardNumber = decryptedCardNumber
	} else {
		// Игнорируем ошибки расшифровки для обратной совместимости
		log.Printf("Предупреждение: не удалось расшифровать номер карты: %v", err)
	}
	
	if card.EncryptedExpDate != "" {
		decryptedExpDate, err := s.decryptWithPGP(card.EncryptedExpDate)
		if err == nil {
			// Обновляем только для отображения, не меняем в БД
			log.Printf("Расшифрованный срок действия: %s", decryptedExpDate)
		}
	}
	
	return card, nil
}

// GetPaymentCardByNumber получает платежную карту по номеру
func (s *PaymentCardServiceImpl) GetPaymentCardByNumber(cardNumber string) (*models.PaymentCard, error) {
	// Вычисляем HMAC для поиска
	cardNumberHMAC := s.computeHMAC(cardNumber)
	
	// Ищем карту по HMAC
	card, err := s.cardStorage.GetPaymentCardByHMAC(cardNumberHMAC)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карту по номеру: %v", err)
	}
	
	// Расшифровываем номер карты для отображения
	decryptedCardNumber, err := s.decryptWithPGP(card.CardNumber)
	if err == nil {
		card.CardNumber = decryptedCardNumber
	} else {
		// Игнорируем ошибки расшифровки
		log.Printf("Предупреждение: не удалось расшифровать номер карты: %v", err)
	}
	
	return card, nil
}

// GetPaymentCardsByAccountID получает все платежные карты по ID счета
func (s *PaymentCardServiceImpl) GetPaymentCardsByAccountID(accountID uint) ([]models.PaymentCard, error) {
	cards, err := s.cardStorage.GetPaymentCardsByAccountID(accountID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карты по ID счета: %v", err)
	}
	
	// Расшифровываем данные карт для отображения
	for i := range cards {
		decryptedCardNumber, err := s.decryptWithPGP(cards[i].CardNumber)
		if err == nil {
			cards[i].CardNumber = decryptedCardNumber
		} else {
			// Игнорируем ошибки расшифровки
			log.Printf("Предупреждение: не удалось расшифровать данные карты %d: %v", cards[i].ID, err)
		}
	}
	
	return cards, nil
}

// GetPaymentCardsByClientID получает все платежные карты по ID клиента
func (s *PaymentCardServiceImpl) GetPaymentCardsByClientID(clientID uint) ([]models.PaymentCard, error) {
	cards, err := s.cardStorage.GetPaymentCardsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить карты по ID клиента: %v", err)
	}
	
	// Расшифровываем данные карт для отображения
	for i := range cards {
		decryptedCardNumber, err := s.decryptWithPGP(cards[i].CardNumber)
		if err == nil {
			cards[i].CardNumber = decryptedCardNumber
		} else {
			// Игнорируем ошибки расшифровки
			log.Printf("Предупреждение: не удалось расшифровать данные карты %d: %v", cards[i].ID, err)
		}
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

// ValidatePaymentCard проверяет валидность карты
func (s *PaymentCardServiceImpl) ValidatePaymentCard(cardNumber, cvv string, expirationDate time.Time) (bool, error) {
	// Проверяем валидность номера карты по алгоритму Луна
	if !s.validateLuhn(cardNumber) {
		return false, errors.New("недействительный номер карты")
	}
	
	// Получаем карту по номеру
	card, err := s.GetPaymentCardByNumber(cardNumber)
	if err != nil {
		return false, fmt.Errorf("карта не найдена: %v", err)
	}
	
	// Проверяем, не истек ли срок действия
	now := time.Now()
	if now.After(card.ExpirationDate) {
		return false, errors.New("срок действия карты истек")
	}
	
	// Проверяем, не заблокирована ли карта
	if card.Status != "active" {
		return false, errors.New("карта заблокирована")
	}
	
	// Проверяем CVV
	if !s.verifyCVV(cvv, card.CVV) {
		return false, errors.New("неверный CVV код")
	}
	
	return true, nil
}
