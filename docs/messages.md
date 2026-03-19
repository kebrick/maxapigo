## Messages — работа с сообщениями

Сервис `client.Messages()` оборачивает методы раздела `messages` из доки MAX.

---

### Отправить сообщение (`POST /messages`)

Ответ API оборачивает сообщение в поле `message` — библиотека разбирает это автоматически.

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

В личку по `user_id` (query `user_id` в API):

```go
_, err := client.Messages().SendToUser(ctx, 5236658, maxapi.NewMessageBody{
    Text: "Сообщение в диалог",
})
```

Все query‑параметры отправки (`chat_id`, `user_id`, `disable_link_preview`) — через `SendMessage`:

```go
chatID := int64(123456)
// По доке MAX: disable_link_preview=false — без превью ссылок (проверьте на своём окружении).
noLinkPreview := false
_, err := client.Messages().SendMessage(ctx, maxapi.SendMessageParams{
    ChatID:             &chatID,
    DisableLinkPreview: &noLinkPreview,
}, maxapi.NewMessageBody{Text: "https://example.com"})
```

Коротко про `NewMessageBody`:

- `Text` — текст сообщения (до 4000 символов);
- `Attachments` — вложения (файлы, картинки, видео, inline‑клавиатура и т.д.);
- `Link` — объект ответа/пересылки (`NewMessageLink` в доке MAX), тип в Go — `any`;
- `Notify` — слать ли уведомление участникам;
- `Format` — `markdown` или `html`.

---

### Получить сообщения чата (`GET /messages`)

```go
chatID := int64(123456)

messages, err := client.Messages().ListMessages(ctx, maxapi.ListMessagesParams{
    ChatID: &chatID,
    Count:  50, // максимум 100
})
if err != nil {
    log.Fatalf("list messages: %v", err)
}

for _, m := range messages {
    log.Printf("message at %d: %q", m.Timestamp, m.Text())
}
```

Поддерживаемые параметры:

- `ChatID` **или** `MessageIDs` — одно из двух обязательно;
- `From` / `To` — фильтр по времени (Unix timestamp);
- `Count` — сколько сообщений вернуть (1–100).

---

### Получить одно сообщение (`GET /messages/{messageId}`)

```go
msg, err := client.Messages().GetMessage(ctx, "abcdef123456")
if err != nil {
    log.Fatalf("get message: %v", err)
}

log.Printf("sender=%v, text=%q", msg.Sender, msg.Text())
```

---

### Редактировать сообщение (`PUT /messages`)

```go
res, err := client.Messages().EditMessage(ctx, "abcdef123456", maxapi.EditMessageBody{
    Text:   "Новый текст",
    Format: "markdown",
})
if err != nil {
    log.Fatalf("edit message: %v", err)
}

if !res.Success {
    log.Printf("edit failed: %s", res.Message)
}
```

Замечания:

- `attachments: null` — вложения остаются как были;
- `attachments: []` — вложения удаляются.

---

### Удалить сообщение (`DELETE /messages`)

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

### Inline‑клавиатура (`inline_keyboard`)

#### Отправка сообщения с inline‑клавиатурой

Для inline‑клавиатуры есть типизированные helper‑ы:

- кнопки:
  - `NewInlineLinkButton(text, url string)`
  - `NewInlineCallbackButton(text, payload string)`
  - `NewInlineRequestGeoButton(text string, quick bool)`
  - `NewInlineRequestContactButton(text string)`
  - `NewInlineOpenAppButton(text, webApp string, contactID *int64, payload string)`
  - `NewInlineMessageButton(text string)`
- вложение:
  - `NewInlineKeyboardAttachment(rows [][]InlineKeyboardButton)`

Пример:

```go
rows := [][]maxapi.InlineKeyboardButton{
    {
        maxapi.NewInlineCallbackButton("Жми", "button1_pressed"),
        maxapi.NewInlineLinkButton("Сайт", "https://example.com"),
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

#### Обработка callback‑события и ответ (`POST /answers`)

```go
router.HandleUpdateType(maxapi.UpdateTypeMessageCallback, func(ctx context.Context, upd maxapi.Update) error {
    if upd.Callback == nil {
        return nil
    }

    callbackID := upd.Callback.CallbackID

    notification := "Кнопка нажата ✅"
    res, err := client.Messages().AnswerCallback(ctx, callbackID, maxapi.CallbackAnswerBody{
        Message: &maxapi.NewMessageBody{
            Text: "Спасибо за нажатие кнопки",
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

