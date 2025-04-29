# Финансовая Система

Финансовая система - это современное приложение для управления личными финансами, включающее функционал для работы со счетами, картами, займами и аналитику финансов.

## Основные возможности

- Регистрация и авторизация пользователей
- Управление личными финансовыми счетами
- Создание и управление платежными картами
- Оформление и погашение займов
- Выполнение денежных операций (пополнение, снятие, переводы)
- Аналитика финансов и прогнозирование
- Получение информации о ключевой ставке и курсах валют

## Структура проекта

```
FinanceSystem/
├── cmd/
│   └── app/
│       └── main.go             # Точка входа в приложение
├── pkg/
│   ├── adapters/
│   │   └── storage/            # Репозитории для работы с базой данных
│   ├── application/
│   │   └── services/           # Сервисы бизнес-логики
│   ├── domain/
│   │   └── models/             # Модели предметной области
│   ├── infrastructure/
│   │   ├── db/                 # Работа с базой данных
│   │   └── security/           # Безопасность и аутентификация
│   ├── interfaces/
│   │   └── handlers/           # Обработчики HTTP-запросов
│   └── presentation/
│       └── dto/                # Объекты передачи данных
├── database_config.json        # Конфигурация базы данных
├── go.mod                      # Зависимости проекта
└── README.md                   # Документация проекта
```

## Запуск проекта

1. Клонируйте репозиторий:
```bash
git clone https://github.com/username/FinanceSystem.git
cd FinanceSystem
```

2. Убедитесь, что у вас установлен Go версии 1.18 или выше:
```bash
go version
```

3. Соберите и запустите приложение:
```bash
go run cmd/app/main.go
```

4. Откройте в браузере: http://localhost:8080

## API Эндпоинты

### Аутентификация
- `POST /api/auth/register` - Регистрация нового клиента
- `POST /api/auth/login` - Вход в систему
- `GET /api/auth/status` - Проверка статуса аутентификации
- `GET /api/auth/profile` - Получение профиля клиента

### Счета
- `GET /api/accounts` - Получение списка счетов клиента
- `POST /api/accounts` - Создание нового счета
- `GET /api/accounts/:id` - Получение счета по ID
- `POST /api/accounts/:id/deposit` - Пополнение счета
- `POST /api/accounts/:id/withdraw` - Снятие средств со счета
- `POST /api/accounts/:id/transfer` - Перевод средств между счетами
- `GET /api/accounts/:id/operations` - Получение операций по счету

### Карты
- `POST /api/cards` - Создание новой карты
- `GET /api/cards/:id` - Получение карты по ID
- `POST /api/cards/:id/block` - Блокировка карты
- `POST /api/cards/:id/unblock` - Разблокировка карты
- `POST /api/cards/validate` - Валидация карты
- `GET /api/cards/account/:accountId` - Получение карт по ID счета

### Займы
- `GET /api/loans` - Получение списка займов клиента
- `POST /api/loans` - Создание нового займа
- `GET /api/loans/:id` - Получение займа по ID
- `GET /api/loans/:id/payments` - Получение графика платежей
- `POST /api/loans/:id/payments` - Совершение платежа
- `POST /api/loans/calculate` - Расчет параметров займа

### Аналитика
- `GET /api/analytics/summary` - Получение общей информации о финансах
- `POST /api/analytics/report` - Получение отчета за период
- `GET /api/analytics/spending` - Получение анализа расходов
- `GET /api/analytics/forecast` - Получение финансового прогноза

### Внешние данные
- `GET /api/external/key-rate/current` - Получение текущей ключевой ставки
- `GET /api/external/key-rate/history` - Получение истории ключевой ставки
- `GET /api/external/key-rate/currency/:code` - Получение курса валюты

## Лицензия

MIT
