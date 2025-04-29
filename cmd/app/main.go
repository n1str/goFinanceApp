package main

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/application/services"
	"FinanceSystem/pkg/infrastructure/db"
	"FinanceSystem/pkg/interfaces/handlers"
	"log"
	"time"
)

func main() {
	// Инициализация базы данных
	db.InitDatabase()
	db.CreateTables(db.Database)

	// Создание репозиториев
	clientStorage := storage.NewClientStorage(db.Database)
	accessStorage := storage.NewAccessStorage(db.Database)
	accountStorage := storage.NewFinancialAccountStorage(db.Database)
	operationStorage := storage.NewOperationStorage(db.Database)
	cardStorage := storage.NewPaymentCardStorage(db.Database)
	loanStorage := storage.NewLoanStorage(db.Database)

	// Создание сервисов
	accessService := services.NewAccessService(accessStorage)
	
	// Инициализация стандартных прав доступа
	if err := accessService.InitializeDefaultAccess(); err != nil {
		log.Fatalf("Ошибка при инициализации прав доступа: %v", err)
	}
	
	// Другие сервисы
	authService := services.NewAuthService(clientStorage, accessStorage, "fM7NVJqxErP3GYzH5tLW9FdZ2cRbTgKj", 24*time.Hour)
	accountService := services.NewFinancialAccountService(accountStorage, operationStorage)
	
	// Ключ и соль для шифрования данных карты
	encryptionKey := "MyEncryptionKey123"
	saltBytes := []byte{12, 34, 56, 78, 90}
	
	cardService := services.NewPaymentCardService(cardStorage, accountStorage, encryptionKey, saltBytes)
	
	// Создание внешнего сервиса для интеграций
	externalService := services.NewExternalService(
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		10*time.Second,
		"FinanceSystem/1.0",
		"FinanceSystem Client",
		false, // Используем реальное API ЦБ РФ вместо демо-данных
	)
	
	loanService := services.NewLoanService(loanStorage, accountStorage, operationStorage, externalService)
	analyticsService := services.NewAnalyticsService(operationStorage, accountStorage, loanStorage)
	predictionService := services.NewPredictionService(accountStorage, loanStorage, operationStorage)

	// Создание обработчиков HTTP запросов
	authHandler := handlers.NewAuthHandler(authService)
	accountHandler := handlers.NewFinancialAccountHandler(accountService)
	cardHandler := handlers.NewPaymentCardHandler(cardService, accountService)
	loanHandler := handlers.NewLoanHandler(loanService, accountService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	externalDataHandler := handlers.NewExternalDataHandler(externalService)
	predictionHandler := handlers.NewPredictionHandler(predictionService, loanService)

	// Создание маршрутизатора
	router := handlers.NewRouter(
		authHandler,
		accountHandler,
		cardHandler,
		loanHandler,
		analyticsHandler,
		externalDataHandler,
		clientStorage,
		predictionHandler,
	)

	// Настройка и запуск веб-сервера
	r := router.SetupRouter()
	
	log.Println("Запуск сервера на порту 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
