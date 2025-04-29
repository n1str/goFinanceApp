package handlers

import (
	"FinanceSystem/pkg/adapters/storage"
	"FinanceSystem/pkg/infrastructure/security"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Константы путей API
const (
	APIPathRegister      = "/register"
	APIPathLogin         = "/login"
	APIPathAuthStatus    = "/status"
	APIPathProfile       = "/profile"
	APIPathAccounts      = "/accounts"
	APIPathCards         = "/cards"
	APIPathOperations    = "/operations"
	APIPathTransfer      = "/transfer"
	APIPathDeposit       = "/deposit"
	APIPathWithdraw      = "/withdraw"
	APIPathLoans         = "/loans"
	APIPathPayments      = "/payments"
	APIPathAnalytics     = "/analytics"
	APIPathForecast      = "/forecast"
	APIPathSummary       = "/summary"
	APIPathSpending      = "/spending"
	APIPathKeyRate       = "/key-rate"
	APIPathCurrent       = "/current"
	APIPathHistory       = "/history"
	APIPathCurrency      = "/currency"
)

// Router управляет маршрутизацией запросов API
type Router struct {
	authHandler        *AuthHandler
	accountHandler     *FinancialAccountHandler
	cardHandler        *PaymentCardHandler
	loanHandler        *LoanHandler
	analyticsHandler   *AnalyticsHandler
	externalDataHandler *ExternalDataHandler
	clientStorage      storage.ClientStorage
	predictionHandler  *PredictionHandler
}

// NewRouter создает новый экземпляр маршрутизатора
func NewRouter(
	authHandler *AuthHandler,
	accountHandler *FinancialAccountHandler,
	cardHandler *PaymentCardHandler,
	loanHandler *LoanHandler,
	analyticsHandler *AnalyticsHandler,
	externalDataHandler *ExternalDataHandler,
	clientStorage storage.ClientStorage,
	predictionHandler *PredictionHandler,
) *Router {
	return &Router{
		authHandler:        authHandler,
		accountHandler:     accountHandler,
		cardHandler:        cardHandler,
		loanHandler:        loanHandler,
		analyticsHandler:   analyticsHandler,
		externalDataHandler: externalDataHandler,
		clientStorage:      clientStorage,
		predictionHandler:  predictionHandler,
	}
}

// SetupRouter настраивает и возвращает роутер Gin с зарегистрированными маршрутами
func (r *Router) SetupRouter() *gin.Engine {
	router := gin.Default()

	// Применяем глобальные middleware
	router.Use(security.LoggerMiddleware())
	router.Use(security.CorsMiddleware())
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(security.RateLimitMiddleware(100, time.Minute))

	// Регистрация маршрутов
	r.registerAuthRoutes(router.Group("/api/auth"))
	r.registerAccountRoutes(router.Group("/api/accounts"))
	r.registerCardRoutes(router.Group("/api/cards"))
	r.registerLoanRoutes(router.Group("/api/loans"))
	r.registerAnalyticsRoutes(router.Group("/api/analytics"))
	r.registerExternalDataRoutes(router.Group("/api/external"))
	r.registerPredictionRoutes(router.Group("/api/predictions"))

	return router
}

// registerAuthRoutes регистрирует маршруты аутентификации
func (r *Router) registerAuthRoutes(group *gin.RouterGroup) {
	group.POST(APIPathRegister, r.authHandler.Register)
	group.POST(APIPathLogin, r.authHandler.Login)
	
	// Маршруты, требующие аутентификации
	authGroup := group.Group("/")
	authGroup.Use(security.AuthMiddleware(r.clientStorage))
	
	authGroup.GET(APIPathAuthStatus, r.authHandler.CheckAuthStatus)
	authGroup.GET(APIPathProfile, r.authHandler.GetUserProfile)
}

// registerAccountRoutes регистрирует маршруты для работы со счетами
func (r *Router) registerAccountRoutes(group *gin.RouterGroup) {
	// Маршруты, требующие аутентификации
	group.Use(security.AuthMiddleware(r.clientStorage))
	
	// Получение списка счетов клиента
	group.GET("", r.accountHandler.GetMyAccounts)
	
	// Создание нового счета
	group.POST("", r.accountHandler.CreateAccount)
	
	// Получение счета по ID
	group.GET("/:id", r.accountHandler.GetAccountByID)
	
	// Операции со счетом
	group.POST("/:id"+APIPathDeposit, r.accountHandler.AddFunds)
	group.POST("/:id"+APIPathWithdraw, r.accountHandler.WithdrawFunds)
	group.POST(APIPathTransfer, r.accountHandler.TransferFunds)
	
	// Получение операций по счету
	group.GET("/:id"+APIPathOperations, r.accountHandler.GetAccountOperations)
	
	// Маршруты для администраторов
	adminGroup := group.Group("/admin")
	adminGroup.Use(security.AdminMiddleware())
	
	adminGroup.GET("", r.accountHandler.GetAllAccounts)
}

// registerCardRoutes регистрирует маршруты для работы с картами
func (r *Router) registerCardRoutes(group *gin.RouterGroup) {
	// Маршруты, требующие аутентификации
	group.Use(security.AuthMiddleware(r.clientStorage))
	
	// Создание новой карты
	group.POST("", r.cardHandler.CreateCard)
	
	// Получение всех карт клиента
	group.GET("", r.cardHandler.GetClientCards)
	
	// Получение карты по ID
	group.GET("/:id", r.cardHandler.GetCardByID)
	
	// Операции с картой
	group.POST("/:id/block", r.cardHandler.BlockCard)
	group.POST("/:id/unblock", r.cardHandler.UnblockCard)
	
	// Валидация карты
	group.POST("/validate", r.cardHandler.ValidateCard)
	
	// Получение карт по ID счета
	group.GET("/account/:accountId", r.cardHandler.GetCardsByAccountID)
}

// registerLoanRoutes регистрирует маршруты для работы с займами
func (r *Router) registerLoanRoutes(api *gin.RouterGroup) {
	// Публичные маршруты
	api.GET("", r.loanHandler.GetAllLoans)
	api.GET("/:id", r.loanHandler.GetLoanByID)
	api.POST("/calculate", r.loanHandler.CalculateLoanDetails)
	
	// Защищенные маршруты
	authorized := api.Group("", security.AuthMiddleware(r.clientStorage))
	{
		authorized.GET("/client/:id", r.loanHandler.GetLoansByClientID)
		authorized.POST("", r.loanHandler.CreateLoan)
		authorized.GET("/:id/payment-plan", r.loanHandler.GetLoanPaymentPlan)
		authorized.POST("/:id/payment", r.loanHandler.MakePayment)
		authorized.GET("/status/:status", r.loanHandler.GetLoansByStatus)
		authorized.PUT("/update-payments", r.loanHandler.UpdateAllLoanPayments)
	}
}

// registerAnalyticsRoutes регистрирует маршруты для аналитики
func (r *Router) registerAnalyticsRoutes(group *gin.RouterGroup) {
	// Маршруты, требующие аутентификации
	group.Use(security.AuthMiddleware(r.clientStorage))
	
	// Получение аналитики счетов
	group.GET("/accounts", r.analyticsHandler.GetAccountsAnalytics)
	
	// Получение общей информации о финансах
	group.GET(APIPathSummary, r.analyticsHandler.GetClientSummary)
	
	// Получение отчета за период
	group.POST("/report", r.analyticsHandler.GetPeriodReport)
	
	// Получение анализа расходов
	group.GET(APIPathSpending, r.analyticsHandler.GetSpendingAnalytics)
	
	// Получение финансового прогноза
	group.GET(APIPathForecast, r.analyticsHandler.GetFinancialForecast)
}

// registerExternalDataRoutes регистрирует маршруты для внешних данных
func (r *Router) registerExternalDataRoutes(group *gin.RouterGroup) {
	// Маршруты, требующие аутентификации
	group.Use(security.AuthMiddleware(r.clientStorage))
	
	// Получение текущей ключевой ставки
	group.GET("/key-rate", r.externalDataHandler.GetCurrentKeyRate)
	
	// Получение истории ключевой ставки
	group.POST("/key-rate/history", r.externalDataHandler.GetKeyRateHistory)
	
	// Получение курса валюты
	group.GET("/currency/:code", r.externalDataHandler.GetCurrencyRate)
}

// registerPredictionRoutes регистрирует маршруты для прогнозирования
func (r *Router) registerPredictionRoutes(group *gin.RouterGroup) {
	// Маршруты, требующие аутентификации
	group.Use(security.AuthMiddleware(r.clientStorage))
	
	// Роуты для прогнозов и аналитики
	group.GET("/clients/:client_id/predict-balance", r.predictionHandler.PredictBalance)
	group.GET("/clients/:client_id/debt-ratio", r.predictionHandler.GetClientDebtRatio)
}
