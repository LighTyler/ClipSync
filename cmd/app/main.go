package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
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
	log.Printf("Starting ClipSync | Listen: %s | Peers: %v", cfg.Listen, cfg.Peers)

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

	// Горутина 6: Сеть → Буфер обмена
	go func() {
		for msg := range peer.MessageChan() {
			log.Printf("Received from network, writing to clipboard")
			if err := cb.Write(msg.Payload); err != nil {
				log.Printf("Failed to write to clipboard: %v", err)
			}
		}
	}()

	// Горутина 2: Буфер обмена → Сеть
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
					continue
				}

				err = peer.SendMessage(network.Message{
					MIME:    "text/plain",
					Payload: data,
				})
				if err != nil {
					log.Printf("Failed to send message: %v", err)
				}
			}
		}
	}()

	// 8. Блокировка основного потока до отмены контекста
	<-ctx.Done()
	log.Println("ClipSync stopped gracefully.")
}
