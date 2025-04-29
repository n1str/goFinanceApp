package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// LoanService представляет интерфейс сервиса займов
type LoanService interface {
	CreateLoan(clientID, accountID uint, principal float64, interestRate float64, durationMonths int, purpose string) (*models.Loan, error)
	GetLoanByID(id uint) (*models.Loan, error)
	GetClientLoans(clientID uint) ([]models.Loan, error)
	GetLoansByStatus(status models.LoanStatus) ([]models.Loan, error)
	GetLoanWithPaymentPlan(loanID uint) (*models.Loan, []models.PaymentPlan, error)
	MakePayment(loanID uint, amount float64) error
	CalculateMonthlyPayment(principal float64, interestRate float64, durationMonths int) float64
	GeneratePaymentPlan(loan *models.Loan) ([]models.PaymentPlan, error)
	UpdateAllLoanMonthlyPayments() error
	GetAllLoans() ([]models.Loan, error)
	UpdatePaymentPlan(plan *models.PaymentPlan) error
	GetClientDebtRatio(clientID uint) (float64, error)
	PredictBalance(clientID uint, days int) ([]models.BalancePrediction, error)
}

// LoanServiceImpl реализует функциональность сервиса займов
type LoanServiceImpl struct {
	loanStorage      storage.LoanStorage
	accountStorage   storage.FinancialAccountStorage
	operationStorage storage.OperationStorage
	externalService  ExternalService
}

// NewLoanService создаёт новый сервис займов
func NewLoanService(
	loanStorage storage.LoanStorage,
	accountStorage storage.FinancialAccountStorage,
	operationStorage storage.OperationStorage,
	externalService ExternalService,
) LoanService {
	return &LoanServiceImpl{
		loanStorage:      loanStorage,
		accountStorage:   accountStorage,
		operationStorage: operationStorage,
		externalService:  externalService,
	}
}

// CalculateMonthlyPayment рассчитывает ежемесячный платеж по займу
func (s *LoanServiceImpl) CalculateMonthlyPayment(principal float64, interestRate float64, durationMonths int) float64 {
	// Защита от ошибок при расчете
	if durationMonths <= 0 {
		log.Printf("ОШИБКА: Срок займа должен быть положительным числом")
		return 0
	}

	// Для займов с нулевой процентной ставкой используем простое деление
	if interestRate <= 0 {
		payment := principal / float64(durationMonths)
		return math.Round(payment*100) / 100
	}

	// Для обычных займов используем формулу аннуитетного платежа
	monthlyRate := interestRate / 12 / 100
	payment := principal * monthlyRate * math.Pow(1+monthlyRate, float64(durationMonths)) / (math.Pow(1+monthlyRate, float64(durationMonths)) - 1)
	return math.Round(payment*100) / 100
}

// CreateLoan создает новый займ
func (s *LoanServiceImpl) CreateLoan(clientID, accountID uint, principal float64, interestRate float64, durationMonths int, purpose string) (*models.Loan, error) {
	// Проверка на существование счета
	account, err := s.accountStorage.GetFinancialAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти счет: %v", err)
	}

	// Расчет ежемесячного платежа
	monthlyPayment := s.CalculateMonthlyPayment(principal, interestRate, durationMonths)
	log.Printf("Рассчитан ежемесячный платеж для займа: %.2f руб. (сумма: %.2f, ставка: %.2f%%, срок: %d мес.)", 
		monthlyPayment, principal, interestRate, durationMonths)

	// Создание займа
	now := time.Now().UTC() // Используем UTC для единообразия
	
	// Очищаем время, оставляем только дату
	issueDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	maturityDate := issueDate.AddDate(0, durationMonths, 0)
	
	loan := &models.Loan{
		ClientID:       clientID,
		AccountID:      accountID,
		Principal:      principal,
		InterestRate:   interestRate,
		DurationMonths: durationMonths,
		MonthlyPayment: monthlyPayment,
		Status:         models.LoanStatusActive,
		IssueDate:      issueDate,
		MaturityDate:   maturityDate,
		Purpose:        purpose,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Сохранение займа
	if err := s.loanStorage.CreateLoan(loan); err != nil {
		return nil, fmt.Errorf("ошибка при создании займа: %v", err)
	}

	// Создание операции выдачи займа
	operation := &models.Operation{
		OperationType:   models.OperationLoanIssue,
		TargetAccountID: int(accountID),
		Sum:             principal,
		Details:         fmt.Sprintf("Выдача займа №%d", loan.ID),
		ExecutedAt:      now,
		Result:          "completed",
	}

	if err := s.operationStorage.CreateOperation(operation); err != nil {
		return nil, fmt.Errorf("ошибка при создании операции для займа: %v", err)
	}

	// Обновление баланса счета
	account.Funds += principal
	if err := s.accountStorage.UpdateFinancialAccount(account); err != nil {
		return nil, fmt.Errorf("ошибка при обновлении баланса счета: %v", err)
	}

	// Генерация графика платежей
	paymentPlans, err := s.GeneratePaymentPlan(loan)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании графика платежей: %v", err)
	}
	
	// Сохранение графика платежей
	if err := s.loanStorage.SavePaymentPlan(loan.ID, paymentPlans); err != nil {
		return nil, fmt.Errorf("ошибка при сохранении графика платежей: %v", err)
	}

	return loan, nil
}

