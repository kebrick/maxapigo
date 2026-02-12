## Subscriptions — подписки и обновления

Сервис `client.Subscriptions()` покрывает:

- long polling (`GET /updates`);
- Webhook‑подписки (`GET/POST/DELETE /subscriptions`).

---

### Long polling (`GET /updates`) через `LongPoller`

Простой бот, который отвечает на `/start`:

```go
// Router с автоматической поддержкой упоминаний бота (@username /start).
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

`LongPoller`:

- сам вызывает `GET /updates` с `marker`;
- передаёт каждое `Update` в обработчик;
- запоминает новый `marker`.

Если нужно больше контроля, можно работать напрямую с `GetUpdates`:

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

### Webhook‑подписки

Webhook встроен в API MAX и удобен для production‑окружения.

#### Список подписок (`GET /subscriptions`)

```go
subs, err := client.Subscriptions().ListSubscriptions(ctx)
if err != nil {
    log.Fatalf("list subscriptions: %v", err)
}

for _, s := range subs {
    log.Printf("subscription url=%s types=%v", s.URL, s.UpdateTypes)
}
```

#### Создать подписку (`POST /subscriptions`)

```go
res, err := client.Subscriptions().CreateSubscription(ctx, maxapi.CreateSubscriptionBody{
    URL:         "https://your-domain.com/max/webhook",
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

#### Удалить подписку (`DELETE /subscriptions`)

```go
res, err := client.Subscriptions().DeleteSubscription(ctx, "https://your-domain.com/max/webhook")
if err != nil {
    log.Fatalf("delete subscription: %v", err)
}
if !res.Success {
    log.Printf("delete subscription failed: %s", res.Message)
}
```

