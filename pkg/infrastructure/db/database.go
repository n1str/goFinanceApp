package db

import (
	"FinanceSystem/pkg/domain/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite" // Используем pure Go реализацию SQLite
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig содержит конфигурацию базы данных
type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// Database глобальная переменная для доступа к базе данных
var Database *gorm.DB

// InitDatabase инициализирует соединение с базой данных
func InitDatabase() {
	var err error
	config, err := loadDatabaseConfig()
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации базы данных: %v. Используем SQLite по умолчанию.", err)
		initSQLiteDatabase()
		return
	}

	switch config.Type {
	case "sqlite":
		initSQLiteDatabase()
	default:
		log.Printf("Неизвестный тип базы данных: %s. Используем SQLite по умолчанию.", config.Type)
		initSQLiteDatabase()
	}
}

// loadDatabaseConfig загружает конфигурацию базы данных из файла
func loadDatabaseConfig() (*DatabaseConfig, error) {
	file, err := os.Open("database_config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config DatabaseConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// initSQLiteDatabase инициализирует базу данных SQLite
func initSQLiteDatabase() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	Database, err = gorm.Open(sqlite.Open("finance_system.db"), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}

	log.Println("Успешное подключение к базе данных SQLite")
}

// CreateTables создает необходимые таблицы в базе данных
func CreateTables(db *gorm.DB) {
	// Список всех моделей, которые нужно создать в базе данных
	models := []interface{}{
		&models.Client{},
		&models.Access{},
		&models.FinancialAccount{},
		&models.PaymentCard{},
		&models.Operation{},
		&models.Loan{},
		&models.PaymentPlan{},
	}

	// Создаем таблицы для всех моделей
	for _, model := range models {
		err := db.AutoMigrate(model)
		if err != nil {
			log.Fatalf("Ошибка при создании таблицы для модели %T: %v", model, err)
		}
	}

	fmt.Println("Все таблицы успешно созданы")
}
