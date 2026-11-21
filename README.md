# Ski Resort — система управления горнолыжным курортом

Веб-приложение для управления услугами горнолыжного курорта, прокатом снаряжения и бронированием занятий с инструкторами.

## Технологии

**Backend:** Go, REST API, PostgreSQL, pgx, JWT.

**Frontend:** Vue 3, TypeScript, Vite.

**Инфраструктура:** Docker, Docker Compose, Nginx, PostgreSQL 17, golang-migrate.

## Требования для запуска

На компьютере должны быть установлены:

- Git
- Docker с поддержкой Docker Compose

Для запуска через Docker не требуется отдельно устанавливать Go, Node.js или PostgreSQL.
Перед запуском убедитесь, что Docker Desktop (Windows/macOS) или Docker Engine с Compose (Linux) установлен и запущен.

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/darkline28/ski-resort.git
cd ski-resort
```

### 2. Настройка переменных окружения

Создайте файл `.env` в корневой папке проекта на основе `.env.example`.

Для macOS/Linux:

```bash
cp .env.example .env
```

Для Windows (PowerShell):

```powershell
Copy-Item .env.example .env
```

Откройте `.env` и укажите настройки PostgreSQL и JWT.

Пример:

```env
POSTGRES_USER=ski_user
POSTGRES_PASSWORD=change_me
POSTGRES_DB=ski_resort
POSTGRES_PORT=5434

JWT_SECRET=replace_with_random_secret_at_least_32_characters
JWT_TTL=30m
```

Замените пароль базы данных и JWT_SECRET на собственные значения. Секрет JWT должен содержать не менее 32 символов. Для пароля PostgreSQL в этой конфигурации рекомендуется использовать латинские буквы и цифры без специальных URL-символов.

Файл `.env` содержит конфиденциальные настройки и не должен публиковаться в Git.

### 3. Запуск приложения

Выполните в корневой папке проекта:

```bash
docker compose up --build -d
```

Docker автоматически соберёт backend и frontend, запустит PostgreSQL и применит миграции базы данных.

### Примечание при первом запуске
При первом запуске запуске Docker автоматически скачивает необходимые образы и собирает приложение. Это может занять несколько минут или больше в зависимости от скорости интернета и доступности Docker Hub. Последующие запуски обычно выполняются значительно быстрее благодаря кэшированию.

### 4. Открытие сайта

После успешного запуска откройте в браузере:

**http://localhost:8081**

Backend API доступен локально по адресу `http://localhost:8080`.

Frontend перенаправляет API-запросы в backend через Nginx.

## Управление приложением

Просмотр состояния контейнеров:

```bash
docker compose ps -a
```

Просмотр логов:

```bash
docker compose logs -f
```

Остановка приложения:

```bash
docker compose down
```

Повторный запуск:

```bash
docker compose up -d
```

Пересборка после изменений:

```bash
docker compose up --build -d
```

## База данных

Проект использует PostgreSQL 17.

При первом запуске автоматически создаётся база данных и применяются SQL-миграции из папки `backend/migrations`.

Данные PostgreSQL сохраняются в Docker volume `postgres_data` и остаются после обычной остановки контейнеров.

**Внимание:** команда `docker compose down -v` удаляет том с данными базы данных.

## Возможные проблемы

### Ошибка скачивания Docker-образов

Если при запуске появляется ошибка `DeadlineExceeded`,
`context deadline exceeded` или `failed to resolve reference`,
проверьте подключение к интернету и доступность Docker Hub.

После восстановления соединения повторите:

```bash
docker compose up --build -d
```

**Сайт не открывается**

Проверьте, что Docker запущен, а контейнеры работают:

```bash
docker compose ps -a
```

**Ошибка подключения к PostgreSQL**

Проверьте переменные окружения и убедитесь, что порт `5434` свободен.

**Порт уже занят**

Освободите соответствующий порт или измените его привязку в `compose.yaml`.

**Ошибки при запуске контейнеров**

Просмотрите сообщения:

```bash
docker compose logs --tail=100
```

**Изменение данных в `.env` не помогло**

После изменения настроек пересоздайте контейнеры:

```bash
docker compose up -d --force-recreate
```

Обратите внимание: параметры пользователя, пароля и имени базы, переданные PostgreSQL при первой инициализации, не изменяются автоматически в уже созданном хранилище.

## Структура проекта

```text
ski-resort/
├── backend/
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   └── montis-vue3/
│       ├── src/
│       ├── Dockerfile
│       └── nginx.conf
├── compose.yaml
├── .env.example
└── README.md
```

## Основные возможности

- Регистрация и авторизация пользователей
- Каталог горнолыжного снаряжения
- Просмотр услуг и инструкторов
- Бронирование занятий с инструкторами
- Бронирование и прокат оборудования
- Управление бронированиями
- Административные функции


Репозиторий: https://github.com/darkline28/ski-resort
