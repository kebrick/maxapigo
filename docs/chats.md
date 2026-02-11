## Chats — групповые чаты

Сервис `client.Chats()` оборачивает методы `GET/ PATCH/ DELETE /chats`.

---

### Список чатов (`GET /chats`)

```go
res, err := client.Chats().List(ctx, maxapi.ListChatsParams{
    Count:  50,
    Marker: nil, // первая страница
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

Полезно, чтобы пройтись по всем групповым чатам, где бот когда‑либо участвовал.

---

### Информация о чате (`GET /chats/{chatId}`)

```go
chat, err := client.Chats().Get(ctx, 123456)
if err != nil {
    log.Fatalf("get chat: %v", err)
}

log.Printf(
    "chat %d: title=%q status=%s public=%v",
    chat.ChatID, chat.Title, chat.Status, chat.IsPublic,
)
```

---

### Изменить чат (`PATCH /chats/{chatId}`)

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

- `Title` — новое имя чата;
- `Icon` — объект иконки (структура из доки MAX, можно передать как `map[string]any`);
- `Pin` — `message_id` сообщения, которое нужно закрепить;
- `Notify` — расслать ли системное уведомление об изменениях.

---

### Удалить чат (`DELETE /chats/{chatId}`)

```go
res, err := client.Chats().Delete(ctx, 123456)
if err != nil {
    log.Fatalf("delete chat: %v", err)
}

if !res.Success {
    log.Printf("delete chat failed: %s", res.Message)
}
```

