## Обзор maxapigo

`maxapigo` — это клиентская библиотека на Go для работы с API мессенджера MAX.

- Официальная документация API: [`https://dev.max.ru/docs-api`](https://dev.max.ru/docs-api).
- Рекомендуемый лимит: до **30 rps** на `platform-api.max.ru`; токен передаётся только в заголовке `Authorization` (query с токеном не поддерживается).
- Библиотека берёт на себя:
  - базовый URL `https://platform-api.max.ru`;
  - заголовок `Authorization: <token>`;
  - разбор/сборку JSON‑ответов;
  - удобные сервисы для ботов, сообщений, чатов, подписок и загрузки файлов.

### Установка

```bash
go get github.com/kebrick/maxapigo
```

### Создание клиента

```go
import (
    "context"
    "log"

    "github.com/kebrick/maxapigo/maxapi"
)

func main() {
    ctx := context.Background()

    client, err := maxapi.NewClient(maxapi.Config{
        Token:  "<BOT_TOKEN>",            // токен бота из MAX
        Logger: maxapi.NewStdLogger(nil), // можно отключить, если не нужен
        Debug:  true,                     // включить подробные логи
    })
    if err != nil {
        log.Fatalf("create client: %v", err)
    }

    // пример: проверим, что токен рабочий
    me, err := client.Bots().Me(ctx)
    if err != nil {
        log.Fatalf("me: %v", err)
    }
    log.Printf("bot: %s (%d)", me.Name, me.UserID)
}
```

### Основные сервисы

- `client.Bots()` — информация о боте.
- `client.Messages()` — отправка, получение, редактирование и удаление сообщений.
- `client.Chats()` — получение и изменение групповых чатов.
- `client.Subscriptions()` — long polling и Webhook‑подписки.
- `client.Uploads()` — инициализация загрузки файлов и информация о видео.
- `maxapi.Router` и `maxapi.LongPoller` — удобная обвязка для бота.

Детали по каждому разделу — в отдельных файлах:

- [`bots.md`](./bots.md)
- [`messages.md`](./messages.md)
- [`chats.md`](./chats.md)
- [`subscriptions.md`](./subscriptions.md)
- [`uploads.md`](./uploads.md)
- [`router.md`](./router.md)

