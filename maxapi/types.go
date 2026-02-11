package maxapi

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

// InlineKeyboardButton описывает кнопку inline-клавиатуры.
// См. раздел "Клавиатура" в документации:
// https://dev.max.ru/docs-api
type InlineKeyboardButton struct {
	// Type — тип кнопки: "callback", "link", "open_app",
	// "request_geo_location", "request_contact", "message" и т.п.
	Type string `json:"type"`
	// Text — подпись на кнопке (для типов, где она отображается).
	Text string `json:"text,omitempty"`
	// Payload — произвольные данные, которые вернутся в событии
	// message_callback для кнопки типа "callback".
	Payload string `json:"payload,omitempty"`
	// URL — ссылка для кнопки типа "link".
	URL string `json:"url,omitempty"`
}

// InlineKeyboardPayload — payload вложения с type="inline_keyboard".
type InlineKeyboardPayload struct {
	// Buttons — массив строк с кнопками, до 7 в строке.
	Buttons [][]InlineKeyboardButton `json:"buttons"`
}

// NewInlineKeyboardAttachment создаёт вложение типа inline-клавиатуры.
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
	Body      *MessageBody `json:"body,omitempty"`
	// link, stat, url и т.п. можно добавить позже при необходимости.
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

// NewMessageBody — тело запроса для отправки сообщения.
// Полностью следует объекту NewMessageBody из документации.
// https://dev.max.ru/docs-api/objects/NewMessageBody
type NewMessageBody struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
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


