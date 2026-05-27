# Workout Tracker

Сервис для отслеживания тренировок. Пользователи логируют тренировки, отслеживают упражнения, видят личные рекорды и прогресс.

## Стек

- **Бэкенд:** Go
- **База данных:** PostgreSQL
- **Авторизация:** JWT
- **Фронтенд:** HTML/JS

## Быстрый старт

```bash
# Запуск через Docker Compose
docker compose up -d

# Применить миграции
docker compose run --rm migrate

# Или локально
cp .env.example .env
make migrate-up
make run
```

## Структура проекта

```
├── cmd/              — точка входа
├── config/           — загрузка конфигурации
├── internal/
│   ├── handler/      — HTTP хэндлеры
│   ├── service/      — бизнес-логика
│   ├── repository/   — работа с БД
│   └── models/       — структуры данных
├── migrations/       — SQL миграции
├── bot/              — Telegram бот
└── web/              — HTML шаблоны и статика
```

## API

| Метод  | Путь                    | Описание                  | Роль  |
|--------|-------------------------|---------------------------|-------|
| POST   | /api/auth/register      | Регистрация               | -     |
| POST   | /api/auth/login         | Вход                      | -     |
| GET    | /api/exercises          | Список упражнений         | user  |
| POST   | /api/exercises          | Создать упражнение        | admin |
| PUT    | /api/exercises/:id      | Обновить упражнение       | admin |
| DELETE | /api/exercises/:id      | Удалить упражнение        | admin |
| GET    | /api/workouts           | Список тренировок         | user  |
| POST   | /api/workouts           | Создать тренировку        | user  |
| GET    | /api/workouts/:id       | Детали тренировки         | user  |
| PUT    | /api/workouts/:id       | Обновить тренировку       | user  |
| DELETE | /api/workouts/:id       | Удалить тренировку        | user  |
| GET    | /api/stats/pr           | Личные рекорды            | user  |
| GET    | /api/stats/volume       | Объём за неделю           | user  |
| GET    | /api/templates          | Список шаблонов           | user  |
| POST   | /api/templates          | Создать шаблон            | user  |

## Переменные окружения

| Переменная         | Описание                | По умолчанию |
|--------------------|-------------------------|--------------|
| PORT               | Порт сервера            | 8080         |
| DATABASE_URL       | URL подключения к БД    | -            |
| JWT_SECRET         | Секрет для JWT токенов  | -            |
| TELEGRAM_BOT_TOKEN | Токен Telegram бота     | -            |
