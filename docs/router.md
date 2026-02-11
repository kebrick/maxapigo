## Router и LongPoller — каркас бота

`Router` и `LongPoller` позволяют быстро написать бота, который:

- обрабатывает команды (`/start`, `/help` и т.п.);
- реагирует на смайлики/подстроки;
- умеет обрабатывать `message_callback` от inline‑клавиатуры.

Полезно в комбинации с long polling (`GET /updates`).

---

### Middleware (промежуточная обработка)

`Router` поддерживает middleware — функции, которые выполняются **до** вызова хендлеров.

```go
// UpdateMiddleware имеет сигнатуру:
// func(ctx context.Context, upd maxapi.Update) (handled bool, err error)
//
// handled == true означает, что апдейт уже обработан
// и вызывать другие middleware/хендлеры не нужно.
```

#### Пример: логирование всех апдейтов

```go
router.Use(func(ctx context.Context, upd maxapi.Update) (bool, error) {
    if upd.Message != nil {
        log.Printf(
            "[update] type=%s chat=%d user=%d text=%q",
            upd.UpdateType,
            upd.Message.Recipient.ChatID,
            upd.Message.Sender.UserID,
            upd.Message.Text(),
        )
    } else {
        log.Printf("[update] type=%s (non-message)", upd.UpdateType)
    }
    return false, nil // не останавливаем обработку
})
```

#### Пример: whitelist по user_id с ответом «нет доступа»

```go
allowed := map[int64]bool{
    123: true,
    456: true,
}

router.Use(func(ctx context.Context, upd maxapi.Update) (bool, error) {
    if upd.Message == nil || upd.Message.Sender == nil {
        return false, nil
    }

    userID := upd.Message.Sender.UserID
    if allowed[userID] {
        return false, nil // пользователь разрешён — продолжаем
    }

    // Пользователь не в списке — отвечаем и останавливаем обработку
    _, err := client.Messages().SendToChat(ctx, upd.Message.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "У вас нет доступа к этому боту.",
            Format: "markdown",
        })
    if err != nil {
        return true, err
    }

    return true, nil // handled=true — команды и остальные хендлеры не вызовутся
})
```

#### Пример: «тихий бан» — без ответа

```go
blocked := map[int64]bool{
    999: true,
}

router.Use(func(ctx context.Context, upd maxapi.Update) (bool, error) {
    if upd.Message == nil || upd.Message.Sender == nil {
        return false, nil
    }

    if blocked[upd.Message.Sender.UserID] {
        // Ничего не отвечаем и не даём апдейту пройти дальше
        return true, nil
    }

    return false, nil
})
```

---

### Базовый бот с командами и смайликом

```go
router := maxapi.NewRouter()

// /start — приветственное сообщение
router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Привет! Я демонстрационный бот.",
            Format: "markdown",
        })
    return err
})

// 😀 — простая реакция на смайлик
router.HandleEmoji("😀", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Вижу смайлик 😀 — настроение огонь!",
            Format: "markdown",
        })
    return err
})

// Обработчик по умолчанию — эхо без команд
router.HandleDefaultMessage(func(ctx context.Context, msg *maxapi.Message) error {
    if strings.HasPrefix(strings.TrimSpace(msg.Text()), "/") {
        return nil
    }
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text: "Вы написали: " + msg.Text(),
        })
    return err
})
```

---

### Подключаем LongPoller

```go
poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})

if err := poller.Run(ctx); err != nil {
    log.Fatal(err)
}
```

`LongPoller` берёт на себя цикл запросов `GET /updates` и передачу событий в `router.HandleUpdate`.

---

### Пример обработки callback‑события (inline‑клавиатура)

```go
router.HandleUpdateType("message_callback", func(ctx context.Context, upd maxapi.Update) error {
    if upd.Callback == nil {
        return nil
    }

    callbackID := upd.Callback.CallbackID
    notification := "Кнопка нажата ✅"

    _, err := client.Messages().AnswerCallback(ctx, callbackID, maxapi.CallbackAnswerBody{
        Message: &maxapi.NewMessageBody{
            Text: "Сообщение после нажатия кнопки",
        },
        Notification: &notification,
    })
    return err
})
```

---

### Где посмотреть живые примеры

- `examples/basic-bot` — минимальный бот, который отвечает на `/start`.
- `examples/router` — команды, смайлы и inline‑клавиатура.
- `examples/files` — пример обработки входящих сообщений с вложениями.

