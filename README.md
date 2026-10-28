# ClipSync

P2P-синхронизатор буфера обмена между устройствами в локальной сети. Копируешь текст на одном компьютере — он мгновенно появляется на другом.

## Возможности

- Мгновенная синхронизация буфера обмена между устройствами
- P2P mesh-сеть — все устройства равноправны, нет центрального сервера
- Whitelist IP-адресов — защита от случайных подключений
- Защита от петель — хэширование сообщений предотвращает бесконечную пересылку
- Бинарный протокол — эффективная передача данных без накладных расходов JSON
- Поддержка буферов до 500 МБ (тексты, изображения)
- Кроссплатформенность — Windows и Linux (Wayland)

## Требования

### Windows
- Go 1.21+ (для сборки)
- Ничего дополнительного не нужно

### Linux (Wayland / Niri)
- Go 1.21+ (для сборки)
- `wl-clipboard` — для работы с буфером обмена

```bash
# Arch / CachyOS
sudo pacman -S wl-clipboard

# Ubuntu / Debian
sudo apt install wl-clipboard

# Fedora
sudo dnf install wl-clipboard
```

Важно: Linux-версия работает только под Wayland. Для X11 нужна отдельная реализация.

## Сборка

### Через Makefile (рекомендуется для Linux/macOS)

```bash
# Собрать под текущую ОС
make build

# Собрать и запустить
make run

# Собрать под все платформы сразу
make build-all

# Установить в систему
sudo make install

# Установить как systemd-сервис (Linux)
make install-service

# Очистить артефакты
make clean
```

### Ручная сборка

**Linux (CachyOS / Niri):**
```bash
go build -o clipsync ./cmd/app
```

**Windows (из PowerShell):**
```powershell
go build -o clipsync.exe .\cmd\app
```

**Кросс-компиляция (например, собрать для Windows с Linux):**
```bash
GOOS=windows GOARCH=amd64 go build -o clipsync.exe ./cmd/app
```

## Конфигурация

Конфиг по умолчанию находится в системной директории:
- **Linux:** `~/.config/ClipSync/config.json`
- **Windows:** `%APPDATA%\ClipSync\config.json`

Можно указать свой путь через флаг `-config`:
```bash
./clipsync -config /путь/к/config.json
```

### Пример config.json

```json
{
  "listen": ":9090",
  "peers": [
    "192.168.1.100:9090",
    "192.168.1.101:9090"
  ]
}
```

- `listen` — порт, на котором слушать входящие подключения
- `peers` — список адресов других устройств в сети

Совет: на каждом устройстве в `peers` указывай IP всех остальных устройств. Порядок не важен.

## Запуск

```bash
# С конфигом по умолчанию
./clipsync

# С кастомным конфигом
./clipsync -config ./my-config.json
```

Для остановки нажми `Ctrl+C` — программа корректно завершит все соединения.

## Запуск в фоне (как демон)

### Linux (systemd)

Автоматическая установка через Makefile:
```bash
sudo make install-service
```

Или вручную:
```bash
# Создать директорию для конфига
mkdir -p ~/.config/ClipSync

# Создать конфиг
nano ~/.config/ClipSync/config.json

# Создать systemd-сервис
mkdir -p ~/.config/systemd/user/
cat > ~/.config/systemd/user/clipsync.service <<EOF
[Unit]
Description=ClipSync P2P Clipboard Sync
After=network.target

[Service]
ExecStart=/usr/local/bin/clipsync
Restart=on-failure
Environment=WAYLAND_DISPLAY=wayland-1

[Install]
WantedBy=default.target
EOF

# Включить и запустить
systemctl --user daemon-reload
systemctl --user enable --now clipsync.service
```

Просмотр логов:
```bash
# Последние логи
journalctl --user -u clipsync -n 50

# Следить за логами в реальном времени
journalctl --user -u clipsync -f
```

### Windows

Добавь ярлык на `clipsync.exe` в папку автозагрузки:
1. Нажми `Win+R`, введи `shell:startup`
2. Скопируй ярлык на `clipsync.exe` в эту папку

Логи в Windows при запуске из автозагрузки по умолчанию не сохраняются. Если нужны логи — запусти программу вручную в терминале или используй PowerShell-скрипт с перенаправлением вывода.

## Архитектура

```
cmd/app/
└── main.go              # Точка входа, связывает все компоненты

config/
└──config.go             # Загрузка и валидация конфигурации

internal/
├── clipboard/
│   ├── clipboard.go         # Интерфейс Clipboard
│   ├── clipboard_windows.go # Реализация для Windows
│   └── clipboard_wayland.go # Реализация для Wayland
│ 
└── network/
    ├── protocol.go          # Бинарный протокол сообщений
    └── peer.go              # P2P-менеджер соединений
```

### Как это работает

1. Монитор локального буфера обмена (каждые 500 мс) проверяет изменения
2. При изменении — хэширует данные и рассылает всем подключённым пирам
3. Получатель проверяет хэш: если новый — записывает в свой буфер обмена
4. Защита от эха: отправитель помечает свои сообщения, получатель их игнорирует при обратной отправке

## Разработка

```bash
# Запуск с hot-reload (нужен air)
go install github.com/air-verse/air@latest
air

# Тесты
go test ./...

# Линтер
golangci-lint run
```

## Лицензия

MIT

## Благодарности

- [golang.design/x/clipboard](https://github.com/golang-design/clipboard) — кроссплатформенная работа с буфером обмена
- [wl-clipboard](https://github.com/bugaevc/wl-clipboard) — утилиты для Wayland