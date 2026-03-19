package maxapi

import "strings"

// Типы событий long polling / webhook (см. GET /updates и объекты Update в документации MAX).
const (
	UpdateTypeMessageCreated  = "message_created"
	UpdateTypeMessageCallback = "message_callback"
	UpdateTypeBotStarted      = "bot_started"
)

// Базовые модели соответствуют объектам из документации:
// https://dev.max.ru/docs-api/objects/User
// https://dev.max.ru/docs-api/objects/Chat
// https://dev.max.ru/docs-api/objects/Message
// https://dev.max.ru/docs-api/objects/NewMessageBody
// https://dev.max.ru/docs-api/objects/Update

// User описывает пользователя/бота.
type User struct {
	UserID           int64  `json:"user_id"`
	Name             string `json:"name"`
	Username         string `json:"username"`
	IsBot            bool   `json:"is_bot"`
	LastActivityTime int64  `json:"last_activity_time"`
}

// Chat описывает групповой чат.
type Chat struct {
	ChatID          int64  `json:"chat_id"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	Title           string `json:"title"`
	LastEventTime   int64  `json:"last_event_time"`
	ParticipantsCnt int32  `json:"participants_count"`
	IsPublic        bool   `json:"is_public"`
	Link            string `json:"link"`
	Description     string `json:"description"`
	// Остальные поля из объекта Chat (icon, participants, owner_id и т.п.)
	// могут быть добавлены по мере необходимости.
}

// Recipient описывает получателя сообщения (диалог или чат).
type Recipient struct {
	ChatID   int64  `json:"chat_id"`
	ChatType string `json:"chat_type"`
	UserID   int64  `json:"user_id"`
}

// Attachment — вложение сообщения (в т.ч. inline-клавиатура).
//
// В MAX вложения могут быть разного типа:
//   - "inline_keyboard" — inline-клавиатура под сообщением
//   - "file", "image", "video", "sticker" и т.д. — медиа и файлы
//
// Поле Payload оставляем обобщённым, чтобы можно было
// как отправлять типизированные структуры, так и работать
// с произвольным JSON, если понадобятся редкие типы вложений.
type Attachment struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

// ImageAttachmentPayload — payload для вложения type="image".
// См. объект PhotoAttachmentPayload в документации MAX.
type ImageAttachmentPayload struct {
	PhotoID int64  `json:"photo_id,omitempty"`
	Token   string `json:"token,omitempty"`
	URL     string `json:"url,omitempty"`
}

// NewImageAttachment создаёт вложение с изображением.
func NewImageAttachment(p ImageAttachmentPayload) Attachment {
	return Attachment{
		Type:    "image",
		Payload: p,
	}
}

// MediaAttachmentPayload — общий payload для type="video" и type="audio".
// См. MediaAttachmentPayload в документации MAX.
type MediaAttachmentPayload struct {
	URL          string                 `json:"url,omitempty"`
	Token        string                 `json:"token,omitempty"`
	Thumbnail    *ImageAttachmentPayload `json:"thumbnail,omitempty"`
	Width        *int                   `json:"width,omitempty"`
	Height       *int                   `json:"height,omitempty"`
	Duration     *int                   `json:"duration,omitempty"`
	Transcription string                `json:"transcription,omitempty"` // только для audio
}

// NewVideoAttachment создаёт вложение с видео.
func NewVideoAttachment(p MediaAttachmentPayload) Attachment {
	return Attachment{
		Type:    "video",
		Payload: p,
	}
}

// NewAudioAttachment создаёт вложение с аудио.
func NewAudioAttachment(p MediaAttachmentPayload) Attachment {
	return Attachment{
		Type:    "audio",
		Payload: p,
	}
}

// FileAttachmentPayload — payload для type="file".
type FileAttachmentPayload struct {
	URL      string `json:"url,omitempty"`
	Token    string `json:"token,omitempty"`
	Filename string `json:"filename,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// NewFileAttachment создаёт вложение с файлом.
func NewFileAttachment(p FileAttachmentPayload) Attachment {
	return Attachment{
		Type:    "file",
		Payload: p,
	}
}

// StickerAttachmentPayload — payload для type="sticker".
type StickerAttachmentPayload struct {
	URL    string `json:"url,omitempty"`
	Code   string `json:"code,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// NewStickerAttachment создаёт вложение со стикером.
func NewStickerAttachment(p StickerAttachmentPayload) Attachment {
	return Attachment{
		Type:    "sticker",
		Payload: p,
	}
}

// ContactAttachmentPayload — payload для type="contact".
type ContactAttachmentPayload struct {
	VCFInfo string `json:"vcf_info,omitempty"`
	MaxInfo *User  `json:"max_info,omitempty"`
}

// NewContactAttachment создаёт вложение с контактом.
func NewContactAttachment(p ContactAttachmentPayload) Attachment {
	return Attachment{
		Type:    "contact",
		Payload: p,
	}
}

// ShareAttachmentPayload — payload для type="share".
type ShareAttachmentPayload struct {
	URL       string `json:"url,omitempty"`
	Token     string `json:"token,omitempty"`
	Title     string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
}

// NewShareAttachment создаёт вложение для предпросмотра ссылки.
func NewShareAttachment(p ShareAttachmentPayload) Attachment {
	return Attachment{
		Type:    "share",
		Payload: p,
	}
}

// InlineKeyboardButton описывает кнопку inline-клавиатуры.
// См. раздел "Клавиатура" в документации:
// https://dev.max.ru/docs-api
type InlineKeyboardButton struct {
	// Type — тип кнопки: "callback", "link", "open_app",
	// "request_geo_location", "request_contact", "message" и т.п.
	Type string `json:"type"`

	// Text — подпись на кнопке (для большинства типов).
	Text string `json:"text,omitempty"`

	// URL — ссылка для кнопки типа "link".
	URL string `json:"url,omitempty"`

	// Payload — произвольные данные, которые вернутся в событии
	// message_callback для кнопки типа "callback" или будут
	// переданы в мини-приложение для open_app.
	Payload string `json:"payload,omitempty"`

	// Quick — для request_geo_location: если true, отправляет
	// геопозицию без дополнительного подтверждения.
	Quick *bool `json:"quick,omitempty"`

	// WebApp — имя (username) или ссылка на бота, чьё мини-приложение
	// нужно открыть (для кнопки типа "open_app").
	WebApp string `json:"web_app,omitempty"`

	// ContactID — идентификатор бота, чьё мини-приложение
	// нужно запустить (для кнопки типа "open_app").
	ContactID *int64 `json:"contact_id,omitempty"`
}

// Удобные конструкторы для разных типов кнопок.

// NewInlineLinkButton создаёт кнопку типа "link".
func NewInlineLinkButton(text, url string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type: "link",
		Text: text,
		URL:  url,
	}
}

// NewInlineCallbackButton создаёт кнопку типа "callback".
func NewInlineCallbackButton(text, payload string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type:    "callback",
		Text:    text,
		Payload: payload,
	}
}

// NewInlineRequestGeoButton создаёт кнопку типа "request_geo_location".
func NewInlineRequestGeoButton(text string, quick bool) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type:  "request_geo_location",
		Text:  text,
		Quick: &quick,
	}
}

// NewInlineRequestContactButton создаёт кнопку типа "request_contact".
func NewInlineRequestContactButton(text string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type: "request_contact",
		Text: text,
	}
}

// NewInlineOpenAppButton создаёт кнопку типа "open_app".
// webApp — username/URL бота с мини-приложением,
// contactID — опциональный ID бота, payload — initData.
func NewInlineOpenAppButton(text, webApp string, contactID *int64, payload string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type:      "open_app",
		Text:      text,
		WebApp:    webApp,
		ContactID: contactID,
		Payload:   payload,
	}
}

// NewInlineMessageButton создаёт кнопку типа "message".
// При нажатии текст кнопки будет отправлен в чат от лица пользователя.
func NewInlineMessageButton(text string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Type: "message",
		Text: text,
	}
}

// InlineKeyboardPayload — payload вложения с type="inline_keyboard".
type InlineKeyboardPayload struct {
	// Buttons — массив строк с кнопками, до 7 в строке.
	Buttons [][]InlineKeyboardButton `json:"buttons"`
}

// NewInlineKeyboardAttachment создаёт вложение типа inline-клавиатуры.
// По документации MAX: до 210 кнопок, до 30 рядов, до 7 кнопок в ряду
// (до 3 в ряду для link, open_app, request_geo_location, request_contact).
//
// Пример:
//
//	rows := [][]maxapi.InlineKeyboardButton{
//		{
//			{Type: "callback", Text: "Нажми меня", Payload: "btn_1"},
//		},
//	}
//	body := maxapi.NewMessageBody{
//		Text: "Сообщение с кнопкой",
//		Attachments: []maxapi.Attachment{
//			maxapi.NewInlineKeyboardAttachment(rows),
//		},
//	}
func NewInlineKeyboardAttachment(rows [][]InlineKeyboardButton) Attachment {
	return Attachment{
		Type: "inline_keyboard",
		Payload: InlineKeyboardPayload{
			Buttons: rows,
		},
	}
}

// MessageBody — содержимое сообщения.
type MessageBody struct {
	MID         string       `json:"mid"`
	Seq         int64        `json:"seq"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	// В дальнейшем сюда можно добавить дополнительные поля (stat, link и т.п.).
}

// Message описывает сообщение согласно документации.
// https://dev.max.ru/docs-api/objects/Message
type Message struct {
	Sender    *User        `json:"sender,omitempty"`
	Recipient Recipient    `json:"recipient"`
	Timestamp int64        `json:"timestamp"`
	Link      any          `json:"link,omitempty"` // LinkedMessage: переслано / ответ
	Body      *MessageBody `json:"body,omitempty"`
	Stat      any          `json:"stat,omitempty"` // статистика (каналы)
	URL       string       `json:"url,omitempty"`  // публичная ссылка на пост в канале
}

// Text возвращает текст сообщения (если он есть).
func (m *Message) Text() string {
	if m == nil || m.Body == nil {
		return ""
	}
	return m.Body.Text
}

// ChatID возвращает chat_id получателя как строку.
func (m *Message) ChatID() string {
	if m == nil {
		return ""
	}
	return int64ToString(m.Recipient.ChatID)
}

// IsFromChannel возвращает true, если сообщение относится к каналу (recipient.chat_type).
// У постов в канале поле sender часто отсутствует; чтобы получать такие апдейты, бот должен быть администратором канала.
func (m *Message) IsFromChannel() bool {
	if m == nil {
		return false
	}
	t := strings.TrimSpace(m.Recipient.ChatType)
	return strings.EqualFold(t, "channel")
}

// NewMessageBody — тело запроса для отправки сообщения.
// Полностью следует объекту NewMessageBody из документации.
// https://dev.max.ru/docs-api/objects/NewMessageBody
type NewMessageBody struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Link        any          `json:"link,omitempty"` // NewMessageLink — ответ / пересылка
	Notify      *bool        `json:"notify,omitempty"`
	Format      string       `json:"format,omitempty"` // "markdown" или "html"
}

