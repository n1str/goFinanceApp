package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"fmt"
	"log"
	"math"
	"time"
)

// AnalyticsService представляет интерфейс сервиса аналитики
type AnalyticsService interface {
	GetClientSummary(clientID uint) (*models.AccountSummary, error)
	GetPeriodReport(clientID uint, startDate, endDate time.Time) (*models.PeriodReport, error)
	GetSpendingAnalytics(clientID uint, period string) (*models.SpendingAnalytics, error)
	GenerateFinancialForecast(clientID uint, months int) (*models.FinancialForecast, error)
}

// AnalyticsServiceImpl реализует функциональность сервиса аналитики
type AnalyticsServiceImpl struct {
	operationStorage storage.OperationStorage
	accountStorage   storage.FinancialAccountStorage
	loanStorage      storage.LoanStorage
}

// NewAnalyticsService создаёт новый сервис аналитики
func NewAnalyticsService(
	operationStorage storage.OperationStorage,
	accountStorage storage.FinancialAccountStorage,
	loanStorage storage.LoanStorage,
) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		operationStorage: operationStorage,
		accountStorage:   accountStorage,
		loanStorage:      loanStorage,
	}
}

// GetClientSummary возвращает общую информацию о финансах клиента
func (s *AnalyticsServiceImpl) GetClientSummary(clientID uint) (*models.AccountSummary, error) {
	// Получаем все счета клиента
	accounts, err := s.accountStorage.GetFinancialAccountsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить счета клиента: %v", err)
	}

	summary := &models.AccountSummary{
		TotalAccounts: len(accounts),
	}

	// Считаем общий баланс и количество карт
	var totalCards int
	for _, account := range accounts {
		summary.TotalBalance += account.Funds
		totalCards += len(account.PaymentCards)
	}
	summary.TotalCards = totalCards

	// Получаем информацию о займах
	activeLoanCount, err := s.loanStorage.GetActiveLoanCount(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить информацию о активных займах: %v", err)
	}
	summary.ActiveLoans = int(activeLoanCount)

	// Получаем общую сумму займов
	totalLoanBalance, err := s.loanStorage.GetTotalLoanAmount(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить общую сумму займов: %v", err)
	}
	summary.TotalLoanBalance = totalLoanBalance

	// Получаем все активные займы клиента для расчета ежемесячного платежа
	activeLoans, err := s.loanStorage.GetLoansByStatusAndClientID(models.LoanStatusActive, clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить активные займы клиента: %v", err)
	}

	// Рассчитываем суммарный ежемесячный платеж по всем займам клиента
	var monthlyLoanPayment float64
	
	// Используем запасной метод расчета ежемесячного платежа, если поле MonthlyPayment не заполнено
	for _, loan := range activeLoans {
		log.Printf("Обработка займа ID=%d со статусом %s, ежемесячным платежом %.2f", 
			loan.ID, loan.Status, loan.MonthlyPayment)
			
		if loan.MonthlyPayment > 0 {
			// Используем предварительно рассчитанное значение
			monthlyLoanPayment += loan.MonthlyPayment
		} else {
			// Если по какой-то причине MonthlyPayment не заполнено, рассчитываем его здесь
			// по аннуитетной формуле
			if loan.InterestRate <= 0 {
				// Если процентная ставка равна 0, считаем простой платеж (только основной долг)
				payment := loan.Principal / float64(loan.DurationMonths)
				payment = math.Round(payment*100) / 100
				log.Printf("Простой расчет платежа для займа ID=%d (нулевая ставка): %.2f руб.", loan.ID, payment)
				monthlyLoanPayment += payment
			} else {
				// Стандартный расчет аннуитетного платежа
				monthlyRate := loan.InterestRate / 12 / 100
				payment := loan.Principal * monthlyRate * math.Pow(1+monthlyRate, float64(loan.DurationMonths)) / 
						(math.Pow(1+monthlyRate, float64(loan.DurationMonths)) - 1)
				payment = math.Round(payment*100) / 100
				log.Printf("Расчет аннуитетного платежа для займа ID=%d: %.2f руб.", loan.ID, payment)
				monthlyLoanPayment += payment
			}
		}
	}
	
	log.Printf("Итоговый ежемесячный платеж для клиента ID=%d: %.2f", clientID, monthlyLoanPayment)
	
	// Окончательная проверка: если всё же получилось NaN, устанавливаем 0
	if math.IsNaN(monthlyLoanPayment) {
		log.Printf("ВНИМАНИЕ: Обнаружено значение NaN для ежемесячного платежа, устанавливаем 0")
		monthlyLoanPayment = 0
	}
	
	summary.MonthlyLoanPayment = monthlyLoanPayment

	return summary, nil
}

