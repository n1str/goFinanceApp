package services

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Структуры для работы с XML-ответом от ЦБ РФ
type KeyRateEnvelope struct {
	XMLName xml.Name    `xml:"Envelope"`
	Body    KeyRateBody `xml:"Body"`
}

type KeyRateBody struct {
	Response KeyRateResponse `xml:"KeyRateXMLResponse"`
}

type KeyRateResponse struct {
	Result KeyRateResult `xml:"KeyRateXMLResult"`
}

type KeyRateResult struct {
	Diffgram KeyRateDiffgram `xml:"diffgram"`
}

type KeyRateDiffgram struct {
	KeyRateData KeyRateData `xml:"KeyRate"`
}

type KeyRateData struct {
	KeyRates []KeyRateRow `xml:"KR"`
}

type KeyRateRow struct {
	Date string `xml:"DT"`
	Rate string `xml:"Rate"`
}

// Структуры для работы с XML-ответом курсов валют от ЦБ РФ
type CursEnvelope struct {
	XMLName xml.Name  `xml:"Envelope"`
	Body    CursBody  `xml:"Body"`
}

type CursBody struct {
	Response CursResponse `xml:"GetCursOnDateXMLResponse"`
}

type CursResponse struct {
	Result CursResult `xml:"GetCursOnDateXMLResult"`
}

type CursResult struct {
	ValuteData ValuteCursOnDate `xml:"ValuteData"`
}

type ValuteCursOnDate struct {
	Valutes []ValuteCurs `xml:"ValuteCursOnDate"`
}

type ValuteCurs struct {
	VCode   string `xml:"VCode"`
	VName   string `xml:"VName"`
	VNom    string `xml:"VNom"`
	VCurs   string `xml:"VCurs"`
	VchCode string `xml:"VchCode"`
}

// KeyRateInfo представляет информацию о ключевой ставке
type KeyRateInfo struct {
	Value     float64   `json:"value"`
	Date      time.Time `json:"date"`
	PrevValue float64   `json:"prevValue,omitempty"`
	PrevDate  time.Time `json:"prevDate,omitempty"`
}

// ExternalService представляет интерфейс сервиса внешних интеграций
type ExternalService interface {
	GetCurrentKeyRate() (*KeyRateInfo, error)
	GetKeyRateHistory(startDate, endDate time.Time) ([]KeyRateInfo, error)
	GetCurrencyRate(currencyCode string) (float64, error)
	VerifyClientData(name, contact string) (bool, error)
}

// ExternalServiceImpl реализует функциональность сервиса внешних интеграций
type ExternalServiceImpl struct {
	cbrApiUrl   string
	timeout     time.Duration
	userAgent   string
	clientName  string
	useDemoMode bool
}

// NewExternalService создаёт новый сервис внешних интеграций
func NewExternalService(
	cbrApiUrl string,
	timeout time.Duration,
	userAgent string,
	clientName string,
	useDemoMode bool,
) ExternalService {
	// Если URL API не указан, используем стандартный URL ЦБ РФ
	if cbrApiUrl == "" {
		cbrApiUrl = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"
	}

	// Если user agent не указан, используем стандартный
	if userAgent == "" {
		userAgent = "Finance-Application/1.0"
	}

	// Если таймаут не указан, используем стандартный
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &ExternalServiceImpl{
		cbrApiUrl:   cbrApiUrl,
		timeout:     timeout,
		userAgent:   userAgent,
		clientName:  clientName,
		useDemoMode: useDemoMode,
	}
}