// Update описывает обновление (event) для long polling / webhook.
// См. https://dev.max.ru/docs-api/objects/Update
type Update struct {
	UpdateType string    `json:"update_type"`
	Timestamp  int64     `json:"timestamp"`
	Message    *Message  `json:"message,omitempty"`
	Callback   *Callback `json:"callback,omitempty"`
	UserLocale string    `json:"user_locale,omitempty"`
}

// Callback — данные callback-события (нажатие на inline-кнопку).
// Используется, в том числе, для получения callback_id,
// который нужен методу POST /answers.
type Callback struct {
	CallbackID string `json:"callback_id"`
}

// Subscription описывает подписку Webhook (GET /subscriptions).
type Subscription struct {
	URL         string   `json:"url"`
	UpdateTypes []string `json:"update_types,omitempty"`
	Secret      string   `json:"secret,omitempty"`
}

// UploadInitResponse — результат вызова POST /uploads.
type UploadInitResponse struct {
	URL   string `json:"url"`
	Token string `json:"token,omitempty"`
}

// VideoInfo — информация о видео (GET /videos/{videoToken}).
type VideoInfo struct {
	Token     string   `json:"token"`
	URLs      any      `json:"urls,omitempty"`      // оставляем как any, чтобы не тянуть все вложенные типы
	Thumbnail any      `json:"thumbnail,omitempty"` // payload миниатюры
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Duration  int      `json:"duration"`
}


