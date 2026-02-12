## Документация maxapigo

Библиотека `maxapigo` упрощает работу с API мессенджера MAX из Go‑кода.  
Ниже описаны основные сущности клиента, используемые методы API и примеры кода.

Актуальная документация по API MAX: [`https://dev.max.ru/docs-api`](https://dev.max.ru/docs-api).

---

### Подключение и инициализация клиента

```bash
go get github.com/kebrick/maxapigo
```

```go
import (
    "context"
    "log"

    "github.com/kebrick/maxapigo/maxapi"
)

func main() {
    ctx := context.Background()

    client, err := maxapi.NewClient(maxapi.Config{
        Token:  "<BOT_TOKEN>",            // токен бота из MAX → «Интеграция» → «Получить токен»
        Logger: maxapi.NewStdLogger(nil), // опционально: логгер для отладки
        Debug:  true,                     // подробные логи запросов/ответов
    })
    if err != nil {
        log.Fatalf("create client: %v", err)
    }

    // дальнейшая работа с client.Bots(), client.Messages() и т.д.
}
```

Клиент инкапсулирует:

- базовый URL `https://platform-api.max.ru`
- авторизацию через заголовок `Authorization: <token>`
- высокоуровневые сервисы:
  - `Bots()` — данные о боте (`GET /me`)
  - `Messages()` — отправка и управление сообщениями
  - `Chats()` — групповые чаты
  - `Subscriptions()` — подписки и получение обновлений
  - `Uploads()` — загрузка файлов и информация о видео

---

## Bots — информация о боте (`GET /me`)

### Получение информации о боте

**Метод API:** `GET /me`  
**Интерфейс:** `client.Bots().Me(ctx)`

```go
ctx := context.Background()

me, err := client.Bots().Me(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("Bot: %s (id=%d, username=%s)", me.Name, me.UserID, me.Username)
```

Полезно для:

- проверки, что токен корректен;
- логирования базовой информации о боте при старте приложения.

---

## Messages — работа с сообщениями

Сервис `Messages()` инкапсулирует методы раздела `messages` из [`https://dev.max.ru/docs-api`](https://dev.max.ru/docs-api).

### Отправка сообщения (`POST /messages`)

**Метод API:** `POST /messages`  
**Интерфейс:** `client.Messages().Send` / `SendToChat`

Простой пример: отправка сообщения по `chat_id`:

```go
msg, err := client.Messages().SendToChat(ctx, 123456,
    maxapi.NewMessageBody{
        Text:   "Привет из Go!",
        Format: "markdown", // см. раздел «Форматирование текста» в доке MAX
    })
if err != nil {
    log.Fatal(err)
}

log.Printf("sent at %d, text=%q", msg.Timestamp, msg.Text())
```

Параметры `NewMessageBody`:

- `Text` — текст сообщения;
- `Attachments` — список вложений (файлы, картинки, видео, inline‑клавиатуры и т.п.);
- `Notify` — уведомлять ли участников чата (`true` по умолчанию);
- `Format` — `markdown` или `html`.

---

### Получение сообщений (`GET /messages`)

**Метод API:** `GET /messages`  
**Интерфейс:** `client.Messages().ListMessages(ctx, params)`

```go
chatID := int64(123456)

messages, err := client.Messages().ListMessages(ctx, maxapi.ListMessagesParams{
    ChatID: &chatID,
    Count:  50, // до 100
})
if err != nil {
    log.Fatalf("list messages: %v", err)
}

for _, m := range messages {
    log.Printf("message at %d: %q", m.Timestamp, m.Text())
}
```

Поддерживаются параметры:

- `ChatID *int64` — получить сообщения конкретного чата;
- `MessageIDs []string` — получить сообщения по списку `message_id`;
- `From`, `To` — временной диапазон (`Unix timestamp`);
- `Count` — количество сообщений (1–100, по умолчанию 50).

Требуется указать либо `ChatID`, либо `MessageIDs`.

---

### Получение одного сообщения (`GET /messages/{messageId}`)

**Метод API:** `GET /messages/{messageId}`  
**Интерфейс:** `client.Messages().GetMessage(ctx, messageID)`

```go
msg, err := client.Messages().GetMessage(ctx, "abcdef123456")
if err != nil {
    log.Fatalf("get message: %v", err)
}

log.Printf("sender=%v, text=%q", msg.Sender, msg.Text())
```

---

### Редактирование сообщения (`PUT /messages`)

**Метод API:** `PUT /messages`  
**Интерфейс:** `client.Messages().EditMessage(ctx, messageID, body)`

```go
res, err := client.Messages().EditMessage(ctx, "abcdef123456", maxapi.EditMessageBody{
    Text:   "Обновлённый текст сообщения",
    Format: "markdown",
})
if err != nil {
    log.Fatalf("edit message: %v", err)
}

if !res.Success {
    log.Printf("edit failed: %s", res.Message)
}
```

Особенности:

- если `attachments` в теле запроса `null` — вложения не меняются;
- если `attachments` пустой список — все вложения будут удалены.

---

### Удаление сообщения (`DELETE /messages`)

**Метод API:** `DELETE /messages`  
**Интерфейс:** `client.Messages().DeleteMessage(ctx, messageID)`

```go
res, err := client.Messages().DeleteMessage(ctx, "abcdef123456")
if err != nil {
    log.Fatalf("delete message: %v", err)
}

if !res.Success {
    log.Printf("delete failed: %s", res.Message)
}
```

---

### Inline‑клавиатура и callback‑ответ (`POST /answers`)

#### Отправка сообщения с inline‑клавиатурой

**Объект вложения:** `type = "inline_keyboard"`  
**Helper:** `maxapi.NewInlineKeyboardAttachment`

```go
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
    Text: "Сообщение с inline‑клавиатурой",
    Attachments: []maxapi.Attachment{
        maxapi.NewInlineKeyboardAttachment(rows),
    },
    Format: "markdown",
}

_, err := client.Messages().SendToChat(ctx, 123456, body)
if err != nil {
    log.Fatalf("send keyboard message: %v", err)
}
```

#### Обработка нажатия на кнопку и ответ (`POST /answers`)

При нажатии на кнопку типа `callback` бот получает `Update` с полем `Callback`:

```go
router.HandleUpdateType("message_callback", func(ctx context.Context, upd maxapi.Update) error {
    if upd.Callback == nil {
        return nil
    }

    callbackID := upd.Callback.CallbackID

    // Обновляем сообщение и/или отправляем краткое уведомление
    notification := "Кнопка обработана ✅"
    res, err := client.Messages().AnswerCallback(ctx, callbackID, maxapi.CallbackAnswerBody{
        Message: &maxapi.NewMessageBody{
            Text: "Сообщение после нажатия кнопки",
        },
        Notification: &notification,
    })
    if err != nil {
        return err
    }
    if !res.Success {
        log.Printf("answer callback failed: %s", res.Message)
    }
    return nil
})
```

Поле `Callback.CallbackID` соответствует `callback_id` из документации [`POST /answers`](https://dev.max.ru/docs-api/methods/POST/answers).

---

## Chats — групповые чаты

Сервис `Chats()` инкапсулирует методы `chats`.

### Список чатов (`GET /chats`)

**Метод API:** `GET /chats`  
**Интерфейс:** `client.Chats().List(ctx, params)`

```go
res, err := client.Chats().List(ctx, maxapi.ListChatsParams{
    Count:  50,
    Marker: nil, // для первой страницы
})
if err != nil {
    log.Fatalf("list chats: %v", err)
}

for _, ch := range res.Chats {
    log.Printf("chat id=%d title=%q status=%s", ch.ChatID, ch.Title, ch.Status)
}

if res.Marker != nil {
    log.Printf("next marker=%d", *res.Marker)
}
```

---

### Информация о чате (`GET /chats/{chatId}`)

**Метод API:** `GET /chats/{chatId}`  
**Интерфейс:** `client.Chats().Get(ctx, chatID)`

```go
chat, err := client.Chats().Get(ctx, 123456)
if err != nil {
    log.Fatalf("get chat: %v", err)
}

log.Printf("chat %d: title=%q status=%s public=%v",
    chat.ChatID, chat.Title, chat.Status, chat.IsPublic,
)
```

---

### Изменение информации о чате (`PATCH /chats/{chatId}`)

**Метод API:** `PATCH /chats/{chatId}`  
**Интерфейс:** `client.Chats().Update(ctx, chatID, body)`

```go
notify := true
chat, err := client.Chats().Update(ctx, 123456, maxapi.ChatUpdateBody{
    Title:  "Новое название чата",
    Notify: &notify,
})
if err != nil {
    log.Fatalf("update chat: %v", err)
}

log.Printf("updated chat title=%q", chat.Title)
```

Поле `Icon` в `ChatUpdateBody` оставлено как `any`, чтобы можно было передать объект, совместимый с `PhotoAttachmentRequestPayload` из доки MAX (описание иконки чата).

---

### Удаление чата (`DELETE /chats/{chatId}`)

**Метод API:** `DELETE /chats/{chatId}`  
**Интерфейс:** `client.Chats().Delete(ctx, chatID)`

```go
res, err := client.Chats().Delete(ctx, 123456)
if err != nil {
    log.Fatalf("delete chat: %v", err)
}
if !res.Success {
    log.Printf("delete chat failed: %s", res.Message)
}
```

---

## Subscriptions — подписки и обновления

Сервис `Subscriptions()` покрывает:

- long polling `GET /updates`;
- управление Webhook‑подписками `GET/POST/DELETE /subscriptions`.

### Long polling (`GET /updates`) и `LongPoller`

Для разработки/тестирования удобно использовать long polling.

```go
// Router с автоподдержкой упоминаний бота (@username /start).
router := maxapi.NewRouterForClient(ctx, client)

router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Привет! Я бот на Go для MAX.",
            Format: "markdown",
        })
    return err
})

poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})

if err := poller.Run(ctx); err != nil {
    log.Fatal(err)
}
```

`LongPoller` сам:

- делает запросы `GET /updates` с `marker`;
- передаёт каждое `Update` в `handler`;
- обновляет `marker`, чтобы не получать уже обработанные события.

Если нужно более тонко управлять параметрами, можно вызывать `GetUpdates` напрямую:

```go
res, err := client.Subscriptions().GetUpdates(ctx, maxapi.UpdatesParams{
    Limit:   100,
    Timeout: 30,
    Marker:  nil,
    Types:   []string{"message_created", "message_callback"},
})
if err != nil {
    log.Fatalf("get updates: %v", err)
}
```

---

### Webhook‑подписки (`GET /subscriptions`, `POST /subscriptions`, `DELETE /subscriptions`)

#### Получение списка подписок (`GET /subscriptions`)

```go
subs, err := client.Subscriptions().ListSubscriptions(ctx)
if err != nil {
    log.Fatalf("list subscriptions: %v", err)
}

for _, s := range subs {
    log.Printf("subscription url=%s types=%v", s.URL, s.UpdateTypes)
}
```

#### Создание подписки (`POST /subscriptions`)

```go
res, err := client.Subscriptions().CreateSubscription(ctx, maxapi.CreateSubscriptionBody{
    URL:         "https://your-domain.com/webhook",
    UpdateTypes: []string{"message_created", "bot_started"},
    Secret:      "your_secret_123",
})
if err != nil {
    log.Fatalf("create subscription: %v", err)
}
if !res.Success {
    log.Printf("create subscription failed: %s", res.Message)
}
```

#### Удаление подписки (`DELETE /subscriptions`)

```go
res, err := client.Subscriptions().DeleteSubscription(ctx, "https://your-domain.com/webhook")
if err != nil {
    log.Fatalf("delete subscription: %v", err)
}
if !res.Success {
    log.Printf("delete subscription failed: %s", res.Message)
}
```

---

## Uploads — загрузка файлов (`POST /uploads`) и информация о видео

Сервис `Uploads()` отвечает за инициализацию загрузки медиа и получение информации о видео.

### Получение URL для загрузки (`POST /uploads`)

**Метод API:** `POST /uploads?type=image|video|audio|file`  
**Интерфейс:** `client.Uploads().InitUpload(ctx, uploadType)`

```go
upload, err := client.Uploads().InitUpload(ctx, "video")
if err != nil {
    log.Fatalf("init upload: %v", err)
}

log.Printf("upload url=%s token=%s", upload.URL, upload.Token)
```

Дальше нужно:

1. Отправить файл по `upload.URL` (см. примеры `multipart/form-data` в доке MAX).  
2. Получить `token` из ответа загрузчика.  
3. Передать `token` в `NewMessageBody.Attachments`:

```go
videoToken := "<video_token_from_upload>"

body := maxapi.NewMessageBody{
    Text: "Сообщение с видео",
    Attachments: []maxapi.Attachment{
        {
            Type: "video",
            Payload: map[string]string{
                "token": videoToken,
            },
        },
    },
}

_, err = client.Messages().SendToChat(ctx, 123456, body)
if err != nil {
    log.Fatalf("send video message: %v", err)
}
```

### Информация о видео (`GET /videos/{videoToken}`)

**Метод API:** `GET /videos/{videoToken}`  
**Интерфейс:** `client.Uploads().GetVideoInfo(ctx, videoToken)`

```go
info, err := client.Uploads().GetVideoInfo(ctx, videoToken)
if err != nil {
    log.Fatalf("get video info: %v", err)
}

log.Printf("video %s: %dx%d, duration=%ds",
    info.Token, info.Width, info.Height, info.Duration,
)
```

---

## Router — маршрутизация сообщений

`Router` упрощает обработку событий `message_created` и `message_callback`:

- по типу обновления;
- по текстовым командам (`/start`, `/help` и т.д.);
- по смайликам/подстрокам;
- с обработчиком по умолчанию.

### Базовый пример Router

```go
// Router с автоподдержкой упоминаний бота (@username /start).
router := maxapi.NewRouterForClient(ctx, client)

// Команда /start
router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Привет! Я демонстрационный бот.",
            Format: "markdown",
        })
    return err
})

// Реакция на смайлик 😀
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

Использование с `LongPoller`:

```go
poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})
if err := poller.Run(ctx); err != nil {
    log.Fatal(err)
}
```

---

## Примеры из каталога `examples/`

- `examples/basic-bot` — минимальный бот, отвечающий на `/start`.
- `examples/router` — использование `Router`, команд, смайликов и inline‑клавиатуры.
- `examples/files` — обработка сообщений с вложениями (логирование `attachments`).

Рекомендуется начать с запуска `examples/basic-bot`, затем изучить `router` и `files` для более продвинутых сценариев.
