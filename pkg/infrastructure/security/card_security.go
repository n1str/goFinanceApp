package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// CardSecurity предоставляет функции безопасности для работы с картами
type CardSecurity struct {
	secret       []byte
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	keyAvailable bool
}

// NewCardSecurity создает новый экземпляр CardSecurity
func NewCardSecurity(secret string) *CardSecurity {
	// Генерация ключей для PGP-шифрования
	privateKey, publicKey, err := generateKeyPair(2048)
	keyAvailable := err == nil

	if err != nil {
		log.Printf("Ошибка при генерации ключей PGP: %v", err)
		log.Printf("Продолжение без PGP-шифрования, используется только HMAC")
	}

	return &CardSecurity{
		secret:       []byte(secret),
		privateKey:   privateKey,
		publicKey:    publicKey,
		keyAvailable: keyAvailable,
	}
}

// ComputeHMAC вычисляет HMAC для данных карты
func (s *CardSecurity) ComputeHMAC(data string) string {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет HMAC
func (s *CardSecurity) VerifyHMAC(data, signature string) bool {
	expectedSignature := s.ComputeHMAC(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// EncryptCardNumber шифрует номер карты с использованием RSA
func (s *CardSecurity) EncryptCardNumber(cardNumber string) (string, error) {
	if !s.keyAvailable {
		return "", errors.New("ключи шифрования недоступны")
	}

	// RSA шифрование
	ciphertext, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		s.publicKey,
		[]byte(cardNumber),
	)
	if err != nil {
		return "", fmt.Errorf("ошибка при шифровании: %w", err)
	}

	// Кодирование в base64 для удобного хранения
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCardNumber расшифровывает номер карты
func (s *CardSecurity) DecryptCardNumber(encryptedCardNumber string) (string, error) {
	if !s.keyAvailable {
		return "", errors.New("ключи шифрования недоступны")
	}

	// Декодируем из base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedCardNumber)
	if err != nil {
		return "", fmt.Errorf("ошибка при декодировании base64: %w", err)
	}

	// RSA расшифровка
	plaintext, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		s.privateKey,
		ciphertext,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка при расшифровке: %w", err)
	}

	return string(plaintext), nil
}

// HashCVV хеширует CVV с помощью bcrypt
func (s *CardSecurity) HashCVV(cvv string) (string, error) {
	hashedCVV, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка при хешировании CVV: %w", err)
	}

	return string(hashedCVV), nil
}

// VerifyCVV проверяет CVV
func (s *CardSecurity) VerifyCVV(cvv, hashedCVV string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedCVV), []byte(cvv))
	return err == nil
}

// GetCardVerificationData создает данные для проверки карты с помощью алгоритма Луна
func (s *CardSecurity) GetCardVerificationData(cardNumber string) string {
	return s.ComputeHMAC(getCleanCardNumber(cardNumber))
}

// VerifyCardNumber проверяет номер карты с помощью алгоритма Луна
func (s *CardSecurity) VerifyCardNumber(cardNumber string) bool {
	cleanNumber := getCleanCardNumber(cardNumber)
	return validateLuhn(cleanNumber)
}

// ExportPublicKeyPEM экспортирует публичный ключ в формате PEM
func (s *CardSecurity) ExportPublicKeyPEM() (string, error) {
	if !s.keyAvailable {
		return "", errors.New("ключи шифрования недоступны")
	}

	// Кодируем публичный ключ в DER формат
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return "", fmt.Errorf("ошибка при кодировании публичного ключа: %w", err)
	}

	// Кодируем в PEM формат
	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	if pemKey == nil {
		return "", errors.New("ошибка при кодировании PEM")
	}

	return string(pemKey), nil
}

// Вспомогательные функции

// generateKeyPair создает пару ключей RSA
func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privKey, &privKey.PublicKey, nil
}

// getCleanCardNumber удаляет все нецифровые символы из номера карты
func getCleanCardNumber(cardNumber string) string {
	var clean []rune
	for _, r := range cardNumber {
		if r >= '0' && r <= '9' {
			clean = append(clean, r)
		}
	}
	return string(clean)
}

// validateLuhn проверяет номер карты с помощью алгоритма Луна
func validateLuhn(number string) bool {
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	sum := 0
	alternate := false

	// Проходим справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// GenerateCardNumber генерирует случайный номер карты, валидный по алгоритму Луна
func GenerateCardNumber(prefix string) (string, error) {
	// Удаляем нецифровые символы из префикса
	prefix = getCleanCardNumber(prefix)

	// Если префикс не задан, используем стандартный
	if prefix == "" {
		prefix = "4100" // Visa
	}

	// Всего длина номера - 16 цифр
	const totalLength = 16
	remainingDigits := totalLength - len(prefix) - 1 // -1 для контрольной цифры

	if remainingDigits < 0 {
		return "", errors.New("префикс слишком длинный")
	}

	// Генерируем случайные цифры для оставшейся части номера
	randomDigits := make([]byte, remainingDigits)
	_, err := rand.Read(randomDigits)
	if err != nil {
		return "", err
	}

	// Преобразуем случайные байты в цифры
	cardNumber := prefix
	for _, b := range randomDigits {
		digit := int(b) % 10
		cardNumber += fmt.Sprintf("%d", digit)
	}

	// Вычисляем контрольную цифру по алгоритму Луна
	checkDigit := calculateLuhnCheckDigit(cardNumber)
	cardNumber += fmt.Sprintf("%d", checkDigit)

	// Форматируем номер карты для удобочитаемости (группы по 4 цифры)
	formatted := ""
	for i, r := range cardNumber {
		if i > 0 && i%4 == 0 {
			formatted += " "
		}
		formatted += string(r)
	}

	return formatted, nil
}

// calculateLuhnCheckDigit вычисляет контрольную цифру по алгоритму Луна
func calculateLuhnCheckDigit(partialNumber string) int {
	// Добавляем 0 как заполнитель для контрольной цифры
	number := partialNumber + "0"

	sum := 0
	alternate := false

	// Проходим справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	// Контрольная цифра - это значение, которое делает сумму кратной 10
	if sum%10 == 0 {
		return 0
	}
	return 10 - (sum % 10)
}
