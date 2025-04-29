package smtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"gopkg.in/gomail.v2"
)

// SMTPConfig содержит конфигурацию SMTP-сервера
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	FromName string
}

// DefaultSMTPConfig возвращает конфигурацию SMTP по умолчанию
func DefaultSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		User:     "noreply@yourbank.com",
		Password: "your_smtp_password",
		FromName: "Finance System",
	}
}

// SMTPService предоставляет функциональность для отправки электронных писем
type SMTPService struct {
	config SMTPConfig
	dialer *gomail.Dialer
}

// NewSMTPService создает новый сервис для отправки электронных писем
func NewSMTPService(config SMTPConfig) *SMTPService {
	dialer := gomail.NewDialer(config.Host, config.Port, config.User, config.Password)
	dialer.TLSConfig = &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false,
	}

	return &SMTPService{
		config: config,
		dialer: dialer,
	}
}

// SendEmail отправляет электронное письмо
func (s *SMTPService) SendEmail(to string, subject string, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.User))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	// Попытка отправить письмо
	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("Ошибка при отправке email: %v", err)
		return fmt.Errorf("не удалось отправить email: %w", err)
	}

	log.Printf("Email успешно отправлен на адрес %s", to)
	return nil
}

// SendPaymentNotification отправляет уведомление о платеже
func (s *SMTPService) SendPaymentNotification(to string, amount float64, date time.Time, 
	loanID uint, status string) error {
	
	subject := "Уведомление о платеже по кредиту"
	
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; }
				.container { width: 100%%; max-width: 600px; margin: 0 auto; }
				.header { background-color: #3498db; color: white; padding: 10px; text-align: center; }
				.content { padding: 20px; }
				.footer { background-color: #f8f8f8; padding: 10px; text-align: center; font-size: 12px; }
				.info { border: 1px solid #ddd; padding: 10px; margin-bottom: 20px; }
				.success { color: #27ae60; }
				.warning { color: #e74c3c; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Уведомление о платеже</h2>
				</div>
				<div class="content">
					<p>Уважаемый клиент,</p>
					<p>Информируем вас о платеже по вашему кредиту:</p>
					
					<div class="info">
						<p><strong>Кредит №:</strong> %d</p>
						<p><strong>Сумма платежа:</strong> %.2f руб.</p>
						<p><strong>Дата платежа:</strong> %s</p>
						<p><strong>Статус:</strong> <span class="%s">%s</span></p>
					</div>
					
					<p>Если у вас возникли вопросы, пожалуйста, свяжитесь с нашей службой поддержки.</p>
					<p>С уважением,<br>Команда Finance System</p>
				</div>
				<div class="footer">
					<p>Это автоматическое уведомление. Пожалуйста, не отвечайте на него.</p>
					<p>&copy; %d Finance System. Все права защищены.</p>
				</div>
			</div>
		</body>
		</html>
	`, loanID, amount, date.Format("02.01.2006"), 
		status == "completed" ? "success" : "warning", 
		status == "completed" ? "Оплачен" : "Ожидает оплаты",
		time.Now().Year())
	
	return s.SendEmail(to, subject, htmlBody)
}

// SendLoanApprovalNotification отправляет уведомление об одобрении кредита
func (s *SMTPService) SendLoanApprovalNotification(to string, loanID uint, amount float64, 
	term int, monthlyPayment float64) error {
	
	subject := "Ваш кредит одобрен!"
	
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; }
				.container { width: 100%%; max-width: 600px; margin: 0 auto; }
				.header { background-color: #27ae60; color: white; padding: 10px; text-align: center; }
				.content { padding: 20px; }
				.footer { background-color: #f8f8f8; padding: 10px; text-align: center; font-size: 12px; }
				.info { border: 1px solid #ddd; padding: 10px; margin-bottom: 20px; }
				table { width: 100%%; border-collapse: collapse; }
				table, th, td { border: 1px solid #ddd; }
				th, td { padding: 8px; text-align: left; }
				th { background-color: #f2f2f2; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Кредит одобрен</h2>
				</div>
				<div class="content">
					<p>Уважаемый клиент,</p>
					<p>Мы рады сообщить, что ваша заявка на кредит одобрена!</p>
					
					<div class="info">
						<p><strong>Кредит №:</strong> %d</p>
						<p><strong>Сумма кредита:</strong> %.2f руб.</p>
						<p><strong>Срок кредита:</strong> %d месяцев</p>
						<p><strong>Ежемесячный платеж:</strong> %.2f руб.</p>
					</div>
					
					<p>Детальный график платежей доступен в вашем личном кабинете.</p>
					<p>С уважением,<br>Команда Finance System</p>
				</div>
				<div class="footer">
					<p>Это автоматическое уведомление. Пожалуйста, не отвечайте на него.</p>
					<p>&copy; %d Finance System. Все права защищены.</p>
				</div>
			</div>
		</body>
		</html>
	`, loanID, amount, term, monthlyPayment, time.Now().Year())
	
	return s.SendEmail(to, subject, htmlBody)
}

// SendOverduePaymentNotification отправляет уведомление о просроченном платеже
func (s *SMTPService) SendOverduePaymentNotification(to string, loanID uint, amount float64, 
	dueDate time.Time, penalty float64) error {
	
	subject := "ВНИМАНИЕ: Просроченный платеж по кредиту"
	
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; }
				.container { width: 100%%; max-width: 600px; margin: 0 auto; }
				.header { background-color: #e74c3c; color: white; padding: 10px; text-align: center; }
				.content { padding: 20px; }
				.footer { background-color: #f8f8f8; padding: 10px; text-align: center; font-size: 12px; }
				.info { border: 1px solid #ddd; padding: 10px; margin-bottom: 20px; }
				.warning { color: #e74c3c; font-weight: bold; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Просроченный платеж</h2>
				</div>
				<div class="content">
					<p>Уважаемый клиент,</p>
					<p class="warning">Обращаем ваше внимание, что платеж по кредиту просрочен!</p>
					
					<div class="info">
						<p><strong>Кредит №:</strong> %d</p>
						<p><strong>Сумма платежа:</strong> %.2f руб.</p>
						<p><strong>Дата платежа:</strong> %s</p>
						<p><strong>Начисленный штраф:</strong> %.2f руб.</p>
						<p><strong>Итого к оплате:</strong> %.2f руб.</p>
					</div>
					
					<p>Просим вас погасить задолженность в ближайшее время во избежание дальнейшего начисления штрафов.</p>
					<p>С уважением,<br>Команда Finance System</p>
				</div>
				<div class="footer">
					<p>Это автоматическое уведомление. Пожалуйста, не отвечайте на него.</p>
					<p>&copy; %d Finance System. Все права защищены.</p>
				</div>
			</div>
		</body>
		</html>
	`, loanID, amount, dueDate.Format("02.01.2006"), penalty, amount + penalty, time.Now().Year())
	
	return s.SendEmail(to, subject, htmlBody)
}