// GetLoanByID получает займ по ID
func (s *LoanServiceImpl) GetLoanByID(id uint) (*models.Loan, error) {
	loan, err := s.loanStorage.GetLoanByID(id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить займ: %v", err)
	}
	return loan, nil
}

// GetClientLoans получает все займы клиента
func (s *LoanServiceImpl) GetClientLoans(clientID uint) ([]models.Loan, error) {
	loans, err := s.loanStorage.GetLoansByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить займы клиента: %v", err)
	}
	return loans, nil
}

// GetLoansByStatus получает все займы по статусу
func (s *LoanServiceImpl) GetLoansByStatus(status models.LoanStatus) ([]models.Loan, error) {
	return s.loanStorage.GetLoansByStatus(status)
}

// GetLoanWithPaymentPlan получает займ с графиком платежей
func (s *LoanServiceImpl) GetLoanWithPaymentPlan(loanID uint) (*models.Loan, []models.PaymentPlan, error) {
	return s.loanStorage.GetLoanWithPaymentPlan(loanID)
}

// MakePayment выполняет платеж по займу
func (s *LoanServiceImpl) MakePayment(loanID uint, amount float64) error {
	// Получение займа и графика платежей
	loan, paymentPlans, err := s.loanStorage.GetLoanWithPaymentPlan(loanID)
	if err != nil {
		return fmt.Errorf("не удалось получить займ: %v", err)
	}

	// Поиск платежа в графике
	var targetPayment *models.PaymentPlan
	for i := range paymentPlans {
		if paymentPlans[i].Status == models.PaymentStatusScheduled {
			targetPayment = &paymentPlans[i]
			break
		}
	}

	if targetPayment == nil {
		return errors.New("платеж не найден в графике")
	}

	if targetPayment.Status == models.PaymentStatusCompleted {
		return errors.New("платеж уже выполнен")
	}

	// Получение счета
	account, err := s.accountStorage.GetFinancialAccountByID(loan.AccountID)
	if err != nil {
		return fmt.Errorf("не удалось получить счет: %v", err)
	}

	// Проверка наличия средств
	if account.Funds < amount {
		return errors.New("недостаточно средств на счете")
	}

	// Обновление счета
	account.Funds -= amount
	if err := s.accountStorage.UpdateFinancialAccount(account); err != nil {
		return fmt.Errorf("ошибка при обновлении баланса счета: %v", err)
	}

	// Создание операции платежа
	now := time.Now()
	operation := &models.Operation{
		OperationType:   models.OperationWithdraw,
		SourceAccountID: int(loan.AccountID),
		Sum:             amount,
		Details:         fmt.Sprintf("Платеж по займу №%d", loan.ID),
		ExecutedAt:      now,
		Result:          "completed",
	}

	if err := s.operationStorage.CreateOperation(operation); err != nil {
		return fmt.Errorf("ошибка при создании операции платежа: %v", err)
	}

	// Обновление статуса платежа
	targetPayment.Status = models.PaymentStatusCompleted
	targetPayment.PaidDate = &now
	targetPayment.UpdatedAt = now

	// Проверка, является ли этот платеж последним
	allPaid := true
	for _, p := range paymentPlans {
		if p.ID != targetPayment.ID && p.Status != models.PaymentStatusCompleted {
			allPaid = false
			break
		}
	}

	// Если все платежи выполнены, обновляем статус займа
	if allPaid {
		loan.Status = models.LoanStatusCompleted
		loan.UpdatedAt = now
		if err := s.loanStorage.UpdateLoan(loan); err != nil {
			return fmt.Errorf("ошибка при обновлении статуса займа: %v", err)
		}
	}

	return nil
}

