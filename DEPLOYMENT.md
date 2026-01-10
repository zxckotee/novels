# Deployment Guide - Novels Platform

Руководство по развертыванию многоязычной платформы для чтения новелл.

## Содержание

1. [Требования](#требования)
2. [Локальная разработка](#локальная-разработка)
3. [Продакшн деплой](#продакшн-деплой)
4. [Конфигурация](#конфигурация)
5. [База данных](#база-данных)
6. [SSL сертификаты](#ssl-сертификаты)
7. [Мониторинг](#мониторинг)
8. [Резервное копирование](#резервное-копирование)

---

## Требования

### Минимальные системные требования

- **CPU**: 2 vCPU
- **RAM**: 4 GB
- **Disk**: 20 GB SSD
- **OS**: Ubuntu 22.04 LTS (рекомендуется)

### Необходимое ПО

- Docker 24.0+
- Docker Compose 2.20+
- Git

---

## Локальная разработка

### 1. Клонирование репозитория

```bash
git clone https://github.com/your-org/novels.git
cd novels
```

### 2. Настройка окружения

```bash
cp .env.example .env
# Отредактируйте .env файл
```

### 3. Запуск для разработки

```bash
# Запуск базы данных
docker-compose up -d db

# Backend (Go)
cd backend
go mod download
go run cmd/api/main.go

# Frontend (Next.js)
cd frontend
npm install
npm run dev
```

Сервисы будут доступны:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

---

## Продакшн деплой

### 1. Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. Клонирование и настройка

```bash
git clone https://github.com/your-org/novels.git
cd novels

# Настройка переменных окружения
cp .env.example .env
nano .env  # Заполните все значения
```

### 3. Генерация секретов

```bash
# JWT секреты
openssl rand -base64 32  # Для JWT_SECRET
openssl rand -base64 32  # Для JWT_REFRESH_SECRET
```

### 4. SSL сертификаты

```bash
# Создание директории для сертификатов
mkdir -p infra/nginx/ssl

# Вариант A: Let's Encrypt (рекомендуется)
# Используйте certbot для получения сертификатов

# Вариант B: Самоподписанные (только для тестирования)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout infra/nginx/ssl/privkey.pem \
  -out infra/nginx/ssl/fullchain.pem
```

### 5. Деплой

```bash
# Сделать скрипт исполняемым
chmod +x scripts/deploy.sh

# Запуск деплоя
./scripts/deploy.sh deploy

# Применение миграций
./scripts/deploy.sh migrate
```

### 6. Проверка

```bash
# Статус сервисов
./scripts/deploy.sh status

# Логи
./scripts/deploy.sh logs

# Проверка API
curl https://api.novels.example.com/health
```

---

## Конфигурация

### Основные переменные окружения

| Переменная | Описание | Пример |
|------------|----------|--------|
| `DB_PASSWORD` | Пароль БД | `secure_password` |
| `JWT_SECRET` | Секрет для JWT токенов | `base64_string` |
| `NEXT_PUBLIC_API_URL` | URL API для фронтенда | `https://api.novels.com` |
| `NEXT_PUBLIC_SITE_URL` | URL сайта | `https://novels.com` |
| `CORS_ORIGINS` | Разрешенные origins | `https://novels.com` |

### Настройка доменов

Обновите файл `infra/nginx/conf.d/default.conf`:
- Замените `novels.example.com` на ваш домен
- Замените `api.novels.example.com` на ваш API домен

---

## База данных

### Миграции

```bash
# Применение миграций
./scripts/deploy.sh migrate

# Или вручную
docker-compose -f docker-compose.prod.yml exec api ./server migrate up
```

### Создание первого администратора

```sql
-- Подключитесь к БД
docker-compose -f docker-compose.prod.yml exec db psql -U novels -d novels

-- Создайте пользователя
INSERT INTO users (email, password_hash, created_at)
VALUES ('admin@example.com', '$2a$10$...', NOW());

-- Добавьте роль администратора
INSERT INTO user_roles (user_id, role)
SELECT id, 'admin' FROM users WHERE email = 'admin@example.com';
```

---

## SSL сертификаты

### Let's Encrypt с Certbot

```bash
# Установка certbot
sudo apt install certbot

# Получение сертификатов
sudo certbot certonly --standalone -d novels.example.com -d api.novels.example.com

# Копирование сертификатов
sudo cp /etc/letsencrypt/live/novels.example.com/fullchain.pem infra/nginx/ssl/
sudo cp /etc/letsencrypt/live/novels.example.com/privkey.pem infra/nginx/ssl/

# Настройка автообновления
sudo crontab -e
# Добавьте: 0 0 1 * * certbot renew && docker-compose -f /path/to/novels/docker-compose.prod.yml restart nginx
```

---

## Мониторинг

### Проверка здоровья сервисов

```bash
# Все сервисы
./scripts/deploy.sh status

# Логи конкретного сервиса
./scripts/deploy.sh logs api
./scripts/deploy.sh logs web
./scripts/deploy.sh logs db
```

### Endpoints здоровья

- API: `GET /health`
- Frontend: `GET /`

---

## Резервное копирование

### Создание бэкапа

```bash
# Создать директорию для бэкапов
mkdir -p backups

# Создание бэкапа
./scripts/deploy.sh backup
```

### Восстановление из бэкапа

```bash
./scripts/deploy.sh restore backups/backup_20240115_120000.sql
```

### Автоматические бэкапы

```bash
# Добавьте в crontab
crontab -e

# Ежедневный бэкап в 3:00
0 3 * * * cd /path/to/novels && ./scripts/deploy.sh backup
```

---

## Обновление

### Стандартное обновление

```bash
# Получение обновлений
git pull origin main

# Пересборка и перезапуск
./scripts/deploy.sh deploy

# Применение новых миграций
./scripts/deploy.sh migrate
```

### Откат

```bash
# Остановка сервисов
./scripts/deploy.sh stop

# Откат к предыдущей версии
git checkout <previous-commit>

# Перезапуск
./scripts/deploy.sh deploy
```

---

## Устранение неполадок

### Сервис не запускается

```bash
# Проверьте логи
./scripts/deploy.sh logs <service>

# Проверьте конфигурацию
docker-compose -f docker-compose.prod.yml config
```

### Проблемы с БД

```bash
# Подключение к БД
docker-compose -f docker-compose.prod.yml exec db psql -U novels -d novels

# Проверка подключения
\conninfo
```

### Проблемы с SSL

```bash
# Проверка сертификатов
openssl x509 -in infra/nginx/ssl/fullchain.pem -text -noout

# Перезапуск nginx
./scripts/deploy.sh restart nginx
```

---

## Контакты

По вопросам деплоя обращайтесь к команде разработки.
