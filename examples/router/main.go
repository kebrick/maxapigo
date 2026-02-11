package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"maxapigo/maxapi"
)

// Пример: расширенный роутер — команды, смайлы, дефолтный хендлер.
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

	// /start — приветствие.
	router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
		text := "Привет! Я демонстрационный бот.\n\n" +
			"Попробуйте:\n" +
			"- отправить /help\n" +
			"- написать сообщение со смайликом 😀"

		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
			maxapi.NewMessageBody{
				Text:   text,
				Format: "markdown",
			})
		return err
	})

	// /help — справка.
	router.HandleCommand("/help", func(ctx context.Context, msg *maxapi.Message) error {
		text := "Доступные команды:\n" +
			"- /start — приветствие\n" +
			"- /help — эта справка\n" +
			"- /buttons — пример inline-клавиатуры"

		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID, maxapi.NewMessageBody{
			Text:   text,
			Format: "markdown",
		})
		return err
	})

	// /buttons — сообщение с inline-клавиатурой.
	router.HandleCommand("/buttons", func(ctx context.Context, msg *maxapi.Message) error {
		rows := [][]maxapi.InlineKeyboardButton{
			{
				{
					Type:    "callback",
					Text:    "Нажми меня",
					Payload: "button1_pressed",
				},
			},
		}

		body := maxapi.NewMessageBody{
			Text: "Сообщение с inline-клавиатурой:",
			Attachments: []maxapi.Attachment{
				maxapi.NewInlineKeyboardAttachment(rows),
			},
			Format: "markdown",
		}

		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID, body)
		return err
	})

	// Смайлик 😀.
	router.HandleEmoji("😀", func(ctx context.Context, msg *maxapi.Message) error {
		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
			maxapi.NewMessageBody{
				Text:   "Вижу смайлик 😀 — настроение огонь!",
				Format: "markdown",
			})
		return err
	})

	// Обработчик по умолчанию — логируем любые сообщения и отвечаем эхо.
	router.HandleDefaultMessage(func(ctx context.Context, msg *maxapi.Message) error {
		log.Printf("[router example] incoming text=%q from chat=%d", msg.Text(), msg.Recipient.ChatID)

		// Простое эхо без повторения команд.
		if strings.HasPrefix(strings.TrimSpace(msg.Text()), "/") {
			return nil
		}

		_, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
			maxapi.NewMessageBody{
				Text: "Вы написали: " + msg.Text(),
			})
		return err
	})

	poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})

	log.Println("router example is running (long polling)...")
	if err := poller.Run(ctx); err != nil {
		log.Fatalf("poller error: %v", err)
	}
}