// GetPeriodReport возвращает отчет о финансах за указанный период
func (s *AnalyticsServiceImpl) GetPeriodReport(clientID uint, startDate, endDate time.Time) (*models.PeriodReport, error) {
	// Нормализуем даты для единообразия
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)
	
	// Логируем интервал дат для диагностики
	log.Printf("Получение отчета за период: с %s по %s для клиента %d", 
		startDate.Format("2006-01-02 15:04:05"), 
		endDate.Format("2006-01-02 15:04:05"),
		clientID)

	// Получаем все счета клиента
	accounts, err := s.accountStorage.GetFinancialAccountsByClientID(clientID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить счета клиента: %v", err)
	}

	report := &models.PeriodReport{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Собираем ID всех счетов клиента
	var accountIDs []uint
	for _, account := range accounts {
		accountIDs = append(accountIDs, uint(account.ID))
	}

	log.Printf("Найдено %d счетов клиента", len(accountIDs))

	// Получаем все операции за период для всех счетов клиента
	var allOperations []models.Operation
	for _, accountID := range accountIDs {
		operations, err := s.operationStorage.GetOperationsByAccountID(accountID)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить операции для счета %d: %v", accountID, err)
		}

		log.Printf("Найдено %d операций для счета %d", len(operations), accountID)

		// Фильтруем операции по дате
		for _, op := range operations {
			// Нормализуем дату операции для сравнения
			opDate := time.Date(op.ExecutedAt.Year(), op.ExecutedAt.Month(), op.ExecutedAt.Day(), 
				op.ExecutedAt.Hour(), op.ExecutedAt.Minute(), op.ExecutedAt.Second(), 0, time.UTC)
			
			if (opDate.Equal(startDate) || opDate.After(startDate)) && 
			   (opDate.Equal(endDate) || opDate.Before(endDate)) {
				allOperations = append(allOperations, op)
				log.Printf("Операция %d(%s) от %s на сумму %.2f включена в отчет", 
					op.ID, op.OperationType, op.ExecutedAt.Format("2006-01-02"), op.Sum)
			} else {
				log.Printf("Операция %d(%s) от %s на сумму %.2f НЕ включена в отчет (вне диапазона)", 
					op.ID, op.OperationType, op.ExecutedAt.Format("2006-01-02"), op.Sum)
			}
		}
	}

	log.Printf("Всего найдено %d операций в заданном периоде", len(allOperations))

	// Рассчитываем доходы и расходы
	for _, op := range allOperations {
		switch op.OperationType {
		case models.OperationDeposit:
			// Пополнение - это доход
			report.Income += op.Sum
			log.Printf("Учтен доход: %.2f (пополнение)", op.Sum)
		case models.OperationWithdraw:
			// Снятие - это расход
			report.Expense += op.Sum
			log.Printf("Учтен расход: %.2f (снятие)", op.Sum)
			// Проверяем, платеж ли это по займу
			if op.Details != "" && len(op.Details) >= 6 && op.Details[0:6] == "Платеж" {
				report.LoanPayments += op.Sum
				log.Printf("Учтен платеж по займу: %.2f", op.Sum)
			}
		case models.OperationTransfer:
			// Перевод, проверяем направление
			sourceFound := false
			targetFound := false
			for _, accountID := range accountIDs {
				if uint(op.SourceAccountID) == accountID {
					sourceFound = true
				}
				if uint(op.TargetAccountID) == accountID {
					targetFound = true
				}
			}

			if sourceFound && !targetFound {
				// Перевод на внешний счет - расход
				report.Expense += op.Sum
				log.Printf("Учтен расход: %.2f (перевод на внешний счет)", op.Sum)
			} else if !sourceFound && targetFound {
				// Перевод с внешнего счета - доход
				report.Income += op.Sum
				log.Printf("Учтен доход: %.2f (перевод с внешнего счета)", op.Sum)
			} else {
				log.Printf("Перевод между своими счетами - игнорируется: %.2f", op.Sum)
			}
			// Если оба счета принадлежат клиенту, игнорируем операцию
		}
	}

	// Рассчитываем чистую прибыль/убыток
	report.NetChange = report.Income - report.Expense

	log.Printf("Итоговый отчет: доход %.2f, расход %.2f, чистое изменение %.2f, платежи по займам %.2f", 
		report.Income, report.Expense, report.NetChange, report.LoanPayments)

	return report, nil
}

