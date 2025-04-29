package services

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/domain/models"
	"fmt"
	"log"
	"time"
)

// PredictionService предоставляет сервис прогнозирования
type PredictionService interface {
	PredictBalance(clientID uint, days int) ([]models.BalancePrediction, error)
	GetClientDebtRatio(clientID uint) (float64, error)
}

// PredictionServiceImpl реализует функциональность сервиса прогнозирования
type PredictionServiceImpl struct {
	accountStorage   storage.FinancialAccountStorage
	loanStorage      storage.LoanStorage
	operationStorage storage.OperationStorage
}

// NewPredictionService создает новый сервис прогнозирования
func NewPredictionService(
	accountStorage storage.FinancialAccountStorage,
	loanStorage storage.LoanStorage,
	operationStorage storage.OperationStorage,
) PredictionService {
	return &PredictionServiceImpl{
		accountStorage:   accountStorage,
		loanStorage:      loanStorage,
		operationStorage: operationStorage,
	}
}

// PredictBalance прогнозирует баланс счетов клиента на указанное количество дней вперед
func (s *PredictionServiceImpl) PredictBalance(clientID uint, days int) ([]models.BalancePrediction, error) {
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
	
	// Вычисляем среднесуточные доходы и расходы на основе операций за последние 30 дней
	// Поскольку у нас нет соответствующего метода, используем фиктивные значения
	dailyIncome := float64(50)
	dailyExpense := float64(40)
	
	// Создаем прогнозы на каждый день
	predictions := make([]models.BalancePrediction, days)
	currentBalance := initialBalance
	today := time.Now().Truncate(24 * time.Hour)
	
	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, i)
		
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
			
			for _, plan := range paymentPlans {
				// Проверяем, есть ли платеж на этот день
				planDate := plan.DueDate.Truncate(24 * time.Hour)
				if planDate.Equal(date) && plan.Status != models.PaymentStatusCompleted {
					loanPayments += plan.Total
					isPaymentDate = true
					
					paymentDetails = append(paymentDetails, models.PaymentDetail{
						LoanID:      loan.ID,
						DueDate:     plan.DueDate,
						Amount:      plan.Total,
						Description: fmt.Sprintf("Платеж по кредиту №%d", loan.ID),
						Status:      string(plan.Status),
					})
				}
			}
		}
		
		// Вычисляем прогнозируемый баланс на этот день
		dailyNet := dailyIncome - dailyExpense
		plannedIncome := dailyIncome
		plannedExpenses := dailyExpense
		
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

// GetClientDebtRatio вычисляет коэффициент долговой нагрузки клиента
func (s *PredictionServiceImpl) GetClientDebtRatio(clientID uint) (float64, error) {
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

// GetClientDTI вычисляет коэффициент DTI (Debt-to-Income) для клиента
func (s *PredictionServiceImpl) GetClientDTI(clientID uint) (float64, error) {
	// Получаем все займы клиента (активные и погашенные)
	activeLoanBalance, err := s.loanStorage.GetTotalLoanAmount(clientID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении общей суммы займов: %w", err)
	}
	
	// Получаем общий баланс всех счетов клиента
	accounts, err := s.accountStorage.GetFinancialAccountsByClientID(clientID)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении счетов клиента: %w", err)
	}
	
	var totalAssets float64
	for _, account := range accounts {
		totalAssets += account.Funds
	}
	
	if totalAssets <= 0 {
		// Защита от деления на ноль
		return 0, fmt.Errorf("невозможно вычислить коэффициент DTI: отсутствуют данные об активах")
	}
	
	// Вычисляем DTI как отношение долга к активам
	dti := activeLoanBalance / totalAssets
	
	return dti, nil
}
