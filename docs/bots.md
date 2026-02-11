## Bots — информация о боте

Раздел соответствует методу [`GET /me`](https://dev.max.ru/docs-api/methods/GET/me).

### Получить информацию о боте

```go
ctx := context.Background()

me, err := client.Bots().Me(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf(
    "bot name=%s id=%d username=%s isBot=%v",
    me.Name, me.UserID, me.Username, me.IsBot,
)
```

Зачем это нужно:

- проверить, что токен корректен;
- вывести базовую информацию о боте при старте сервиса;
- записать данные в метрики или логи.

