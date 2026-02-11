package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"maxapigo/maxapi"
)

// Пример-заготовка: обработка сообщений с вложениями (файлы, картинки, видео и т.п.).
// На текущем этапе библиотека не декодирует все типы вложений из ответа MAX,
// поэтому пример показывает общий паттерн использования Router и Update.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Хендлер по типу обновления — message_created.
	// Здесь можно разбирать вложения сообщения.
	router.HandleUpdateType("message_created", func(ctx context.Context, upd maxapi.Update) error {
		if upd.Message == nil {
			return nil
		}

		msg := upd.Message
		log.Printf("[files example] got message with timestamp=%d text=%q", msg.Timestamp, msg.Text())

		if msg.Body == nil || len(msg.Body.Attachments) == 0 {
			log.Println("[files example] no attachments in message")
			return nil
		}

		for i, a := range msg.Body.Attachments {
			log.Printf("[files example] attachment #%d: type=%q payload=%T", i, a.Type, a.Payload)
			// Здесь можно делать разбор по типам вложений:
			//   - "file" / "image" / "video" / "sticker"
			//   - "inline_keyboard"
			//
			// Для вложений inline-клавиатуры payload будет соответствовать
			// структуре maxapi.InlineKeyboardPayload, если вы отправляли
			// её через maxapi.NewInlineKeyboardAttachment.
		}

		return nil
	})

	poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})

	log.Println("files example is running (long polling)...")
	if err := poller.Run(ctx); err != nil {
		log.Fatalf("poller error: %v", err)
	}
}

