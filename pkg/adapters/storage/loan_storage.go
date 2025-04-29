package storage

import (
	"FinanceSystem/pkg/domain/models"
	"gorm.io/gorm"
	"log"
)

// LoanStorage определяет интерфейс для работы с хранилищем займов
type LoanStorage interface {
	CreateLoan(loan *models.Loan) error
	GetLoanByID(id uint) (*models.Loan, error)
	GetLoansByClientID(clientID uint) ([]models.Loan, error)
	GetLoansByAccountID(accountID uint) ([]models.Loan, error)
	GetLoansByStatus(status models.LoanStatus) ([]models.Loan, error)
	GetLoansByStatusAndClientID(status models.LoanStatus, clientID uint) ([]models.Loan, error)
	UpdateLoan(loan *models.Loan) error
	GetActiveLoanCount(clientID uint) (int64, error)
	GetTotalLoanAmount(clientID uint) (float64, error)
	GetLoanWithPaymentPlan(id uint) (*models.Loan, []models.PaymentPlan, error)
	SavePaymentPlan(loanID uint, plans []models.PaymentPlan) error
	GetAllLoans() ([]models.Loan, error)
	GetDB() *gorm.DB
}

// LoanStorageImpl реализует функциональность хранилища займов
type LoanStorageImpl struct {
	*BaseRepository
}

// NewLoanStorage создаёт новое хранилище займов
func NewLoanStorage(db *gorm.DB) LoanStorage {
	return &LoanStorageImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// CreateLoan создаёт новый заем
func (r *LoanStorageImpl) CreateLoan(loan *models.Loan) error {
	return r.db.Create(loan).Error
}

// GetLoanByID находит заем по ID
func (r *LoanStorageImpl) GetLoanByID(id uint) (*models.Loan, error) {
	var loan models.Loan
	// Загружаем займ с полной информацией о клиенте и счете
	err := r.db.Preload("Client").Preload("Account").First(&loan, id).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

// GetLoansByClientID находит все займы клиента
func (r *LoanStorageImpl) GetLoansByClientID(clientID uint) ([]models.Loan, error) {
	var loans []models.Loan
	// Предзагружаем связанные данные клиента и счета
	err := r.db.Preload("Client").Preload("Account").Where("client_id = ?", clientID).Find(&loans).Error
	return loans, err
}

// GetLoansByAccountID находит все займы по ID счета
func (r *LoanStorageImpl) GetLoansByAccountID(accountID uint) ([]models.Loan, error) {
	var loans []models.Loan
	// Предзагружаем связанные данные клиента и счета
	err := r.db.Preload("Client").Preload("Account").Where("account_id = ?", accountID).Find(&loans).Error
	return loans, err
}

// GetLoansByStatus находит все займы с указанным статусом
func (r *LoanStorageImpl) GetLoansByStatus(status models.LoanStatus) ([]models.Loan, error) {
	var loans []models.Loan
	err := r.db.Where("status = ?", status).Preload("Client").Preload("Account").Find(&loans).Error
	return loans, err
}

// GetLoansByStatusAndClientID находит все займы с указанным статусом для конкретного клиента
func (r *LoanStorageImpl) GetLoansByStatusAndClientID(status models.LoanStatus, clientID uint) ([]models.Loan, error) {
	var loans []models.Loan
	// Добавляем логирование для отладки
	log.Printf("Запрос займов для клиента ID=%d со статусом %s", clientID, status)
	
	// Используем меньше условий для гарантированного получения данных
	err := r.db.Debug().Where("client_id = ?", clientID).Find(&loans).Error
	
	// Логирование результатов
	log.Printf("Найдено %d займов для клиента ID=%d", len(loans), clientID)
	for i, loan := range loans {
		log.Printf("Заем #%d: ID=%d, Статус=%s, Principal=%.2f, InterestRate=%.2f, MonthlyPayment=%.2f", 
			i+1, loan.ID, loan.Status, loan.Principal, loan.InterestRate, loan.MonthlyPayment)
	}
	
	// Фильтруем займы по статусу в памяти после загрузки
	var activeLoans []models.Loan
	for _, loan := range loans {
		if loan.Status == status {
			activeLoans = append(activeLoans, loan)
		}
	}
	
	log.Printf("После фильтрации по статусу %s: найдено %d займов", status, len(activeLoans))
	
	return activeLoans, err
}

// UpdateLoan обновляет информацию о займе
func (r *LoanStorageImpl) UpdateLoan(loan *models.Loan) error {
	return r.db.Save(loan).Error
}

// GetActiveLoanCount возвращает количество активных займов клиента
func (r *LoanStorageImpl) GetActiveLoanCount(clientID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Loan{}).Where("client_id = ? AND status = ?", clientID, models.LoanStatusActive).Count(&count).Error
	return count, err
}

// GetTotalLoanAmount возвращает общую сумму активных займов клиента
func (r *LoanStorageImpl) GetTotalLoanAmount(clientID uint) (float64, error) {
	var totalAmount float64
	err := r.db.Model(&models.Loan{}).
		Where("client_id = ? AND status = ?", clientID, models.LoanStatusActive).
		Select("SUM(principal)").
		Row().
		Scan(&totalAmount)
	return totalAmount, err
}

// GetLoanWithPaymentPlan возвращает займ и его график платежей
func (r *LoanStorageImpl) GetLoanWithPaymentPlan(id uint) (*models.Loan, []models.PaymentPlan, error) {
	var loan models.Loan
	
	// Загружаем займ с полной информацией о клиенте и счете
	if err := r.db.Preload("Client").Preload("Account").First(&loan, id).Error; err != nil {
		return nil, nil, err
	}

	var paymentPlans []models.PaymentPlan
	// Загружаем график платежей и для каждого платежа загружаем полный займ со связанными данными
	if err := r.db.Preload("Loan").Preload("Loan.Client").Preload("Loan.Account").Where("loan_id = ?", id).Order("installment_num").Find(&paymentPlans).Error; err != nil {
		return nil, nil, err
	}

	// Данные займа уже загружены в Loan, поэтому для совместимости возвращаем и сам займ отдельно
	return &loan, paymentPlans, nil
}

// SavePaymentPlan сохраняет график платежей для займа
func (r *LoanStorageImpl) SavePaymentPlan(loanID uint, plans []models.PaymentPlan) error {
	// Начинаем транзакцию
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Опция отмены транзакции при панике
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Удаляем существующие платежные планы для этого кредита
	if err := tx.Where("loan_id = ?", loanID).Delete(&models.PaymentPlan{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Если нет новых записей для сохранения, просто завершаем транзакцию
	if len(plans) == 0 {
		return tx.Commit().Error
	}
	
	// Устанавливаем ID займа для каждой записи плана платежей
	for i := range plans {
		plans[i].LoanID = loanID
	}
	
	// Сохраняем новые записи
	if err := tx.Create(&plans).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Завершаем транзакцию
	return tx.Commit().Error
}

// GetAllLoans возвращает все займы
func (r *LoanStorageImpl) GetAllLoans() ([]models.Loan, error) {
	var loans []models.Loan
	err := r.db.Preload("Client").Preload("Account").Find(&loans).Error
	return loans, err
}

// GetDB возвращает подключение к базе данных
func (r *LoanStorageImpl) GetDB() *gorm.DB {
	return r.db
}
