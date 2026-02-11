package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"maxapigo/maxapi"
)

// Пример: минимальный бот, отвечающий на /start.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Корректная остановка по Ctrl+C.
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		log.Println("shutting down...")
		cancel()
	}()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	client, err := maxapi.NewClient(maxapi.Config{
		Token:  token,
		Logger: maxapi.NewStdLogger(nil),
		Debug:  true,
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	router := maxapi.NewRouter()

	// /start — приветственное сообщение.
	router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
			maxapi.NewMessageBody{
				Text:   "Привет! Я минимальный бот на Go для MAX.",
				Format: "markdown",
			})
		return err
	})

	poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})

	log.Println("basic-bot is running (long polling)...")
	if err := poller.Run(ctx); err != nil {
		log.Fatalf("poller error: %v", err)
	}
}