// GetSpendingAnalytics возвращает анализ расходов по категориям
func (s *AnalyticsServiceImpl) GetSpendingAnalytics(clientID uint, period string) (*models.SpendingAnalytics, error) {
	// Определяем начальную и конечную дату периода
	now := time.Now()
	var startDate time.Time
	
	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		return nil, fmt.Errorf("неизвестный период: %s", period)
	}

	// Получаем отчет за период
	report, err := s.GetPeriodReport(clientID, startDate, now)
	if err != nil {
		return nil, err
	}

	// Условная категоризация расходов
	// В реальном приложении здесь должна быть более сложная логика
	totalExpense := report.Expense
	
	categories := []models.SpendingCategory{
		{
			Name:       "Займы",
			Amount:     report.LoanPayments,
			Percentage: math.Round(report.LoanPayments / totalExpense * 100),
		},
		{
			Name:       "Другое",
			Amount:     totalExpense - report.LoanPayments,
			Percentage: math.Round((totalExpense - report.LoanPayments) / totalExpense * 100),
		},
	}

	analytics := &models.SpendingAnalytics{
		Period:     period,
		Categories: categories,
		Total:      totalExpense,
	}

	return analytics, nil
}

// GenerateFinancialForecast генерирует финансовый прогноз на основе исторических данных
func (s *AnalyticsServiceImpl) GenerateFinancialForecast(clientID uint, months int) (*models.FinancialForecast, error) {
	// Получаем текущую сводку по клиенту
	summary, err := s.GetClientSummary(clientID)
	if err != nil {
		return nil, err
	}

	// Получаем отчет за последний месяц
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)
	lastMonthReport, err := s.GetPeriodReport(clientID, monthAgo, now)
	if err != nil {
		return nil, err
	}

	// Простая линейная экстраполяция на основе последнего месяца
	// В реальном приложении здесь должны быть более сложные алгоритмы прогнозирования
	forecast := &models.FinancialForecast{
		Title:       "Прогноз финансового состояния",
		Description: fmt.Sprintf("Прогноз на следующие %d месяцев", months),
		Confidence:  0.7, // Условное значение достоверности прогноза
	}

	// Добавляем текущую точку
	forecast.DataPoints = append(forecast.DataPoints, models.ForecastPoint{
		Date:   now,
		Value:  summary.TotalBalance,
		IsReal: true,
	})

	// Используем NetChange из отчета за последний месяц как предполагаемое изменение
	monthlyChange := lastMonthReport.NetChange

	// Генерируем точки прогноза
	currentValue := summary.TotalBalance
	for i := 1; i <= months; i++ {
		// Простое линейное изменение баланса
		currentValue += monthlyChange
		forecast.DataPoints = append(forecast.DataPoints, models.ForecastPoint{
			Date:   now.AddDate(0, i, 0),
			Value:  currentValue,
			IsReal: false,
		})
	}

	return forecast, nil
}
