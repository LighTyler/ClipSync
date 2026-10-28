package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LighTyler/ClipSync/config"
	"github.com/LighTyler/ClipSync/internal/clipboard"
	"github.com/LighTyler/ClipSync/internal/network"
)

func main() {
	// 1. Парсинг аргументов
	configPath := flag.String("config", config.GetDefaultConfigPath(), "Путь к файлу конфигурации")
	flag.Parse()

	// 2. Загрузка конфигурации
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Starting WeAreClipboard | Listen: %s | Peers: %v", cfg.Listen, cfg.Peers)

	// 3. Инициализация менеджера буфера обмена (выберет реализацию под ОС)
	cb := clipboard.NewManager()

	// 4. Настройка контекста и graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutdown signal received, stopping...")
		cancel()
	}()

	// 5. Инициализация P2P-сети
	peer := network.NewPeer(cfg.Listen, cfg.Peers)
	go peer.Run(ctx)

	// Состояние для предотвращения эха и циклов
	var (
		lastHash         string
		isInternalUpdate bool
		mu               sync.Mutex
	)

	// 6. Горутина: Сеть -> Локальный буфер обмена
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-peer.MessageChan():
				hash := sha256.Sum256(msg.Payload)
				hashStr := hex.EncodeToString(hash[:])

				mu.Lock()
				if hashStr == lastHash {
					mu.Unlock()
					continue // Это эхо нашего же сообщения, игнорируем
				}
				// Помечаем, что следующее изменение локального буфера — это наше собственное действие
				isInternalUpdate = true
				lastHash = hashStr
				mu.Unlock()

				log.Printf("📥 Received from network, writing to clipboard (hash: %s)", hashStr[:8])
				if err := cb.Write(msg.Payload); err != nil {
					log.Printf("⚠️ Failed to write to clipboard: %v", err)
				}
			}
		}
	}()

	// 7. Горутина: Локальный буфер обмена -> Сеть
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				data, err := cb.Read()
				if err != nil || len(data) == 0 {
					continue // Буфер пуст или ошибка чтения, ждём следующего тика
				}

				hash := sha256.Sum256(data)
				hashStr := hex.EncodeToString(hash[:])

				mu.Lock()
				internal := isInternalUpdate
				if internal {
					// Сбрасываем флаг, так как мы уже обработали это изменение
					isInternalUpdate = false
				}
				mu.Unlock()

				if internal {
					// Это изменение вызвано нашей же записью из сети, не отправляем его обратно
					lastHash = hashStr
					continue
				}

				if hashStr == lastHash {
					continue // Буфер не изменился
				}

				// Буфер изменился пользователем, отправляем в сеть
				lastHash = hashStr
				log.Printf("📤 Local clipboard changed, sending to network (hash: %s)", hashStr[:8])

				// Ошибки внутри SendMessage уже залогированы, нам достаточно знать, что попытка была
				_ = peer.SendMessage(network.Message{
					MIME:    "text/plain",
					Payload: data,
				})
			}
		}
	}()

	// 8. Блокировка основного потока до отмены контекста
	<-ctx.Done()
	log.Println("WeAreClipboard stopped gracefully.")
}
