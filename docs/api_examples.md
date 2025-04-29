# Примеры API запросов для Финансовой Системы

В этом документе приведены примеры HTTP запросов к API Финансовой Системы для различных функций.

## Аутентификация

### Регистрация нового клиента

```http
POST /api/auth/register
Content-Type: application/json

{
  "fullName": "Иван Петров",
  "loginName": "ivan_petrov",
  "contact": "ivan@example.com",
  "password": "secure_password123"
}
```

### Вход в систему

```http
POST /api/auth/login
Content-Type: application/json

{
  "loginName": "ivan_petrov",
  "password": "secure_password123"
}
```

Пример ответа:
```json
{
  "message": "Вход выполнен успешно",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "client": {
    "id": 1,
    "fullName": "Иван Петров",
    "loginName": "ivan_petrov",
    "contact": "ivan@example.com"
  }
}
```

### Проверка статуса аутентификации

```http
GET /api/auth/status
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Финансовые счета

### Создание нового счета

```http
POST /api/accounts
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "title": "Основной счет",
  "initialFunds": 5000.00
}
```

### Получение списка счетов

```http
GET /api/accounts
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Пополнение счета

```http
POST /api/accounts/1/deposit
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "amount": 1000.00,
  "details": "Зачисление зарплаты"
}
```

### Перевод между счетами

```http
POST /api/accounts/1/transfer
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "toAccountId": 2,
  "amount": 500.00,
  "details": "Перевод на сберегательный счет"
}
```

## Платежные карты

### Создание новой карты

```http
POST /api/cards
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "accountId": 1,
  "cardholderName": "IVAN PETROV"
}
```

### Получение карт по счету

```http
GET /api/cards/account/1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Блокировка карты

```http
POST /api/cards/1/block
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Займы

### Расчет параметров займа

```http
POST /api/loans/calculate
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "principal": 100000.00,
  "interestRate": 12.5,
  "durationMonths": 24
}
```

### Оформление займа

```http
POST /api/loans
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "accountId": 1,
  "principal": 100000.00,
  "interestRate": 12.5,
  "durationMonths": 24,
  "purpose": "Ремонт квартиры"
}
```

### Получение графика платежей

```http
GET /api/loans/1/payments
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Совершение платежа по займу

```http
POST /api/loans/1/payments
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "paymentPlanId": 3
}
```

## Аналитика

### Получение общей информации о финансах

```http
GET /api/analytics/summary
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Получение отчета за период

```http
POST /api/analytics/report
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "startDate": "2025-01-01",
  "endDate": "2025-03-31"
}
```

### Получение финансового прогноза

```http
GET /api/analytics/forecast?months=12
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Внешние данные

### Получение текущей ключевой ставки

```http
GET /api/external/key-rate/current
```

### Получение курса валюты

```http
GET /api/external/key-rate/currency/USD
```
