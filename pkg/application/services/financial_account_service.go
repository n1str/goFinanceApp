package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"errors"
	"fmt"
	"time"
)

// FinancialAccountService представляет интерфейс сервиса финансовых счетов
type FinancialAccountService interface {
	// Операции со счетами
	CreateFinancialAccount(account *models.FinancialAccount, clientID uint) error
	GetFinancialAccountByID(id uint) (*models.FinancialAccount, error)
	GetFinancialAccountsByClientID(clientID uint) ([]models.FinancialAccount, error)
	GetAllFinancialAccounts() ([]models.FinancialAccount, error)

	// Операции с денежными средствами
	AddFunds(accountID uint, amount float64, details string) error
	WithdrawFunds(accountID uint, amount float64, details string) error
	TransferFunds(sourceAccountID, targetAccountID uint, amount float64, details string) error

	// Операции с историей
	GetAccountOperations(accountID uint) ([]models.Operation, error)
}

// FinancialAccountServiceImpl реализует функциональность сервиса финансовых счетов
type FinancialAccountServiceImpl struct {
	accountStorage  storage.FinancialAccountStorage
	operationStorage storage.OperationStorage
}

// NewFinancialAccountService создаёт новый сервис финансовых счетов
func NewFinancialAccountService(
	accountStorage storage.FinancialAccountStorage,
	operationStorage storage.OperationStorage,
) FinancialAccountService {
	return &FinancialAccountServiceImpl{
		accountStorage:  accountStorage,
		operationStorage: operationStorage,
	}
}

// CreateFinancialAccount создаёт новый финансовый счет
func (s *FinancialAccountServiceImpl) CreateFinancialAccount(account *models.FinancialAccount, clientID uint) error {
	fmt.Println("Создание счета для клиента с ID:", clientID)
	
	// Устанавливаем текущую дату и время в UTC формате
	now := time.Now().UTC()
	// Установка только даты без времени
	account.OpenedDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	
	if err := s.accountStorage.CreateFinancialAccount(account, clientID); err != nil {
		return fmt.Errorf("не удалось создать счет: %v", err)
	}
	return nil
}

// GetFinancialAccountByID получает финансовый счет по ID
func (s *FinancialAccountServiceImpl) GetFinancialAccountByID(id uint) (*models.FinancialAccount, error) {
	account, err := s.accountStorage.GetFinancialAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить счет по ID: %v", err)
	}
	if account == nil {
		return nil, fmt.Errorf("счет с ID %d не найден", id)
	}
	return account, nil
}

// GetFinancialAccountsByClientID получает все финансовые счета клиента
func (s *FinancialAccountServiceImpl) GetFinancialAccountsByClientID(clientID uint) ([]models.FinancialAccount, error) {
	accounts, err := s.accountStorage.GetFinancialAccountsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить счета клиента: %v", err)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("счета не найдены для клиента с ID: %d", clientID)
	}
	return accounts, nil
}

// GetAllFinancialAccounts получает все финансовые счета
func (s *FinancialAccountServiceImpl) GetAllFinancialAccounts() ([]models.FinancialAccount, error) {
	accounts, err := s.accountStorage.GetAllFinancialAccounts()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить все счета: %v", err)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("счета не найдены")
	}
	return accounts, nil
}

// AddFunds пополняет счет
func (s *FinancialAccountServiceImpl) AddFunds(accountID uint, amount float64, details string) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}

	account, err := s.accountStorage.GetFinancialAccountByID(accountID)
	if err != nil {
		return fmt.Errorf("не удалось получить счет: %v", err)
	}

	// Создаем операцию
	operation := &models.Operation{
		OperationType:   models.OperationDeposit,
		TargetAccountID: int(accountID),
		Sum:             amount,
		Details:         details,
		ExecutedAt:      time.Now(),
		Result:          "completed",
	}

	// Обновляем баланс счета
	account.Funds += amount
	if err := s.accountStorage.UpdateFinancialAccount(account); err != nil {
		return fmt.Errorf("не удалось обновить баланс счета: %v", err)
	}

	// Сохраняем операцию
	if err := s.operationStorage.CreateOperation(operation); err != nil {
		return fmt.Errorf("не удалось создать запись об операции: %v", err)
	}

	return nil
}

// WithdrawFunds снимает средства со счета
func (s *FinancialAccountServiceImpl) WithdrawFunds(accountID uint, amount float64, details string) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}

	account, err := s.accountStorage.GetFinancialAccountByID(accountID)
	if err != nil {
		return fmt.Errorf("не удалось получить счет: %v", err)
	}

	if account.Funds < amount {
		return errors.New("недостаточно средств на счете")
	}

	// Создаем операцию
	operation := &models.Operation{
		OperationType:   models.OperationWithdraw,
		SourceAccountID: int(accountID),
		Sum:             amount,
		Details:         details,
		ExecutedAt:      time.Now(),
		Result:          "completed",
	}

	// Обновляем баланс счета
	account.Funds -= amount
	if err := s.accountStorage.UpdateFinancialAccount(account); err != nil {
		return fmt.Errorf("не удалось обновить баланс счета: %v", err)
	}

	// Сохраняем операцию
	if err := s.operationStorage.CreateOperation(operation); err != nil {
		return fmt.Errorf("не удалось создать запись об операции: %v", err)
	}

	return nil
}

// TransferFunds переводит средства между счетами
func (s *FinancialAccountServiceImpl) TransferFunds(sourceAccountID, targetAccountID uint, amount float64, details string) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}

	if sourceAccountID == targetAccountID {
		return errors.New("невозможно перевести средства на тот же счет")
	}

	sourceAccount, err := s.accountStorage.GetFinancialAccountByID(sourceAccountID)
	if err != nil {
		return fmt.Errorf("не удалось получить исходный счет: %v", err)
	}

	targetAccount, err := s.accountStorage.GetFinancialAccountByID(targetAccountID)
	if err != nil {
		return fmt.Errorf("не удалось получить целевой счет: %v", err)
	}

	if sourceAccount.Funds < amount {
		return errors.New("недостаточно средств на исходном счете")
	}

	// Создаем операцию
	operation := &models.Operation{
		OperationType:   models.OperationTransfer,
		SourceAccountID: int(sourceAccountID),
		TargetAccountID: int(targetAccountID),
		Sum:             amount,
		Details:         details,
		ExecutedAt:      time.Now(),
		Result:          "completed",
	}

	// Обновляем балансы счетов
	sourceAccount.Funds -= amount
	targetAccount.Funds += amount

	if err := s.accountStorage.UpdateFinancialAccount(sourceAccount); err != nil {
		return fmt.Errorf("не удалось обновить баланс исходного счета: %v", err)
	}

	if err := s.accountStorage.UpdateFinancialAccount(targetAccount); err != nil {
		return fmt.Errorf("не удалось обновить баланс целевого счета: %v", err)
	}

	// Сохраняем операцию
	if err := s.operationStorage.CreateOperation(operation); err != nil {
		return fmt.Errorf("не удалось создать запись об операции: %v", err)
	}

	return nil
}

// GetAccountOperations получает все операции по счету
func (s *FinancialAccountServiceImpl) GetAccountOperations(accountID uint) ([]models.Operation, error) {
	operations, err := s.operationStorage.GetOperationsByAccountID(accountID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить операции по счету: %v", err)
	}
	return operations, nil
}