// GeneratePaymentPlan создает график платежей для займа
func (s *LoanServiceImpl) GeneratePaymentPlan(loan *models.Loan) ([]models.PaymentPlan, error) {
	var paymentPlans []models.PaymentPlan

	principal := loan.Principal
	monthlyRate := loan.InterestRate / 12 / 100
	duration := loan.DurationMonths
	monthlyPayment := loan.MonthlyPayment

	remainingPrincipal := principal
	now := time.Now().UTC() // Используем UTC для единообразия

	for i := 1; i <= duration; i++ {
		interestPayment := remainingPrincipal * monthlyRate
		principalPayment := monthlyPayment - interestPayment

		if i == duration {
			// Корректировка последнего платежа
			principalPayment = remainingPrincipal
			monthlyPayment = principalPayment + interestPayment
		}

		remainingPrincipal -= principalPayment

		// Используем время без временной зоны для даты платежа
		dueDate := loan.IssueDate.AddDate(0, i, 0)
		// Очищаем время, оставляем только дату
		dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.UTC)

		paymentPlan := models.PaymentPlan{
			LoanID:           loan.ID,
			InstallmentNum:   i,
			DueDate:          dueDate,
			Total:            math.Round((principalPayment+interestPayment)*100) / 100,
			InterestPortion:  math.Round(interestPayment*100) / 100,
			PrincipalPortion: math.Round(principalPayment*100) / 100,
			TotalPayment:     math.Round(monthlyPayment*100) / 100,
			Status:           models.PaymentStatusScheduled,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		paymentPlans = append(paymentPlans, paymentPlan)
	}

	return paymentPlans, nil
}

// UpdateAllLoanMonthlyPayments обновляет значение ежемесячного платежа для всех займов
func (s *LoanServiceImpl) UpdateAllLoanMonthlyPayments() error {
	// Получаем все займы
	loans, err := s.loanStorage.GetAllLoans()
	if err != nil {
		return fmt.Errorf("не удалось получить список займов: %v", err)
	}

	log.Printf("Обновление ежемесячных платежей для %d займов", len(loans))
	updatedCount := 0

	// Обновляем каждый займ
	for _, loan := range loans {
		// Если платеж не установлен, рассчитываем его
		if loan.MonthlyPayment <= 0 {
			// Рассчитываем платеж
			monthlyPayment := s.CalculateMonthlyPayment(loan.Principal, loan.InterestRate, loan.DurationMonths)
			
			// Обновляем значение
			loan.MonthlyPayment = monthlyPayment
			log.Printf("Займ ID=%d: обновлен платеж с %.2f на %.2f", 
				loan.ID, 0.0, monthlyPayment)
			
			// Сохраняем изменения
			if err := s.loanStorage.UpdateLoan(&loan); err != nil {
				log.Printf("ОШИБКА при обновлении займа ID=%d: %v", loan.ID, err)
				continue
			}
			updatedCount++
		}
	}

	log.Printf("Обновлено %d займов из %d", updatedCount, len(loans))
	return nil
}

// GetAllLoans получает все займы
func (s *LoanServiceImpl) GetAllLoans() ([]models.Loan, error) {
	return s.loanStorage.GetAllLoans()
}

// UpdatePaymentPlan обновляет план платежей
func (s *LoanServiceImpl) UpdatePaymentPlan(plan *models.PaymentPlan) error {
	// Проверяем, существует ли план платежей
	_, paymentPlans, err := s.loanStorage.GetLoanWithPaymentPlan(plan.LoanID)
	if err != nil {
		return fmt.Errorf("не удалось найти план платежей: %w", err)
	}

	found := false
	for _, p := range paymentPlans {
		if p.ID == plan.ID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("план платежей с ID %d не найден", plan.ID)
	}

	// Обновляем план платежей в базе данных
	// Для простоты, создаем новый запрос в репозитории
	db := s.loanStorage.(*storage.LoanStorageImpl).GetDB()
	if err := db.Save(plan).Error; err != nil {
		return fmt.Errorf("не удалось обновить план платежей: %w", err)
	}

	log.Printf("План платежей ID=%d успешно обновлен", plan.ID)
	return nil
}

