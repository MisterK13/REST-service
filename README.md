# Subscription Service API

REST-сервис для агрегации данных об онлайн подписках пользователей.

## 🚀 Технологии

- Go 1.23
- PostgreSQL 15
- Docker & Docker Compose
- Gin Web Framework
- GORM ORM
- Swagger (OpenAPI)
- Logrus (логирование)

## 📋 Функциональность

- ✅ CRUDL операции с подписками
- ✅ Подсчет суммарной стоимости подписок за период
- ✅ Фильтрация по пользователю и названию сервиса
- ✅ Swagger документация
- ✅ Логирование всех операций

## 🛠 Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.23+ (для локальной разработки)

### Cтарт через Docker

```bash
# 1. Клонировать репозиторий
git clone <your-repo-url>
cd REST_service

# 2. Создать .env файл
cp .env.example .env

# 3. Запустить сервис
docker-compose up -d

# 4. Проверить работу
curl http://localhost:8080/health