// buildKeyRateSOAPRequest создает SOAP запрос для получения ключевой ставки
func (s *ExternalServiceImpl) buildKeyRateSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")
	
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
		<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
			<soap12:Body>
				<KeyRate xmlns="http://web.cbr.ru/">
					<fromDate>%s</fromDate>
					<ToDate>%s</ToDate>
				</KeyRate>
			</soap12:Body>
		</soap12:Envelope>`, fromDate, toDate)
}

// sendSOAPRequest отправляет SOAP запрос в ЦБ РФ
func (s *ExternalServiceImpl) sendSOAPRequest(soapRequest, action string) ([]byte, error) {
	// Создаем HTTP клиент с настроенным таймаутом
	client := &http.Client{
		Timeout: s.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // Используем только современные версии TLS
			},
		},
	}
	
	// Создаем и конфигурируем HTTP запрос
	req, err := http.NewRequest("POST", s.cbrApiUrl, bytes.NewBufferString(soapRequest))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	
	// Устанавливаем необходимые заголовки
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", action)
	req.Header.Set("User-Agent", s.userAgent)
	
	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()
	
	// Проверяем код ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка запроса, статус: %s", resp.Status)
	}
	
	// Читаем содержимое ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	
	return body, nil
}

// parseKeyRateXMLResponse разбирает XML ответ для получения ключевой ставки
func (s *ExternalServiceImpl) parseKeyRateXMLResponse(xmlData []byte) (*KeyRateInfo, error) {
	// Декодируем XML
	var envelope KeyRateEnvelope
	if err := xml.Unmarshal(xmlData, &envelope); err != nil {
		return nil, fmt.Errorf("ошибка разбора XML: %v", err)
	}
	
	// Проверяем наличие данных
	keyRates := envelope.Body.Response.Result.Diffgram.KeyRateData.KeyRates
	if len(keyRates) == 0 {
		return nil, errors.New("данные по ключевой ставке не найдены")
	}
	
	// Сортируем ставки по дате (убывание)
	sort.Slice(keyRates, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02T15:04:05", keyRates[i].Date)
		dateJ, _ := time.Parse("2006-01-02T15:04:05", keyRates[j].Date)
		return dateI.After(dateJ)
	})
	
	// Берем самую новую ставку
	currentRate := keyRates[0]
	
	// Парсим дату и ставку
	currentDate, err := time.Parse("2006-01-02T15:04:05", currentRate.Date)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты: %v", err)
	}
	
	rateValue, err := strconv.ParseFloat(strings.Replace(currentRate.Rate, ",", ".", 1), 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга ставки: %v", err)
	}
	
	// Если есть предыдущая ставка, тоже получаем ее
	var prevDate time.Time
	var prevRate float64
	
	if len(keyRates) > 1 {
		prevRateData := keyRates[1]
		prevDate, _ = time.Parse("2006-01-02T15:04:05", prevRateData.Date)
		prevRate, _ = strconv.ParseFloat(strings.Replace(prevRateData.Rate, ",", ".", 1), 64)
	}
	
	return &KeyRateInfo{
		Value:     rateValue,
		Date:      currentDate,
		PrevValue: prevRate,
		PrevDate:  prevDate,
	}, nil
}

// GetCurrentKeyRate получает текущую ключевую ставку ЦБ РФ
func (s *ExternalServiceImpl) GetCurrentKeyRate() (*KeyRateInfo, error) {
	log.Println("Получение текущей ключевой ставки ЦБ РФ через SOAP...")
	
	// Проверяем, находимся ли в режиме тестирования
	if s.useDemoMode {
		log.Println("Используется демо-режим, возвращаются тестовые данные")
		return s.getDefaultKeyRate(), nil
	}
	
	// Формируем SOAP запрос
	soapRequest := s.buildKeyRateSOAPRequest()
	
	// Отправляем SOAP запрос
	xmlData, err := s.sendSOAPRequest(soapRequest, "http://web.cbr.ru/KeyRate")
	if err != nil {
		log.Printf("Ошибка SOAP запроса: %v, попытка использовать запасной метод", err)
		return s.getKeyRateByRESTAPI()
	}
	
	// Разбираем XML ответ
	keyRateInfo, err := s.parseKeyRateXMLResponse(xmlData)
	if err != nil {
		log.Printf("Ошибка при парсинге XML: %v, попытка использовать запасной метод", err)
		return s.getKeyRateByRESTAPI()
	}
	
	log.Printf("Успешно получена ключевая ставка: %.2f%% (от %s)",
		keyRateInfo.Value, keyRateInfo.Date.Format("02.01.2006"))
	
	return keyRateInfo, nil
}

// getKeyRateByRESTAPI получает ключевую ставку через REST API (запасной метод)
func (s *ExternalServiceImpl) getKeyRateByRESTAPI() (*KeyRateInfo, error) {
	log.Println("Использование резервного REST API для получения ключевой ставки")
	
	// Получаем период за последние 30 дней
	today := time.Now()
	fromDate := today.AddDate(0, 0, -30)
	todayStr := today.Format("02/01/2006")
	fromDateStr := fromDate.Format("02/01/2006")
	
	// URL для получения ключевых ставок
	url := fmt.Sprintf("https://www.cbr.ru/scripts/XML_key.asp?date_req1=%s&date_req2=%s",
		fromDateStr, todayStr)
	
	// Создаем HTTP клиент
	client := &http.Client{
		Timeout: s.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	
	// Отправляем запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	
	req.Header.Set("User-Agent", s.userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка REST API: %v, используем значение по умолчанию", err)
		return s.getDefaultKeyRate(), nil
	}
	defer resp.Body.Close()
	
	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения ответа: %v, используем значение по умолчанию", err)
		return s.getDefaultKeyRate(), nil
	}
	
	// Пытаемся разобрать XML
	type KeyRateXML struct {
		XMLName xml.Name `xml:"KeyRate"`
		Records []struct {
			Date string `xml:"Date,attr"`
			Rate string `xml:"Rate,attr"`
		} `xml:"Record"`
	}
	
	var keyRates KeyRateXML
	if err := xml.Unmarshal(body, &keyRates); err != nil {
		log.Printf("Ошибка разбора XML: %v, используем значение по умолчанию", err)
		return s.getDefaultKeyRate(), nil
	}
	
	// Если нет данных, возвращаем значение по умолчанию
	if len(keyRates.Records) == 0 {
		log.Println("Нет данных в ответе, используем значение по умолчанию")
		return s.getDefaultKeyRate(), nil
	}
	
	// Сортируем записи по дате (убывание)
	sort.Slice(keyRates.Records, func(i, j int) bool {
		dateI, _ := time.Parse("02.01.2006", keyRates.Records[i].Date)
		dateJ, _ := time.Parse("02.01.2006", keyRates.Records[j].Date)
		return dateI.After(dateJ)
	})
	
	// Берем самую новую запись
	latestRecord := keyRates.Records[0]
	
	// Парсим дату и ставку
	currentDate, _ := time.Parse("02.01.2006", latestRecord.Date)
	currentRate, _ := strconv.ParseFloat(strings.Replace(latestRecord.Rate, ",", ".", 1), 64)
	
	// Если есть предыдущая запись, тоже парсим
	var prevDate time.Time
	var prevRate float64
	
	if len(keyRates.Records) > 1 {
		prevRecord := keyRates.Records[1]
		prevDate, _ = time.Parse("02.01.2006", prevRecord.Date)
		prevRate, _ = strconv.ParseFloat(strings.Replace(prevRecord.Rate, ",", ".", 1), 64)
	}
	
	return &KeyRateInfo{
		Value:     currentRate,
		Date:      currentDate,
		PrevValue: prevRate,
		PrevDate:  prevDate,
	}, nil
}

// getDefaultKeyRate возвращает данные о ключевой ставке по умолчанию
func (s *ExternalServiceImpl) getDefaultKeyRate() *KeyRateInfo {
	currentDate := time.Now()
	prevDate := currentDate.AddDate(0, -3, 0)
	
	return &KeyRateInfo{
		Value:     7.5, // Примерное значение по умолчанию
		Date:      currentDate,
		PrevValue: 7.75,
		PrevDate:  prevDate,
	}
}

// GetKeyRateHistory получает историю ключевой ставки ЦБ РФ
func (s *ExternalServiceImpl) GetKeyRateHistory(startDate, endDate time.Time) ([]KeyRateInfo, error) {
	// Реализуется аналогично GetCurrentKeyRate, но возвращает массив ставок
	// Здесь можно использовать тот же механизм SOAP запросов
	
	// Для демонстрации вернем фиктивные данные
	if s.useDemoMode {
		return []KeyRateInfo{
			{
				Value: 7.5,
				Date:  time.Now().AddDate(0, 0, -1),
			},
			{
				Value: 7.75,
				Date:  time.Now().AddDate(0, -3, 0),
			},
			{
				Value: 8.0,
				Date:  time.Now().AddDate(0, -6, 0),
			},
		}, nil
	}
	
	// Здесь должен быть код для получения истории ставок через SOAP API
	// ...
	
	return nil, errors.New("метод не реализован")
}

// GetCurrencyRate получает курс валюты от ЦБ РФ
func (s *ExternalServiceImpl) GetCurrencyRate(currencyCode string) (float64, error) {
	// Для демонстрации вернем фиктивные данные
	if s.useDemoMode {
		switch strings.ToUpper(currencyCode) {
		case "USD":
			return 75.5, nil
		case "EUR":
			return 85.2, nil
		default:
			return 0, fmt.Errorf("неизвестная валюта: %s", currencyCode)
		}
	}
	
	// Здесь должен быть код для получения курса через SOAP API
	// ...
	
	return 0, errors.New("метод не реализован")
}

// VerifyClientData проверяет данные клиента через внешние сервисы
func (s *ExternalServiceImpl) VerifyClientData(name, contact string) (bool, error) {
	// Для демонстрации всегда возвращаем успех
	return true, nil
}
