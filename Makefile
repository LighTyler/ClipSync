# ClipSync Makefile
# Автоматически определяет ОС и собирает под неё

# Переменные
BINARY_NAME = clipsync
BUILD_DIR = build
CMD_PATH = ./cmd/app/
SERVICE_NAME = clipsync

# Определяем ОС
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows)
ifeq ($(UNAME_S),Linux)
    OS = linux
    EXT =
endif
ifeq ($(UNAME_S),Darwin)
    OS = darwin
    EXT =
endif
ifeq ($(OS),Windows)
    EXT = .exe
endif

# Флаги сборки
LDFLAGS = -s -w
BUILD_FLAGS = -ldflags "$(LDFLAGS)"

# Цели по умолчанию
.PHONY: all build run clean build-all install install-service uninstall-service help

all: build

## build: Собрать бинарник под текущую ОС
build:
	@echo "[build] Сборка для $(OS)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(EXT) $(CMD_PATH)
	@echo "[build] Готово: $(BUILD_DIR)/$(BINARY_NAME)$(EXT)"

## run: Собрать и запустить
run: build
	@echo "[run] Запуск..."
	@$(BUILD_DIR)/$(BINARY_NAME)$(EXT)

## build-all: Собрать под все поддерживаемые платформы
build-all: build-linux build-windows

## build-linux: Собрать для Linux (amd64)
build-linux:
	@echo "[build-linux] Сборка для Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)

## build-windows: Собрать для Windows (amd64)
build-windows:
	@echo "[build-windows] Сборка для Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)

## install: Установить бинарник в систему
install: build
ifeq ($(OS),linux)
	@echo "[install] Установка в /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "[install] Установлено"
else ifeq ($(OS),darwin)
	@echo "[install] Установка в /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "[install] Установлено"
else
	@echo "[install] Автоустановка для Windows не поддерживается. Скопируй $(BUILD_DIR)/$(BINARY_NAME).exe вручную."
endif

## install-service: Установить как systemd-сервис (только Linux)
install-service: install
ifeq ($(OS),linux)
	@echo "[install-service] Создание systemd-сервиса..."
	@mkdir -p ~/.config/systemd/user/
	@echo "[Unit]" > ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "Description=ClipSync P2P Clipboard Sync" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "After=network.target" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "[Service]" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "ExecStart=/usr/local/bin/$(BINARY_NAME)" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "Restart=on-failure" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "Environment=WAYLAND_DISPLAY=wayland-1" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "[Install]" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@echo "WantedBy=default.target" >> ~/.config/systemd/user/$(SERVICE_NAME).service
	@systemctl --user daemon-reload
	@systemctl --user enable --now $(SERVICE_NAME).service
	@echo "[install-service] Сервис установлен и запущен"
	@echo "[install-service] Логи: journalctl --user -u $(SERVICE_NAME) -f"
else
	@echo "[install-service] Установка сервиса поддерживается только в Linux"
endif

## uninstall-service: Удалить systemd-сервис (только Linux)
uninstall-service:
ifeq ($(OS),linux)
	@echo "[uninstall-service] Остановка и удаление сервиса..."
	@systemctl --user stop $(SERVICE_NAME).service 2>/dev/null || true
	@systemctl --user disable $(SERVICE_NAME).service 2>/dev/null || true
	@rm -f ~/.config/systemd/user/$(SERVICE_NAME).service
	@systemctl --user daemon-reload
	@echo "[uninstall-service] Сервис удалён"
else
	@echo "[uninstall-service] Удаление сервиса поддерживается только в Linux"
endif

## clean: Очистить артефакты сборки
clean:
	@echo "[clean] Очистка..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "[clean] Готово"

## test: Запустить тесты
test:
	@echo "[test] Запуск тестов..."
	@go test -v ./...

## lint: Запустить линтер (нужен golangci-lint)
lint:
	@echo "[lint] Запуск линтера..."
	@golangci-lint run ./...

## help: Показать эту справку
help:
	@echo "ClipSync - доступные команды:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Текущая ОС: $(OS)"