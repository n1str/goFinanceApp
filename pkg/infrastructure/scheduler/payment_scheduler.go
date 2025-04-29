package scheduler

import (
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/domain/models"
	"FinanceSystem/pkg/infrastructure/smtp"
	"log"
	"sync"
	"time"
)

// PaymentScheduler представляет планировщик для обработки платежей
type PaymentScheduler struct {
	loanService      services.LoanService
	accountService   services.FinancialAccountService
	operationService services.OperationService
	smtpService      *smtp.SMTPService
	interval         time.Duration
	running          bool
	stopCh           chan struct{}
	mutex            sync.Mutex
}

// NewPaymentScheduler создает новый планировщик платежей
func NewPaymentScheduler(
	loanService services.LoanService,
	accountService services.FinancialAccountService,
	operationService services.OperationService,
	smtpService *smtp.SMTPService,
	interval time.Duration,
) *PaymentScheduler {
	if interval == 0 {
		// По умолчанию запускаем каждые 12 часов
		interval = 12 * time.Hour
	}

	return &PaymentScheduler{
		loanService:      loanService,
		accountService:   accountService,
		operationService: operationService,
		smtpService:      smtpService,
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

// Start запускает планировщик
func (s *PaymentScheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		log.Println("Планировщик платежей уже запущен")
		return
	}

	s.running = true
	go s.run()
	log.Println("Планировщик платежей запущен, интервал:", s.interval)
}

// Stop останавливает планировщик
func (s *PaymentScheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		log.Println("Планировщик платежей уже остановлен")
		return
	}

	close(s.stopCh)
	s.running = false
	log.Println("Планировщик платежей остановлен")
}

// IsRunning возвращает состояние планировщика
func (s *PaymentScheduler) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.running
}

// ProcessPaymentsNow принудительно запускает обработку платежей
func (s *PaymentScheduler) ProcessPaymentsNow() error {
	return s.processPayments()
}

// run запускает основной цикл планировщика
func (s *PaymentScheduler) run() {
	// Сразу выполняем при первом запуске
	if err := s.processPayments(); err != nil {
		log.Printf("Ошибка при обработке платежей: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processPayments(); err != nil {
				log.Printf("Ошибка при обработке платежей: %v", err)
			}
		case <-s.stopCh:
			log.Println("Остановка цикла планировщика платежей")
			return
		}
	}
}

// processPayments обрабатывает все платежи
func (s *PaymentScheduler) processPayments() error {
	log.Println("Начало обработки платежей...")

	// 1. Получаем все активные займы
	loans, err := s.loanService.GetLoansByStatus(models.LoanStatusActive)
	if err != nil {
		return err
	}

	log.Printf("Найдено %d активных займов для обработки", len(loans))

	for _, loan := range loans {
		// 2. Для каждого займа получаем график платежей
		loanWithDetails, paymentPlans, err := s.loanService.GetLoanWithPaymentPlan(loan.ID)
		if err != nil {
			log.Printf("Ошибка при получении графика платежей для займа %d: %v", loan.ID, err)
			continue
		}

		// 3. Проверяем наличие просроченных и предстоящих платежей
		s.processLoanPayments(loanWithDetails, paymentPlans)
	}

	log.Println("Обработка платежей завершена")
	return nil
}

// processLoanPayments обрабатывает платежи по одному займу
func (s *PaymentScheduler) processLoanPayments(loan *models.Loan, payments []models.PaymentPlan) {
	// Текущая дата для сравнения
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Получаем email клиента для уведомлений
	clientEmail := loan.Client.Contact

	for i := range payments {
		payment := &payments[i]
		// Пропускаем уже оплаченные платежи
		if payment.Status == models.PaymentStatusCompleted {
			continue
		}

		// Дата платежа без времени для корректного сравнения
		paymentDate := time.Date(
			payment.DueDate.Year(),
			payment.DueDate.Month(),
			payment.DueDate.Day(),
			0, 0, 0, 0, time.UTC,
		)

		// Проверяем просроченные платежи
		if paymentDate.Before(today) {
			s.handleOverduePayment(loan, payment, clientEmail)
		} else if paymentDate.Equal(today) {
			// Платеж сегодня - пытаемся автоматически списать
			s.handleDuePayment(loan, payment, clientEmail)
		} else if paymentDate.Sub(today) <= 3*24*time.Hour {
			// Платеж в ближайшие 3 дня - отправляем уведомление
			s.sendPaymentReminder(loan, payment, clientEmail)
		}
	}
}

