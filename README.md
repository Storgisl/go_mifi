# go_mifi
Обзор

Данный проект представляет собой REST API для банковского сервиса, разработанный на языке Go. Система обеспечивает безопасное управление банковскими счетами, картами, переводами, кредитными операциями и аналитикой. Проект реализует чистую архитектуру, разделение на слои (модели, репозитории, сервисы, обработчики, middleware), интеграцию с внешними сервисами (Центральный банк РФ для получения ключевой ставки, SMTP для уведомлений). Уделено особое внимание безопасности: шифрование данных карт (PGP + HMAC), хеширование паролей и CVV (bcrypt), аутентификация через JWT.
Функциональные возможности
Пользователи и аутентификация

    Регистрация с проверкой уникальности email и username.

    Аутентификация с выдачей JWT-токена (срок действия 24 часа).

    Защищённые маршруты (требуют валидный JWT).

Управление счетами и переводами

    Создание банковского счёта в валюте RUB.

    Пополнение счёта, списание средств.

    Переводы между счетами с проверкой достаточности средств и права доступа.

Операции с картами

    Генерация виртуальной карты с валидным номером по алгоритму Луна.

    Шифрование номера и срока карты (PGP-подобное симметричное шифрование AES-256-GCM).

    Хеширование CVV через bcrypt.

    HMAC для проверки целостности данных карты.

    Расшифровка и просмотр карт только для владельца счёта.

Кредитные операции

    Оформление кредита с расчётом аннуитетных платежей (ставка = ключевая ставка ЦБ РФ + маржа 5%).

    Генерация графика платежей (помесячно).

    Шедулер, запускаемый каждые 12 часов, автоматически списывает просроченные платежи и начисляет штраф +10% от суммы платежа при недостатке средств на счёте.

    Отправка email-уведомлений о платежах (через SMTP).

Аналитика и прогнозы

    Статистика доходов/расходов за выбранный месяц.

    Прогноз баланса на указанное количество дней (до 365 дней) с учётом будущих кредитных платежей.

Интеграции

    Центральный банк РФ: SOAP-запрос к сервису DailyInfo.asmx для получения актуальной ключевой ставки.

    SMTP: отправка уведомлений о платежах и штрафах на email пользователя.

Технологический стек
Компонент	Технологии
Язык	Go 1.23+
Маршрутизация	gorilla/mux
База данных	PostgreSQL 17 (расширение pgcrypto)
Драйвер БД	lib/pq
Аутентификация	JWT (golang-jwt/jwt/v5)
Логирование	logrus
Хеширование	bcrypt (пароли, CVV)
Шифрование карт	AES-256-GCM (имитация PGP) + HMAC-SHA256
Парсинг XML (ЦБ РФ)	beevik/etree
SMTP	go-mail/mail/v2
Дополнительно	алгоритм Луна для номеров карт
Структура проекта
text

bank-api/
├── main.go                 # точка входа, маршрутизация, запуск сервера и шедулера
├── config/
│   └── config.go           # загрузка конфигурации из env или значений по умолчанию
├── models/
│   └── models.go           # структуры данных, теги JSON, базовая валидация
├── repository/
│   └── repository.go       # инкапсуляция SQL-запросов, транзакции, параметризация
├── service/
│   ├── service.go          # бизнес-логика (регистрация, переводы, кредиты, аналитика)
│   ├── cbr_client.go       # SOAP-клиент для ЦБ РФ
│   ├── email_service.go    # отправка email через SMTP
│   └── scheduler.go        # шедулер для обработки просроченных платежей
├── handler/
│   └── handler.go          # HTTP-обработчики, валидация входных данных
├── middleware/
│   └── middleware.go       # JWT-проверка, добавление userID в контекст
├── utils/
│   ├── crypto.go           # bcrypt для паролей и CVV
│   ├── hmac.go             # HMAC-SHA256
│   ├── luhn.go             # генерация и проверка номеров карт по алгоритму Луна
│   └── pgp.go              # симметричное шифрование (AES-256-GCM) для номеров и сроков карт
├── init.sql                # схема БД (таблицы users, accounts, cards, transactions, credits, payment_schedules)
└── go.mod                  # зависимости