// GetClientDebtRatio вычисляет коэффициент долговой нагрузки клиента
func (s *LoanServiceImpl) GetClientDebtRatio(clientID uint) (float64, error) {
	// Получаем все активные займы клиента
	loans, err := s.loanStorage.GetLoansByStatusAndClientID(models.LoanStatusActive, clientID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении активных займов клиента: %w", err)
	}
	
	// Если нет активных займов, долговая нагрузка равна 0
	if len(loans) == 0 {
		return 0, nil
	}
	
	// Вычисляем общую сумму ежемесячных платежей по всем активным займам
	var totalMonthlyPayments float64
	for _, loan := range loans {
		totalMonthlyPayments += loan.MonthlyPayment
	}
	
	// Получаем операции клиента за последние 3 месяца для оценки среднемесячного дохода
	// Упрощенно, предполагаем средний месячный доход 100000
	monthlyIncome := float64(100000)
	
	if monthlyIncome <= 0 {
		// Защита от деления на ноль
		return 0, fmt.Errorf("невозможно вычислить коэффициент долговой нагрузки: отсутствуют данные о доходах")
	}
	
	// Вычисляем коэффициент долговой нагрузки как отношение платежей к доходу
	debtRatio := totalMonthlyPayments / monthlyIncome
	
	return debtRatio, nil
}

