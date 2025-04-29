#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import os
import json
import random
import string
import requests
import logging
import time
import datetime
import argparse
import traceback
import uuid
from dataclasses import dataclass
from typing import List, Dict, Any, Optional, Tuple, Union

# Пытаемся импортировать reportlab вместо fpdf
try:
    from reportlab.lib.pagesizes import A4
    from reportlab.lib import colors
    from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer
    from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
    from reportlab.pdfbase import pdfmetrics
    from reportlab.pdfbase.ttfonts import TTFont
except ImportError:
    print("Установка необходимых библиотек...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "reportlab", "--break-system-packages"])
    from reportlab.lib.pagesizes import A4
    from reportlab.lib import colors
    from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer
    from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
    from reportlab.pdfbase import pdfmetrics
    from reportlab.pdfbase.ttfonts import TTFont

@dataclass
class TestResult:
    """Класс для хранения результатов теста"""
    name: str
    url: str
    method: str
    request_data: Union[Dict[str, Any], str]
    response_data: Union[Dict[str, Any], str]
    status_code: int
    execution_time: float
    passed: bool
    error: Optional[str] = None


class APITestReport:
    """Класс для генерации отчета о тестировании API"""
    
    def __init__(self, results: List[TestResult], total_time: float):
        """
        Инициализация отчета
        
        Args:
            results: Список результатов тестов
            total_time: Общее время выполнения тестов
        """
        self.results = results
        self.total_time = total_time
        self.passed_count = sum(1 for r in results if r.passed)
        self.failed_count = len(results) - self.passed_count
        
        # Настройка шрифтов для поддержки кириллицы
        try:
            # Пытаемся зарегистрировать шрифт DejaVuSans
            pdfmetrics.registerFont(TTFont('DejaVuSans', '/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf'))
            self.font = 'DejaVuSans'
        except:
            # В случае ошибки используем встроенный шрифт
            self.font = 'Helvetica'
        
        # Определяем стили для отчёта
        self.styles = getSampleStyleSheet()
        self.styles.add(ParagraphStyle(name='Russian', 
                                      fontName=self.font, 
                                      fontSize=10, 
                                      leading=12,
                                      alignment=0))
        self.styles.add(ParagraphStyle(name='RussianBold', 
                                      fontName=self.font, 
                                      fontSize=12, 
                                      leading=14,
                                      alignment=1,
                                      bold=True))
        self.styles.add(ParagraphStyle(name='RussianTitle', 
                                      fontName=self.font, 
                                      fontSize=16, 
                                      leading=18,
                                      alignment=1,
                                      bold=True))
    
    def _truncate_data(self, data: Union[Dict[str, Any], str], max_length: int = 500) -> str:
        """
        Обрезает данные до определенной длины для отображения в отчете
        
        Args:
            data: Данные для обрезки
            max_length: Максимальная длина строки
            
        Returns:
            Строка с обрезанными данными
        """
        if isinstance(data, dict):
            data_str = json.dumps(data, ensure_ascii=False, indent=2)
        else:
            data_str = str(data)
            
        if len(data_str) > max_length:
            return data_str[:max_length] + "..."
        return data_str
    
    def generate_text_report(self, filename: str) -> None:
        """
        Генерирует текстовый отчет
        
        Args:
            filename: Имя файла для сохранения отчета
        """
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(f"========== ОТЧЕТ О ТЕСТИРОВАНИИ API ==========\n")
            f.write(f"Дата: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
            f.write(f"Всего тестов: {len(self.results)}\n")
            f.write(f"Успешных тестов: {self.passed_count}\n")
            f.write(f"Неуспешных тестов: {self.failed_count}\n")
            f.write(f"Общее время выполнения: {self.total_time:.2f} сек\n")
            f.write(f"==============================================\n\n")
            
            for i, result in enumerate(self.results, 1):
                f.write(f"Тест #{i}: {result.name}\n")
                f.write(f"URL: {result.url}\n")
                f.write(f"Метод: {result.method}\n")
                f.write(f"Статус: {'УСПЕХ' if result.passed else 'НЕУДАЧА'}\n")
                f.write(f"Код ответа: {result.status_code}\n")
                f.write(f"Время выполнения: {result.execution_time:.2f} сек\n")
                
                f.write(f"Запрос:\n{self._truncate_data(result.request_data)}\n\n")
                f.write(f"Ответ:\n{self._truncate_data(result.response_data)}\n\n")
                
                if result.error:
                    f.write(f"Ошибка: {result.error}\n")
                
                f.write(f"----------------------------------------------\n\n")
    
    def output(self, filename: str) -> None:
        """
        Создает PDF отчет
        
        Args:
            filename: Имя файла для сохранения отчета
        """
        # Создаем текстовый отчет
        text_filename = filename.replace('.pdf', '.txt')
        self.generate_text_report(text_filename)
        
        # Создаем PDF документ
        doc = SimpleDocTemplate(filename, pagesize=A4)
        elements = []
        
        # Заголовок отчета
        elements.append(Paragraph("ОТЧЕТ О ТЕСТИРОВАНИИ API", self.styles['RussianTitle']))
        elements.append(Spacer(1, 12))
        
        # Общая информация
        elements.append(Paragraph(f"Дата: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}", self.styles['Russian']))
        elements.append(Paragraph(f"Всего тестов: {len(self.results)}", self.styles['Russian']))
        elements.append(Paragraph(f"Успешных тестов: {self.passed_count}", self.styles['Russian']))
        elements.append(Paragraph(f"Неуспешных тестов: {self.failed_count}", self.styles['Russian']))
        elements.append(Paragraph(f"Общее время выполнения: {self.total_time:.2f} сек", self.styles['Russian']))
        elements.append(Spacer(1, 20))
        
        # Результаты тестов
        for i, result in enumerate(self.results, 1):
            elements.append(Paragraph(f"Тест #{i}: {result.name}", self.styles['RussianBold']))
            elements.append(Spacer(1, 6))
            
            # Основная информация о тесте
            data = [
                ["URL:", result.url],
                ["Метод:", result.method],
                ["Статус:", "УСПЕХ" if result.passed else "НЕУДАЧА"],
                ["Код ответа:", str(result.status_code)],
                ["Время выполнения:", f"{result.execution_time:.2f} сек"]
            ]
            
            t = Table(data, colWidths=[100, 400])
            t.setStyle(TableStyle([
                ('FONTNAME', (0, 0), (-1, -1), self.font),
                ('GRID', (0, 0), (-1, -1), 0.25, colors.grey),
                ('BACKGROUND', (0, 0), (0, -1), colors.lightgrey),
                ('VALIGN', (0, 0), (-1, -1), 'MIDDLE'),
            ]))
            elements.append(t)
            elements.append(Spacer(1, 10))
            
            # Запрос
            elements.append(Paragraph("Запрос:", self.styles['RussianBold']))
            elements.append(Paragraph(self._truncate_data(result.request_data), self.styles['Russian']))
            elements.append(Spacer(1, 10))
            
            # Ответ
            elements.append(Paragraph("Ответ:", self.styles['RussianBold']))
            elements.append(Paragraph(self._truncate_data(result.response_data), self.styles['Russian']))
            elements.append(Spacer(1, 10))
            
            # Ошибка (если есть)
            if result.error:
                elements.append(Paragraph("Ошибка:", self.styles['RussianBold']))
                elements.append(Paragraph(result.error, self.styles['Russian']))
                elements.append(Spacer(1, 10))
            
            elements.append(Spacer(1, 20))
        
        # Формирование PDF
        doc.build(elements)


class BankAPITester:
    """Класс для тестирования API банковского сервиса"""
    
    def __init__(self, base_url):
        """
        Инициализация тестера API
        
        Args:
            base_url: Базовый URL API
        """
        self.base_url = base_url.rstrip('/')
        if not self.base_url.endswith('/api'):
            self.base_url += '/api'
            
        self.token = None
        self.client_id = None
        self.account_id = None
        self.card_id = None
        self.loan_id = None
        self.results = []
        self.start_time = time.time()
        
        # Настройка логирования
        self.logger = logging.getLogger('api_tester')
        self.logger.setLevel(logging.DEBUG)
        
        # Добавляем обработчик для записи в файл
        file_handler = logging.FileHandler('api_tests.log')
        file_handler.setLevel(logging.DEBUG)
        
        # Добавляем обработчик для вывода в консоль
        console_handler = logging.StreamHandler()
        console_handler.setLevel(logging.INFO)
        
        # Создаем форматтер для логов
        formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
        file_handler.setFormatter(formatter)
        console_handler.setFormatter(formatter)
        
        # Добавляем обработчики к логеру
        self.logger.addHandler(file_handler)
        self.logger.addHandler(console_handler)
    
    def generate_client_data(self):
        """
        Генерирует случайные данные для регистрации клиента
        
        Returns:
            Dict: Данные клиента
        """
        random_suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=5))
        return {
            "fullName": f"Тестовый Пользователь {random_suffix}",
            "loginName": f"test_user_{random_suffix}",
            "contact": f"test{random_suffix}@example.com",
            "password": f"Password123_{random_suffix}"
        }
    
    def execute_request(self, name, method, endpoint, data=None, expected_status=200, 
                      auth_required=False, account_required=False, card_required=False, 
                      loan_required=False):
        """
        Выполняет запрос к API и обрабатывает результат
        
        Args:
            name: Название теста
            method: HTTP метод (GET, POST, PUT, DELETE)
            endpoint: Эндпоинт API
            data: Данные для отправки
            expected_status: Ожидаемый статус-код ответа
            auth_required: Требуется ли авторизация
            account_required: Требуется ли наличие счета
            card_required: Требуется ли наличие карты
            loan_required: Требуется ли наличие кредита
            
        Returns:
            Dict: Результат запроса
        """
        # Проверка зависимостей
        if auth_required and not self.token:
            self.logger.warning(f"Невозможно выполнить запрос {name}: не выполнена аутентификация")
            test_result = TestResult(
                name=name,
                url=f"{self.base_url}/{endpoint}",
                method=method,
                request_data={},
                response_data={"error": "Не выполнена аутентификация"},
                status_code=0,
                execution_time=0,
                passed=False,
                error="Не выполнена аутентификация"
            )
            self.results.append(test_result)
            return None
            
        if account_required and not self.account_id:
            self.logger.warning(f"Невозможно протестировать {name}: не создан счет")
            test_result = TestResult(
                name=name,
                url=f"{self.base_url}/{endpoint}",
                method=method,
                request_data={},
                response_data={"error": "Не создан счет"},
                status_code=0,
                execution_time=0,
                passed=False,
                error="Не создан счет"
            )
            self.results.append(test_result)
            return None
            
        if card_required and not self.card_id:
            self.logger.warning(f"Невозможно протестировать {name}: не создана карта")
            test_result = TestResult(
                name=name,
                url=f"{self.base_url}/{endpoint}",
                method=method,
                request_data={},
                response_data={"error": "Не создана карта"},
                status_code=0,
                execution_time=0,
                passed=False,
                error="Не создана карта"
            )
            self.results.append(test_result)
            return None
            
        if loan_required and not self.loan_id:
            self.logger.warning(f"Невозможно протестировать {name}: не оформлен кредит")
            test_result = TestResult(
                name=name,
                url=f"{self.base_url}/{endpoint}",
                method=method,
                request_data={},
                response_data={"error": "Не оформлен кредит"},
                status_code=0,
                execution_time=0,
                passed=False,
                error="Не оформлен кредит"
            )
            self.results.append(test_result)
            return None
        
        # Добавляем паузу в 0.5 секунды между запросами
        time.sleep(0.5)
        
        url = f"{self.base_url}/{endpoint}"
        headers = {}
        
        if auth_required and self.token:
            headers["Authorization"] = f"Bearer {self.token}"
            
        headers["Content-Type"] = "application/json"
        
        try:
            start_time = time.time()
            
            if method == "GET":
                response = requests.get(url, headers=headers)
            elif method == "POST":
                response = requests.post(url, json=data, headers=headers)
            elif method == "PUT":
                response = requests.put(url, json=data, headers=headers)
            elif method == "DELETE":
                response = requests.delete(url, json=data, headers=headers)
            else:
                raise ValueError(f"Неподдерживаемый метод: {method}")
                
            time_taken = time.time() - start_time
            
            # Попытка получить данные в формате JSON
            try:
                response_data = response.json()
            except:
                response_data = {"text": response.text}
                
            # Проверка статус-кода
            passed = response.status_code == expected_status
            
            self.logger.log(
                logging.INFO if passed else logging.ERROR, 
                f"{name}: статус {response.status_code}, время {time_taken:.2f}с"
            )
            
            if not passed:
                self.logger.error(f"Ожидался статус {expected_status}, получен {response.status_code}")
                self.logger.error(f"Ответ: {response_data}")
            
            test_result = TestResult(
                name=name,
                url=url,
                method=method,
                request_data=data or {},
                response_data=response_data,
                status_code=response.status_code,
                execution_time=time_taken,
                passed=passed
            )
            
            self.results.append(test_result)
            return response_data
            
        except Exception as e:
            time_taken = time.time() - start_time
            self.logger.error(f"Ошибка при выполнении запроса {name}: {str(e)}")
            traceback.print_exc()
            
            test_result = TestResult(
                name=name,
                url=url,
                method=method,
                request_data=data or {},
                response_data={"error": str(e)},
                status_code=0,
                execution_time=time_taken,
                passed=False,
                error=str(e)
            )
            
            self.results.append(test_result)
            return None

    def test_register(self):
        """Тестирует API регистрации нового клиента"""
        client_data = self.generate_client_data()
        # Сохраняем данные для последующей аутентификации
        self.client_data = client_data
        
        return self.execute_request(
            "Регистрация нового клиента",
            "POST",
            "auth/register",
            data=client_data,
            expected_status=201  # Изменено на 201 Created - стандартный код для создания ресурса
        )
    
    def test_login(self):
        """Тестирует API аутентификации клиента"""
        if not hasattr(self, 'client_data'):
            self.logger.error("Невозможно протестировать аутентификацию: нет данных клиента")
            return None
        
        login_data = {
            "loginName": self.client_data["loginName"],
            "password": self.client_data["password"]
        }
        
        response = self.execute_request(
            "Аутентификация клиента",
            "POST",
            "auth/login",
            data=login_data
        )
        
        if response and "token" in response:
            self.token = response["token"]
            self.logger.info(f"Получен токен: {self.token[:10]}...")
            
            # Если в ответе есть ID клиента, сохраняем его
            if "clientId" in response:
                self.client_id = response["clientId"]
                
        return response
    
    def test_auth_status(self):
        """Тестирует API получения статуса аутентификации"""
        return self.execute_request(
            "Проверка статуса аутентификации",
            "GET",
            "auth/status",
            auth_required=True
        )
    
    def test_get_profile(self):
        """Тестирует API получения профиля клиента"""
        return self.execute_request(
            "Получение профиля клиента",
            "GET",
            "auth/profile",
            auth_required=True
        )
    
    def test_create_account(self):
        """Тестирует API создания банковского счета"""
        account_data = {
            "title": "Тестовый счет",
            "initialFunds": 1000
        }
        
        response = self.execute_request(
            "Создание банковского счета",
            "POST",
            "accounts",
            data=account_data,
            auth_required=True
        )
        
        if response and "id" in response:
            self.account_id = response["id"]
        elif response and "accountId" in response:
            self.account_id = response["accountId"]
            
        return response
    
    def test_get_accounts(self):
        """Тестирует API получения списка счетов"""
        return self.execute_request(
            "Получение списка счетов",
            "GET",
            "accounts",
            auth_required=True
        )
    
    def test_get_account(self):
        """Тестирует API получения информации о конкретном счете"""
        if not self.account_id:
            self.logger.warning("Невозможно протестировать получение счета: не создан счет")
            return None
        
        return self.execute_request(
            "Получение информации о счете",
            "GET",
            f"accounts/{self.account_id}",
            auth_required=True,
            account_required=True
        )
    
    def test_deposit(self):
        """Тестирует API пополнения счета"""
        deposit_data = {
            "amount": 5000,
            "details": "Тестовое пополнение"
        }
        
        return self.execute_request(
            "Пополнение счета",
            "POST",
            f"accounts/{self.account_id}/deposit",
            data=deposit_data,
            auth_required=True,
            account_required=True
        )
    
    def test_withdraw(self):
        """Тестирует API снятия средств со счета"""
        withdraw_data = {
            "amount": 1000,
            "details": "Тестовое снятие"
        }
        
        return self.execute_request(
            "Снятие средств со счета",
            "POST",
            f"accounts/{self.account_id}/withdraw",
            data=withdraw_data,
            auth_required=True,
            account_required=True
        )
    
    def test_transfer(self):
        """Тестирует API перевода между счетами"""
        # Сначала создаем второй счет
        account_data = {
            "title": "Второй тестовый счет",
            "initialFunds": 1000
        }
        
        second_account_response = self.execute_request(
            "Создание второго счета",
            "POST",
            "accounts",
            data=account_data,
            auth_required=True
        )
        
        if not second_account_response or not ("id" in second_account_response or "accountId" in second_account_response):
            self.logger.warning("Невозможно протестировать перевод: не создан второй счет")
            return None
            
        second_account_id = second_account_response.get("id") or second_account_response.get("accountId")
        
        transfer_data = {
            "sourceAccountID": self.account_id,
            "targetAccountID": second_account_id,
            "amount": 500,
            "details": "Тестовый перевод"
        }
        
        return self.execute_request(
            "Перевод между счетами",
            "POST",
            "accounts/transfer",
            data=transfer_data,
            auth_required=True,
            account_required=True
        )
    
    def test_get_account_operations(self):
        """Тестирует API получения истории операций по счету"""
        return self.execute_request(
            "Получение истории операций",
            "GET",
            f"accounts/{self.account_id}/operations",
            auth_required=True,
            account_required=True
        )
    
    def test_create_card(self):
        """Тестирует API создания платежной карты"""
        card_data = {
            "accountId": self.account_id,
            "cardholderName": "TEST USER"
        }
        
        response = self.execute_request(
            "Создание платежной карты",
            "POST",
            "cards",
            data=card_data,
            auth_required=True,
            account_required=True
        )
        
        if response and "id" in response:
            self.card_id = response["id"]
        elif response and "cardId" in response:
            self.card_id = response["cardId"]
            
        return response
    
    def test_get_cards(self):
        """Тестирует API получения списка карт"""
        return self.execute_request(
            "Получение списка карт",
            "GET",
            "cards",
            auth_required=True
        )
    
    def test_get_card(self):
        """Тестирует API получения информации о конкретной карте"""
        return self.execute_request(
            "Получение информации о карте",
            "GET",
            f"cards/{self.card_id}",
            auth_required=True,
            card_required=True
        )
    
    def test_block_card(self):
        """Тестирует API блокировки карты"""
        return self.execute_request(
            "Блокировка карты",
            "POST",
            f"cards/{self.card_id}/block",
            auth_required=True,
            card_required=True
        )
    
    def test_unblock_card(self):
        """Тестирует API разблокировки карты"""
        return self.execute_request(
            "Разблокировка карты",
            "POST",
            f"cards/{self.card_id}/unblock",
            auth_required=True,
            card_required=True
        )
    
    def test_validate_card(self):
        """Тестирует API валидации карты"""
        # Для этого теста нужно получить данные карты
        card_response = self.execute_request(
            "Получение данных карты для валидации",
            "GET",
            f"cards/{self.card_id}",
            auth_required=True,
            card_required=True
        )
        
        if not card_response:
            return None
            
        # Формируем запрос на валидацию
        validate_data = {
            "cardNumber": card_response.get("cardNumber", ""),
            "expirationDate": card_response.get("expirationDate", ""),
            "cvv": card_response.get("cvv", "")
        }
        
        return self.execute_request(
            "Валидация карты",
            "POST",
            "cards/validate",
            data=validate_data,
            auth_required=True,
            card_required=True
        )
    
    def test_create_loan(self):
        """Тестирует API оформления кредита"""
        loan_data = {
            "accountId": self.account_id,
            "amount": 50000,
            "term": 12,
            "purpose": "Тестовый кредит"
        }
        
        response = self.execute_request(
            "Оформление кредита",
            "POST",
            "loans",
            data=loan_data,
            auth_required=True,
            account_required=True
        )
        
        if response and "id" in response:
            self.loan_id = response["id"]
        elif response and "loanId" in response:
            self.loan_id = response["loanId"]
            
        return response
    
    def test_get_loans(self):
        """Тестирует API получения списка кредитов"""
        return self.execute_request(
            "Получение списка кредитов",
            "GET",
            "loans",
            auth_required=True
        )
    
    def test_get_loan(self):
        """Тестирует API получения информации о конкретном кредите"""
        return self.execute_request(
            "Получение информации о кредите",
            "GET",
            f"loans/{self.loan_id}",
            auth_required=True,
            loan_required=True
        )
    
    def test_update_loan_payments(self):
        """Тестирует API обновления расчетов по всем кредитам"""
        return self.execute_request(
            "Обновление расчетов по кредитам",
            "PUT",
            "loans/update-payments",
            data={},
            auth_required=True
        )
    
    def test_get_loan_payment_plan(self):
        """Тестирует API получения графика платежей по кредиту"""
        return self.execute_request(
            "Получение графика платежей",
            "GET",
            f"loans/{self.loan_id}/payment-plan",
            auth_required=True,
            loan_required=True
        )
    
    def test_make_loan_payment(self):
        """Тестирует API внесения платежа по кредиту"""
        payment_data = {
            "amount": 5000
        }
        
        return self.execute_request(
            "Внесение платежа по кредиту",
            "POST",
            f"loans/{self.loan_id}/payment",
            data=payment_data,
            auth_required=True,
            loan_required=True
        )
    
    def test_calculate_loan(self):
        """Тестирует API расчета параметров кредита"""
        calc_data = {
            "amount": 100000,
            "term": 24,
            "purpose": "Расчет тестового кредита"
        }
        
        return self.execute_request(
            "Расчет параметров кредита",
            "POST",
            "loans/calculate",
            data=calc_data,
            auth_required=True
        )
    
    def test_get_analytics(self):
        """Тестирует API получения аналитики"""
        return self.execute_request(
            "Получение аналитики",
            "GET",
            "analytics/accounts",
            auth_required=True
        )
    
    def test_get_report(self):
        """Тестирует API получения отчета за период"""
        # Создаем даты для отчета за последний месяц
        end_date = datetime.datetime.now()
        start_date = end_date - datetime.timedelta(days=30)
        
        report_data = {
            "startDate": start_date.strftime("%Y-%m-%d"),
            "endDate": end_date.strftime("%Y-%m-%d")
        }
        
        return self.execute_request(
            "Получение отчета за период",
            "POST",
            "analytics/report",
            data=report_data,
            auth_required=True
        )
    
    def test_get_cb_rate(self):
        """Тестирует API получения ключевой ставки ЦБ"""
        return self.execute_request(
            "Получение ключевой ставки ЦБ РФ",
            "GET",
            "external/key-rate/current"
        )
    
    def test_get_cb_rate_history(self):
        """Тестирует API получения истории ключевой ставки ЦБ"""
        return self.execute_request(
            "Получение истории ключевой ставки",
            "GET",
            "external/key-rate/history"
        )
    
    def test_get_currency_rate(self):
        """Тестирует API получения курса валюты"""
        return self.execute_request(
            "Получение курса валюты",
            "GET",
            "external/currency-rate/USD"
        )
    
    def test_verify_client_data(self):
        """Тестирует API проверки данных клиента"""
        verify_data = {
            "fullName": "Тестовый Клиент",
            "documentNumber": "1234567890"
        }
        
        return self.execute_request(
            "Проверка данных клиента",
            "POST",
            "external/verify-client",
            data=verify_data,
            auth_required=True
        )
    
    def run_all_tests(self):
        """Запускает все тесты API в правильной последовательности"""
        start_time = time.time()
        self.start_time = start_time
        self.logger.info("Начало тестирования API")
        
        # Регистрация и аутентификация
        self.logger.info("=== Тестирование аутентификации ===")
        self.test_register()
        self.test_login()
        self.test_auth_status()
        self.test_get_profile()
        
        # Тесты для счетов
        self.logger.info("=== Тестирование счетов ===")
        self.test_create_account()
        self.test_get_accounts()
        self.test_get_account()
        self.test_deposit()
        self.test_withdraw()
        self.test_transfer()
        self.test_get_account_operations()
        
        # Тесты для карт
        self.logger.info("=== Тестирование карт ===")
        self.test_create_card()
        self.test_get_cards()
        self.test_get_card()
        self.test_block_card()
        self.test_unblock_card()
        self.test_validate_card()
        
        # Тесты для кредитов
        self.logger.info("=== Тестирование кредитов ===")
        self.test_create_loan()
        self.test_get_loans()
        self.test_get_loan()
        self.test_update_loan_payments()
        self.test_get_loan_payment_plan()
        self.test_make_loan_payment()
        self.test_calculate_loan()
        
        # Тесты для аналитики и прогнозов
        self.logger.info("=== Тестирование аналитики ===")
        self.test_get_analytics()
        self.test_get_report()
        
        # Тесты для внешних сервисов
        self.logger.info("=== Тестирование внешних сервисов ===")
        self.test_get_cb_rate()
        self.test_get_cb_rate_history()
        self.test_get_currency_rate()
        self.test_verify_client_data()
        
        total_time = time.time() - start_time
        self.logger.info(f"Завершено тестирование API за {total_time:.2f} секунд")
        
        # Вывод краткой сводки
        passed_tests = sum(1 for r in self.results if r.passed)
        failed_tests = len(self.results) - passed_tests
        
        print("\n==== Сводка тестирования ====")
        print(f"Всего тестов: {len(self.results)}")
        print(f"Успешно: {passed_tests}")
        print(f"Ошибки: {failed_tests}")
        print()
        
        if failed_tests > 0:
            print("Провалившиеся тесты:")
            for result in self.results:
                if not result.passed:
                    print(f"- {result.name}: код {result.status_code}")
            print()
        
        return total_time
    
    def generate_report(self, output_file: str = "api_test_report.pdf"):
        """Генерирует PDF-отчет о результатах тестирования"""
        self.logger.info(f"Создание отчета в файле {output_file}")
        
        report = APITestReport(self.results, total_time=time.time() - self.start_time)
        
        # Добавляем результаты всех тестов
        report.output(output_file)
        self.logger.info(f"Отчет сохранен в файле {output_file}")
        
        # Также создаем текстовый отчет для быстрого просмотра
        with open(output_file.replace(".pdf", ".txt"), "w", encoding="utf-8") as f:
            f.write("==== Отчет о тестировании API банковского сервиса ====\n\n")
            f.write(f"Дата: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
            
            for result in self.results:
                status = "УСПЕШНО" if result.passed else "ОШИБКА"
                f.write(f"{result.name}: {status}\n")
                f.write(f"  URL: {result.url}\n")
                f.write(f"  Метод: {result.method}\n")
                f.write(f"  Статус-код: {result.status_code}\n")
                f.write(f"  Время: {result.execution_time:.2f} сек\n\n")
            
            f.write("==== Сводка ====\n")
            f.write(f"Всего тестов: {len(self.results)}\n")
            f.write(f"Успешно: {sum(1 for r in self.results if r.passed)}\n")
            f.write(f"Ошибки: {sum(1 for r in self.results if not r.passed)}\n")
            f.write(f"Общее время выполнения: {time.time() - self.start_time:.2f} секунд\n")
        
        self.logger.info(f"Текстовый отчет сохранен в файле {output_file.replace('.pdf', '.txt')}")


def main():
    """Основная функция для запуска тестирования"""
    parser = argparse.ArgumentParser(description="Тестирование API банковского сервиса")
    parser.add_argument("--url", default="http://localhost:8080", help="Базовый URL API (по умолчанию http://localhost:8080)")
    parser.add_argument("--output", default="api_test_report.pdf", help="Имя файла для сохранения отчета (по умолчанию api_test_report.pdf)")
    args = parser.parse_args()
    
    try:
        # Проверяем, установлен ли модуль fpdf
        import fpdf
    except ImportError:
        print("Ошибка: библиотека fpdf не установлена. Устанавливаем...")
        try:
            import pip
            pip.main(['install', 'fpdf'])
        except Exception as e:
            print(f"Не удалось автоматически установить fpdf: {e}")
            print("Пожалуйста, установите вручную: pip install fpdf")
            sys.exit(1)
    
    # Создаем и запускаем тестирование
    tester = BankAPITester(args.url)
    tester.run_all_tests()
    tester.generate_report(args.output)
    
    # Выводим сводку в консоль
    total_tests = len(tester.results)
    passed_tests = sum(1 for r in tester.results if r.passed)
    failed_tests = total_tests - passed_tests
    
    print("\n==== Сводка тестирования ====")
    print(f"Всего тестов: {total_tests}")
    print(f"Успешно: {passed_tests}")
    print(f"Ошибки: {failed_tests}")
    
    if failed_tests > 0:
        print("\nПровалившиеся тесты:")
        for result in tester.results:
            if not result.passed:
                print(f"- {result.name}: код {result.status_code}")
    
    print(f"\nОтчет сохранен в файле {args.output}")
    print(f"Текстовый отчет сохранен в файле {args.output.replace('.pdf', '.txt')}")
    print(f"Лог сохранен в файле api_tests.log")


if __name__ == "__main__":
    main()