API Эндпоинты
Публичные

    POST /register – регистрация нового пользователя
    Тело: {"username":"alice","email":"alice@ex.com","password":"123"}

    POST /login – аутентификация, возвращает JWT
    Тело: {"email":"alice@ex.com","password":"123"}

Защищённые (требуется заголовок Authorization: Bearer <token>)
Метод	Путь	Описание
POST	/api/accounts	Создать новый счёт для текущего пользователя
POST	/api/accounts/{accountId}/deposit	Пополнить счёт (тело {"amount":1000})
POST	/api/transfer/{fromAccountId}	Перевод между счетами (тело {"to_account_id":2,"amount":500})
POST	/api/accounts/{accountId}/cards	Выпустить карту (тело {"cvv":"123"})
GET	/api/accounts/{accountId}/cards	Получить список карт (номер, срок расшифровываются для владельца)
POST	/api/accounts/{accountId}/credits	Оформить кредит (тело {"amount":50000,"months":12})
GET	/api/credits/{creditId}/schedule	График платежей по кредиту (реализация может быть расширена)
GET	/api/analytics?account_id=1&year=2026&month=6	Статистика доходов/расходов за месяц
GET	/api/accounts/{accountId}/predict?days=30	Прогноз баланса на N дней
Установка и запуск
Требования

    Go 1.23 или выше

    PostgreSQL 17 (или 16, порт 5432, либо измените конфиг)

    SMTP-сервер для уведомлений (опционально – для тестирования можно оставить заглушку)

Шаг 1. Клонирование и настройка БД
bash

git clone <your-repo-url> bank-api
cd bank-api
# Создайте базу данных и пользователя (пример для PostgreSQL)
sudo -u postgres psql <<EOF
CREATE USER bankuser WITH PASSWORD 'bankpass';
CREATE DATABASE bankdb OWNER bankuser;
GRANT ALL PRIVILEGES ON DATABASE bankdb TO bankuser;
EOF
# Импорт схемы
psql -h localhost -U bankuser -d bankdb -f init.sql

Шаг 2. Переменные окружения (опционально)

Можно задать через env или отредактировать config/config.go:
bash

export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=bankuser
export DB_PASSWORD=bankpass
export DB_NAME=bankdb
export JWT_SECRET=your_secret_key
export HMAC_SECRET=hmac_secret
export SMTP_HOST=smtp.example.com   # для реальных уведомлений
export SMTP_PORT=587
export SMTP_USER=noreply@example.com
export SMTP_PASS=your_password

Шаг 3. Запуск
bash

go mod tidy
go run main.go

Сервер запустится на http://localhost:8080.
Безопасность

    Пароли: хешируются через bcrypt (стойкость bcrypt.DefaultCost).

    Данные карт: номер и срок шифруются с помощью AES-256-GCM (симуляция PGP), CVV хешируется bcrypt. Для целостности используется HMAC-SHA256 номера карты.

    Аутентификация: JWT с секретом из JWT_SECRET. Проверка прав доступа к счёту/карте (пользователь может управлять только своими ресурсами).

    SQL-инъекции: все запросы параметризованы.

    Транзакции: переводы между счетами выполняются в БД-транзакциях.

Логирование

Логирование ведётся через logrus в формате JSON. Уровень по умолчанию – Info. Логируются регистрации, переводы, ошибки, запуск шедулера и т.д.
Тестирование

Примеры запросов с использованием curl:
bash

# Регистрация
curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"john","email":"john@doe.com","password":"123456"}'

# Логин (сохраните token)
TOKEN=$(curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"john@doe.com","password":"123456"}' | jq -r .token)

# Создать счёт
curl -X POST http://localhost:8080/api/accounts -H "Authorization: Bearer $TOKEN"

# Пополнить счёт 1
curl -X POST http://localhost:8080/api/accounts/1/deposit -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"amount":10000}'

# Выпустить карту
curl -X POST http://localhost:8080/api/accounts/1/cards -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"cvv":"123"}'
