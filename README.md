# LestaStartTest

## Версия приложения

Текущая версия: **2.2.0**

---

## Описание проекта

LestaStartTest — это веб-приложение, разработанное на языке Go с использованием фреймворка Gin. Оно предоставляет API для работы с документами, коллекциями, пользователями, а также аутентификацию на основе JWT. В последней версии добавлен новый фронтенд, реализованный как SPA (Single Page Application).

---

## Структура проекта

Проект организован следующим образом:

```
LestaStartTest/
├── cmd/
│   └── main.go                // Точка входа в приложение
├── docs/
│   ├── changelog.md           // История изменений проекта
│   ├── database_schema.md     // Описание структуры базы данных
│   ├── docs.go                // Генератор документации
│   ├── swagger.json           // Swagger-документация (JSON)
│   └── swagger.yaml           // Swagger-документация (YAML)
├── internal/
│   ├── calculation/
│   │   ├── calculation.go     // Логика вычисления TF-IDF
│   │   └── calculation_test.go// Тесты для модуля вычислений
│   ├── controllers/
│   │   ├── auth.go            // API для аутентификации
│   │   ├── collections.go     // API для работы с коллекциями
│   │   ├── controllers.go     // Общая логика контроллеров
│   │   ├── documents.go       // API для работы с документами
│   │   ├── monitoring.go      // Метрики и статус приложения
│   │   └── user.go            // API для работы с пользователями
│   ├── db/
│   │   └── db.go              // Инициализация базы данных
│   ├── middleware/
│   │   └── jwt.go             // Middleware для JWT-аутентификации
│   ├── models/
│   │   └── models.go          // Определение моделей базы данных
│   ├── monitoring/
│   │   └── metrics.go         // Метрики приложения
├── static/
│   └── index.html             // Статика для SPA
├── uploads/                   // Директория для загруженных файлов
├── .env                       // Конфигурационные переменные окружения
├── .env.example               // Пример файла конфигурации
├── .gitignore                 // Игнорируемые файлы Git
├── docker-compose.yaml        // Конфигурация Docker Compose
├── Dockerfile                 // Dockerfile для сборки приложения
├── go.mod                     // Файл модулей Go
└── README.md                  // Документация проекта
```

---

## Основной функционал

- Регистрация и вход по JWT
- Загрузка/удаление текстовых документов
- Получение списков и содержимого документов
- Группировка документов в коллекции
- Подсчёт TF-IDF статистики по текстам
- Кодирование содержимого документа алгоритмом Хаффмана (с ограничением по размеру)
- Swagger-документация (см. `/swagger/index.html`)

---

## Стек технологий

- **Backend:** Golang (Gin)
- **JWT авторизация**
- **Swagger**
- **GORM**
- **PostgreSQL**
- **Frontend:** HTML (минимальный, для тестирования API)

---

## Запуск приложения

### Предварительные требования

- **Go**: версия 1.23
- **Docker**: убедитесь, что Docker и Docker Compose установлены.
- **PostgreSQL**: приложение зависит от PostgreSQL в качестве базы данных.

### Шаги запуска

1. Клонируйте репозиторий:
    ```bash
    git clone <адрес-репозитория>
    cd LestaStartTest
    ```

2. Создайте файл `.env` в корневой директории со следующим содержимым:
    ```
    VERSION=2.1.1 
    JWT_SECRET=uFuHAcBfwJXlAwy+jMRwVWA2K5DGQ7twFCFHSsAq98U=
    MAIN_PORT=:8080
    DB_HOST=db
    DB_USER=postgres
    DB_PASSWORD=2390
    DB_NAME=tfidf
    DB_PORT=5432
    ```

3. Запустите приложение с помощью Docker Compose:
    ```bash
    docker-compose up --build
    ```

4. Откройте `http://localhost:8080`, чтобы получить доступ к приложению.

---

## Конфигурируемые параметры

Все параметры приложения вынесены в файл `.env`. Вот список доступных конфигураций:

### Основные параметры

- `MAIN_PORT` — порт, на котором запускается приложение.
- `JWT_SECRET` — секрет для генерации JWT-токенов.

### Параметры БД

- `DB_HOST` — хост базы данных PostgreSQL.
- `DB_PORT` — порт подключения к базе данных (по умолчанию: `5432`).
- `DB_USER` — имя пользователя для подключения к базе данных.
- `DB_PASSWORD` — пароль для подключения к базе данных.
- `DB_NAME` — имя базы данных.

---

## Основные эндпоинты

### Аутентификация

- `POST /login` — Вход
- `POST /register` — Регистрация
- `GET /api/logout` — Выход

### Документы

- `GET /api/documents` — Список документов пользователя
- `POST /api/documents/upload` — Загрузка документа
- `GET /api/documents/{id}` — Получить документ по ID
- `GET /api/documents/{id}/statistics` — TF-IDF статистика для документа
- `GET /api/documents/{id}/huffman` — Получить Хаффман-код содержимого документа
- `DELETE /api/documents/{id}` — Удаление документа

### Коллекции

- `POST /api/collections` — Создать коллекцию
- `GET /api/collections` — Список коллекций пользователя
- `GET /api/collections/{id}` — Получить коллекцию
- `GET /api/collections/{id}/statistics` — TF-IDF статистика для коллекции
- `POST /api/collection/{collection_id}/{document_id}` — Добавить документ в коллекцию
- `DELETE /api/collection/{collection_id}/{document_id}` — Удалить документ из коллекции
- `DELETE /api/collections/{id}` — Удалить коллекцию

### Системные

- `GET /api/status` — Статус сервера
- `GET /api/metrics` — Метрики
- `GET /api/version` — Версия API

---

## Работа с Swagger

Swagger UI доступен по адресу [`/swagger/index.html`](http://localhost:8080/swagger/index.html).

### Особенности авторизации

- Для доступа к защищённым эндпоинтам требуется JWT-токен.
- После регистрации или входа вы получите токен (поле `"token"`).
- При работе с Swagger UI (кнопка **Authorize**):
   - Вставьте токен **с префиксом**:
     ```
     Bearer <ваш-токен>
     ```
     Например:
     ```
     Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
     ```
   - Без приставки `Bearer` авторизация работать не будет.

- Если отправлять запросы через curl или Postman, также используйте:
    ```
    -H "Authorization: Bearer <ваш-токен>"
    ```

### Документация по API

- Актуальный список эндпоинтов, схемы запросов и ответов всегда доступен в Swagger UI.
- Если после обновления кода Swagger UI начинает показывать неверные маршруты — пересоздайте документацию с помощью:
    ```
    swag init --generalInfo cmd/main.go --output docs --parseDependency --parseInternal --parseVendor=false
    ```

---

## Документация

- [Структура базы данных](docs/database_schema.md)
- [История изменений (changelog)](docs/changelog.md)