// PredictBalance прогнозирует баланс счетов клиента на указанное количество дней вперед
func (s *LoanServiceImpl) PredictBalance(clientID uint, days int) ([]models.BalancePrediction, error) {
	if days <= 0 {
		return nil, fmt.Errorf("количество дней должно быть положительным числом")
	}
	
	// Ограничиваем максимальный период до 365 дней, согласно ТЗ
	if days > 365 {
		days = 365
	}
	
	// Получаем все счета клиента
	accounts, err := s.accountStorage.GetFinancialAccountsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении счетов клиента: %w", err)
	}
	
	// Получаем все активные займы клиента
	loans, err := s.loanStorage.GetLoansByStatusAndClientID(models.LoanStatusActive, clientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении активных займов клиента: %w", err)
	}
	
	// Вычисляем общий начальный баланс клиента по всем счетам
	var initialBalance float64
	for _, account := range accounts {
		initialBalance += account.Funds
	}
	
	// Получаем историю операций за последние 30 дней для анализа паттернов расходов и доходов
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	
	// Получаем список ID счетов
	accountIDs := make([]uint, 0, len(accounts))
	for _, account := range accounts {
		accountIDs = append(accountIDs, uint(account.ID))
	}
	
	// Анализируем доходы и расходы на основе истории операций
	var totalIncome, totalExpense float64
	
	// Получаем операции по каждому счету отдельно
	for _, accountID := range accountIDs {
		operations, err := s.operationStorage.GetOperationsByAccountID(accountID)
		if err != nil {
			log.Printf("Ошибка при получении операций для счета %d: %v", accountID, err)
			continue
		}
		
		// Фильтруем операции за последние 30 дней
		for _, op := range operations {
			if op.ExecutedAt.Before(startDate) || op.ExecutedAt.After(endDate) {
				continue
			}
			
			// Анализируем тип операции
			switch op.OperationType {
			case models.OperationDeposit:
				// Пополнение - это доход
				totalIncome += op.Sum
			case models.OperationLoanIssue:
				// Выдача займа - это доход
				if op.TargetAccountID == int(accountID) {
					totalIncome += op.Sum
				}
			case models.OperationWithdraw:
				// Снятие - это расход
				totalExpense += op.Sum
			case models.OperationTransfer:
				// Для переводов учитываем направление
				if op.SourceAccountID == int(accountID) {
					// Перевод с этого счета - расход
					totalExpense += op.Sum
				} else if op.TargetAccountID == int(accountID) {
					// Перевод на этот счет - доход
					totalIncome += op.Sum
				}
			}
		}
	}
	
	// Вычисляем среднесуточные значения
	days30 := float64(30) // период анализа
	dailyIncome := totalIncome / days30
	dailyExpense := totalExpense / days30
	
	// Если данных недостаточно, используем минимальные базовые значения
	if dailyIncome < 10 {
		dailyIncome = 100 // базовый ежедневный доход
	}
	if dailyExpense < 5 {
		dailyExpense = 70 // базовый ежедневный расход
	}
	
	// Создаем прогнозы на каждый день
	predictions := make([]models.BalancePrediction, days)
	currentBalance := initialBalance
	today := time.Now().Truncate(24 * time.Hour)
	
	// Добавляем небольшую случайную вариацию для более реалистичного прогноза
	rand.Seed(time.Now().UnixNano())
	
	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, i)
		
		// Ежедневная вариация доходов и расходов для реалистичности (±20%)
		dayIncome := dailyIncome * (0.8 + 0.4*rand.Float64())
		dayExpense := dailyExpense * (0.8 + 0.4*rand.Float64())
		
		// Проверяем платежи по займам на этот день
		var loanPayments float64
		var paymentDetails []models.PaymentDetail
		isPaymentDate := false
		
		for _, loan := range loans {
			// Получаем график платежей для каждого займа
			_, paymentPlans, err := s.loanStorage.GetLoanWithPaymentPlan(loan.ID)
			if err != nil {
				log.Printf("Ошибка при получении графика платежей для займа %d: %v", loan.ID, err)
				continue
			}
			
			// Проверяем существующие платежи в графике
			existingPaymentFound := false
			
			for _, plan := range paymentPlans {
				// Проверяем, есть ли платеж на этот день
				planDate := plan.DueDate.Truncate(24 * time.Hour)
				if planDate.Equal(date) && plan.Status != models.PaymentStatusCompleted {
					loanPayments += plan.Total
					isPaymentDate = true
					existingPaymentFound = true
					
					paymentDetails = append(paymentDetails, models.PaymentDetail{
						LoanID:      loan.ID,
						DueDate:     plan.DueDate,
						Amount:      plan.Total,
						Description: fmt.Sprintf("Платеж по кредиту №%d", loan.ID),
						Status:      string(plan.Status),
					})
				}
			}
			
			// Если у нас нет существующего платежа на эту дату, проверяем, должен ли быть 
			// ежемесячный платеж по займу в этот день на основе даты создания займа
			if !existingPaymentFound && loan.MonthlyPayment > 0 {
				// Определяем день платежа на основе даты создания займа
				paymentDayOfMonth := loan.CreatedAt.Day()
				currentDayOfMonth := date.Day()
				
				// Проверяем, совпадает ли день месяца с днем платежа
				// Также учитываем особые случаи (например, 31-е число в месяцах с 30 днями)
				isPaymentDay := currentDayOfMonth == paymentDayOfMonth
				
				// Для месяцев с меньшим количеством дней: если день платежа выпадает на несуществующий день,
				// то последний день месяца становится днем платежа
				if !isPaymentDay && paymentDayOfMonth > 28 {
					// Получаем последний день текущего месяца
					currentYear, currentMonth, _ := date.Date()
					lastDayOfMonth := time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
					
					// Если текущий день - последний в месяце И день платежа больше, чем количество дней в месяце
					if currentDayOfMonth == lastDayOfMonth && paymentDayOfMonth > lastDayOfMonth {
						isPaymentDay = true
					}
				}
				
				// Если сегодня день платежа и этот день находится в период действия займа
				if isPaymentDay && date.After(loan.CreatedAt) && 
				   (loan.MaturityDate.IsZero() || date.Before(loan.MaturityDate)) {
					// Проверяем, что между этой датой и датой создания прошел хотя бы месяц
					monthsPassed := monthsBetween(loan.CreatedAt, date)
					if monthsPassed > 0 {
						loanPayments += loan.MonthlyPayment
						isPaymentDate = true
						
						paymentDetails = append(paymentDetails, models.PaymentDetail{
							LoanID:      loan.ID,
							DueDate:     date,
							Amount:      loan.MonthlyPayment,
							Description: fmt.Sprintf("Регулярный платеж по кредиту №%d", loan.ID),
							Status:      string(models.PaymentStatusScheduled),
						})
					}
				}
			}
		}
		
		// Вычисляем прогнозируемый баланс на этот день
		dailyNet := dayIncome - dayExpense
		plannedIncome := dayIncome
		plannedExpenses := dayExpense
		
		if isPaymentDate {
			plannedExpenses += loanPayments
		}
		
		predictedBalance := currentBalance + dailyNet - loanPayments
		
		predictions[i] = models.BalancePrediction{
			Date:             date,
			InitialBalance:   initialBalance,
			PredictedBalance: predictedBalance,
			PlannedIncome:    plannedIncome,
			PlannedExpenses:  plannedExpenses,
			LoanPayments:     loanPayments,
			IsPaymentDate:    isPaymentDate,
			PaymentDetails:   paymentDetails,
		}
		
		// Обновляем баланс для следующего дня
		currentBalance = predictedBalance
	}
	
	return predictions, nil
}

// monthsBetween вычисляет количество месяцев между двумя датами
func monthsBetween(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	
	years := end.Year() - start.Year()
	months := int(end.Month() - start.Month())
	
	totalMonths := years*12 + months
	
	// Если день в конечной дате меньше, чем в начальной, то месяц еще не полностью прошел
	if end.Day() < start.Day() {
		totalMonths--
	}
	
	return totalMonths
}

// getAccountIDs извлекает массив ID счетов из массива счетов
func getAccountIDs(accounts []models.FinancialAccount) []uint {
	ids := make([]uint, len(accounts))
	for i, account := range accounts {
		ids[i] = uint(account.ID)
	}
	return ids
}