// handleOverduePayment обрабатывает просроченный платеж
func (s *PaymentScheduler) handleOverduePayment(loan *models.Loan, payment *models.PaymentPlan, email string) {
	log.Printf("Обработка просроченного платежа для займа %d (план %d)", loan.ID, payment.ID)

	// Если платеж уже помечен как просроченный, пропускаем
	if payment.Status == models.PaymentStatusDelayed {
		return
	}

	// Начисляем штраф (10% от суммы платежа)
	penaltyAmount := payment.Total * 0.1
	log.Printf("Начисление штрафа %.2f руб. для займа %d (план %d)", 
		penaltyAmount, loan.ID, payment.ID)

	// Увеличиваем сумму платежа на размер штрафа
	payment.Total += penaltyAmount
	payment.Status = models.PaymentStatusDelayed

	// Обновляем запись в базе данных
	// Этот метод нужно реализовать в LoanService
	if err := s.loanService.UpdatePaymentPlan(payment); err != nil {
		log.Printf("Ошибка при обновлении плана платежей %d: %v", payment.ID, err)
		return
	}

	// Отправляем уведомление о просроченном платеже
	if s.smtpService != nil {
		err := s.smtpService.SendOverduePaymentNotification(
			email, 
			loan.ID, 
			payment.Total - penaltyAmount, 
			payment.DueDate, 
			penaltyAmount,
		)
		if err != nil {
			log.Printf("Ошибка при отправке уведомления о просрочке: %v", err)
		}
	}
}

// handleDuePayment обрабатывает платеж, срок которого наступил сегодня
func (s *PaymentScheduler) handleDuePayment(loan *models.Loan, payment *models.PaymentPlan, email string) {
	log.Printf("Обработка текущего платежа для займа %d (план %d)", loan.ID, payment.ID)

	// Проверяем наличие средств на счете
	account, err := s.accountService.GetAccountByID(loan.AccountID)
	if err != nil {
		log.Printf("Ошибка при получении счета %d: %v", loan.AccountID, err)
		return
	}

	// Если на счете достаточно средств, выполняем платеж
	if account.Funds >= payment.Total {
		log.Printf("Автоматическое списание платежа %.2f руб. для займа %d", 
			payment.Total, loan.ID)

		// Создаем операцию списания
		err = s.operationService.CreateOperation(
			loan.AccountID,
			-payment.Total,
			"Автоматическое списание по кредиту",
			"LOAN_PAYMENT",
		)
		if err != nil {
			log.Printf("Ошибка при создании операции: %v", err)
			return
		}

		// Обновляем статус платежа
		payment.Status = models.PaymentStatusCompleted
		payment.PaidDate = &time.Time{}
		*payment.PaidDate = time.Now().UTC()

		// Обновляем запись в базе данных
		if err := s.loanService.UpdatePaymentPlan(payment); err != nil {
			log.Printf("Ошибка при обновлении плана платежей %d: %v", payment.ID, err)
			return
		}

		// Отправляем уведомление об успешном платеже
		if s.smtpService != nil {
			err := s.smtpService.SendPaymentNotification(
				email,
				payment.Total,
				*payment.PaidDate,
				loan.ID,
				"completed",
			)
			if err != nil {
				log.Printf("Ошибка при отправке уведомления об оплате: %v", err)
			}
		}
	} else {
		// Недостаточно средств - отправляем уведомление
		log.Printf("Недостаточно средств на счете %d для платежа по займу %d", 
			loan.AccountID, loan.ID)
		
		if s.smtpService != nil {
			err := s.smtpService.SendPaymentNotification(
				email,
				payment.Total,
				payment.DueDate,
				loan.ID,
				"pending",
			)
			if err != nil {
				log.Printf("Ошибка при отправке уведомления о необходимости оплаты: %v", err)
			}
		}
	}
}

// sendPaymentReminder отправляет напоминание о предстоящем платеже
func (s *PaymentScheduler) sendPaymentReminder(loan *models.Loan, payment *models.PaymentPlan, email string) {
	// Отправляем только если есть сервис SMTP
	if s.smtpService == nil {
		return
	}

	log.Printf("Отправка напоминания о платеже для займа %d (план %d), дата: %s", 
		loan.ID, payment.ID, payment.DueDate.Format("02.01.2006"))

	err := s.smtpService.SendPaymentNotification(
		email,
		payment.Total,
		payment.DueDate,
		loan.ID,
		"pending",
	)
	if err != nil {
		log.Printf("Ошибка при отправке напоминания о платеже: %v", err)
	}
}
