## maxapigo

Библиотека-клиент на Go для работы с API мессенджера MAX.  
Основана на документации [`https://dev.max.ru/docs-api`](https://dev.max.ru/docs-api).

### Подключение

```bash
go get github.com/kebrick/maxapigo
```


### Конфигурация клиента

```go
client, err := maxapi.NewClient(maxapi.Config{
    Token:  "<BOT_TOKEN>",              // токен бота из раздела «Интеграция» → «Получить токен»
    Logger: maxapi.NewStdLogger(nil),   // опционально: логирование запросов
    Debug:  true,                       // подробные логи (метод, URL, статус, тело ответа)
})
```

Клиент инкапсулирует:

- **HTTP-клиент** с базовым URL `https://platform-api.max.ru`
- **Авторизацию** через заголовок `Authorization: <token>` (см. [доку MAX](https://dev.max.ru/docs-api))
- **Сервисы**:
  - `Bots()` — работа с ботом (`GET /me`)
  - `Messages()` — отправка сообщений (`POST /messages`)
  - `Subscriptions()` — получение обновлений (`GET /updates`, long polling)

### Получение информации о боте

```go
ctx := context.Background()

me, err := client.Bots().Me(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("Bot: %s (id=%d)", me.Name, me.UserID)
```

### Отправка сообщения

Отправка сообщения в чат по `chat_id` (см. [`POST /messages`](https://dev.max.ru/docs-api/methods/POST/messages)):

```go
msg, err := client.Messages().SendToChat(ctx, 123456,
    maxapi.NewMessageBody{
        Text:   "Привет из Go!",
        Format: "markdown",
    })
if err != nil {
    log.Fatal(err)
}

log.Printf("sent at %d, text=%q", msg.Timestamp, msg.Text())
```

### Long polling (`GET /updates`)

Для разработки и тестирования библиотека предоставляет удобную обёртку над методом  
[`GET /updates`](https://dev.max.ru/docs-api/methods/GET/updates), который использует long polling.

```go
router := maxapi.NewRouter()

// Реакция на команду /start.
router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Привет! Я бот на Go для MAX.",
            Format: "markdown",
        })
    return err
})

// Реакция на смайлик 😀 в сообщении.
router.HandleEmoji("😀", func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text:   "Вижу хороший настрой! 😀",
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

- делает запросы `GET /updates` с параметром `marker`
- передаёт каждое `Update` в указанную функцию-обработчик
- обновляет `marker`, чтобы не повторять уже обработанные события

### Маршрутизация сообщений (Router)

Объект `Router` упрощает обработку событий `message_created`:

- **По типу обновления** (`update_type`, напр. `message_created`, `message_callback`)
- **По текстовым командам** (строки вида `/start`, `/help`)
- **По смайликам и произвольным подстрокам** в тексте
- **Обработчик по умолчанию** для любых сообщений

Примеры:

```go
router := maxapi.NewRouter()

// Хендлер для конкретной команды.
router.HandleCommand("/start", startHandler)

// Хендлер для смайлика.
router.HandleEmoji("🔥", fireHandler)

// Хендлер по типу обновления (например, message_created).
router.HandleUpdateType("message_created", func(ctx context.Context, upd maxapi.Update) error {
    // общий лог/метрика по всем новым сообщениям
    return nil
})

// Хендлер по умолчанию — срабатывает, если ни одна команда/смайлик не подошли.
router.HandleDefaultMessage(func(ctx context.Context, msg *maxapi.Message) error {
    _, err := client.Messages().SendToChat(ctx, msg.Recipient.ChatID,
        maxapi.NewMessageBody{
            Text: "Я пока не знаю эту команду. Напишите /start.",
        })
    return err
})
```

### Примеры

В каталоге `examples/` находятся готовые примеры:

- `examples/basic-bot` — минимальный бот, отвечающий на `/start`
- `examples/router` — использование `Router` для команд и смайликов
- `examples/files` — заготовка для обработки вложений (файлы, картинки, видео, стикеры)

См. файлы внутри `examples/` для подробных сценариев использования.

g### Подробная документация

В каталоге `docs/` находится более развёрнутая документация по библиотеке:

- `docs/overview.md` — установка, создание клиента, обзор сервисов.
- `docs/bots.md` — работа с `Bots()` и методом `GET /me`.
- `docs/messages.md` — отправка, чтение, редактирование и удаление сообщений, inline‑клавиатура, callback‑ответы.
- `docs/chats.md` — методы для групповых чатов (`/chats`).
- `docs/subscriptions.md` — long polling (`GET /updates`) и Webhook‑подписки (`/subscriptions`).
- `docs/uploads.md` — загрузка файлов (`/uploads`) и получение информации о видео.
- `docs/router.md` — Router и LongPoller, включая middleware (логирование, проверки доступа и т.п.).

