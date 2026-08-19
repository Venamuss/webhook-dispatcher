.PHONY: all vendor build up down restart logs clean migrate-up

# Переменные проекта
COMPOSE = docker compose

# Команда по умолчанию
all: run

# 1. Скачивание и вендоринг зависимостей на хосте
vendor:
	@echo "==> Tidying and vendoring Go modules..."
	go mod tidy
	go mod vendor

# 2. Сборка Docker-образов через вендоринг
build: vendor
	@echo "==> Building Docker images..."
	$(COMPOSE) build

# 3. Запуск всех сервисов в фоновом режиме
up:
	@echo "==> Starting all services..."
	$(COMPOSE) up -d

# 4. Полный цикл: вендоринг -> пересборка -> запуск
run: vendor
	@echo "==> Building and running stack..."
	$(COMPOSE) up --build -d

# 5. Остановка и удаление контейнеров
down:
	@echo "==> Stopping services..."
	$(COMPOSE) down

# 6. Перезапуск всего стека
restart: down up

# 7. Просмотр логов в реальном времени
logs:
	$(COMPOSE) logs -f

# 8. Запуск только базы и разовый накат миграций
migrate: vendor
	@echo "==> Running migrations container..."
	$(COMPOSE) up --build migrator

# 9. Полная очистка контейнеров, сетей и данных Postgres (volumes)
clean:
	@echo "==> Cleaning up containers, volumes and vendor..."
	$(COMPOSE) down -v --remove-orphans
	rm -rf vendor