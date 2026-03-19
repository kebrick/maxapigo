## Uploads — загрузка файлов и видео

Сервис `client.Uploads()` помогает:

- получить URL для загрузки файла (`POST /uploads`);
- узнать информацию о видео (`GET /videos/{videoToken}`).

Полные детали — в разделе `upload` доки MAX.

---

### Получить URL для загрузки (`POST /uploads`)

Параметр `type=photo` в API **больше не поддерживается** — используйте `"image"`.

```go
upload, err := client.Uploads().InitUpload(ctx, "video") // "image", "audio" или "file"
if err != nil {
    log.Fatalf("init upload: %v", err)
}

log.Printf("upload url=%s token=%s", upload.URL, upload.Token)
```

Дальше:

1. по `upload.URL` загружаем файл через `multipart/form-data` (см. примеры в доке MAX);
2. из ответа загрузчика получаем `token` (особенно важно для `video` и `audio`);
3. передаём `token` в `attachments` при отправке сообщения:

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

---

### Информация о видео (`GET /videos/{videoToken}`)

```go
info, err := client.Uploads().GetVideoInfo(ctx, videoToken)
if err != nil {
    log.Fatalf("get video info: %v", err)
}

log.Printf("video %s: %dx%d, duration=%ds",
    info.Token, info.Width, info.Height, info.Duration,
)
```

Это полезно, если нужно:

- отрисовать миниатюру видео;
- отображать длительность/размер видео в интерфейсе;
- валидировать, что видео доступно.

