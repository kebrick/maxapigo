package maxapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// BotService описывает операции с ботом.
type BotService interface {
	// Me возвращает информацию о боте (GET /me).
	Me(ctx context.Context) (*User, error)
}

// ChatService описывает операции с групповыми чатами.
type ChatService interface {
	// List возвращает список групповых чатов (GET /chats).
	List(ctx context.Context, params ListChatsParams) (ListChatsResult, error)
	// Get возвращает информацию о конкретном групповом чате (GET /chats/{chatId}).
	Get(ctx context.Context, chatID int64) (*Chat, error)
	// Update изменяет информацию о групповом чате (PATCH /chats/{chatId}).
	Update(ctx context.Context, chatID int64, body ChatUpdateBody) (*Chat, error)
	// Delete удаляет групповой чат (DELETE /chats/{chatId}).
	Delete(ctx context.Context, chatID int64) (SimpleResult, error)
}

// MessageService описывает операции с сообщениями.
type MessageService interface {
	// Send отправляет новое сообщение (POST /messages) без query-параметров получателя.
	// Обычно нужен SendToChat, SendToUser или SendMessage.
	Send(ctx context.Context, body NewMessageBody) (*Message, error)
	// SendMessage отправляет сообщение с параметрами POST /messages (chat_id, user_id, disable_link_preview).
	SendMessage(ctx context.Context, params SendMessageParams, body NewMessageBody) (*Message, error)
	// SendToChat отправляет сообщение в указанный чат (chat_id как в API).
	SendToChat(ctx context.Context, chatID int64, body NewMessageBody) (*Message, error)
	// SendToUser отправляет сообщение пользователю по user_id (query user_id).
	SendToUser(ctx context.Context, userID int64, body NewMessageBody) (*Message, error)

	// ListMessages получает сообщения по chat_id или message_ids (GET /messages).
	ListMessages(ctx context.Context, params ListMessagesParams) ([]Message, error)

	// GetMessage получает одно сообщение по его messageId (GET /messages/{messageId}).
	GetMessage(ctx context.Context, messageID string) (*Message, error)

	// EditMessage редактирует сообщение (PUT /messages).
	EditMessage(ctx context.Context, messageID string, body EditMessageBody) (SimpleResult, error)

	// DeleteMessage удаляет сообщение (DELETE /messages).
	DeleteMessage(ctx context.Context, messageID string) (SimpleResult, error)

	// AnswerCallback отправляет ответ на callback-кнопку (POST /answers).
	AnswerCallback(ctx context.Context, callbackID string, body CallbackAnswerBody) (SimpleResult, error)
}

// SubscriptionService описывает работу с подписками и long polling.
type SubscriptionService interface {
	// GetUpdates получает обновления (GET /updates).
	GetUpdates(ctx context.Context, params UpdatesParams) (UpdatesResult, error)

	// ListSubscriptions возвращает список подписок Webhook (GET /subscriptions).
	ListSubscriptions(ctx context.Context) ([]Subscription, error)
	// CreateSubscription создаёт подписку Webhook (POST /subscriptions).
	CreateSubscription(ctx context.Context, body CreateSubscriptionBody) (SimpleResult, error)
	// DeleteSubscription удаляет подписку Webhook по URL (DELETE /subscriptions?url=...).
	DeleteSubscription(ctx context.Context, url string) (SimpleResult, error)
}

// botService — реализация BotService.
type botService struct {
	client *apiClient
}

func (s *botService) Me(ctx context.Context) (*User, error) {
	var user User
	if err := s.client.do(ctx, http.MethodGet, "/me", url.Values{}, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// messageService — реализация MessageService.
type messageService struct {
	client *apiClient
}

// chatService — реализация ChatService.
type chatService struct {
	client *apiClient
}

// SendMessageParams — query-параметры POST /messages (см. dev.max.ru/docs-api/methods/POST/messages).
// Укажите ровно один из ChatID или UserID — кому адресовано сообщение.
type SendMessageParams struct {
	ChatID *int64
	UserID *int64
	// DisableLinkPreview — query disable_link_preview (см. POST /messages).
	// В документации MAX: при значении false превью ссылок не генерируются.
	DisableLinkPreview *bool
}

// postMessageResponseFlexible декодирует ответ POST /messages: по документации поле "message",
// с обратной совместимостью для «плоского» JSON с полями Message.
type postMessageResponseFlexible struct {
	Message *Message
}

func (p *postMessageResponseFlexible) UnmarshalJSON(data []byte) error {
	var wrapped struct {
		M *Message `json:"message"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if wrapped.M != nil {
		p.Message = wrapped.M
		return nil
	}
	var direct Message
	if err := json.Unmarshal(data, &direct); err != nil {
		return err
	}
	p.Message = &direct
	return nil
}

func (s *messageService) Send(ctx context.Context, body NewMessageBody) (*Message, error) {
	return s.SendMessage(ctx, SendMessageParams{}, body)
}

func (s *messageService) SendMessage(ctx context.Context, params SendMessageParams, body NewMessageBody) (*Message, error) {
	q := url.Values{}
	if params.ChatID != nil {
		q.Set("chat_id", int64ToString(*params.ChatID))
	}
	if params.UserID != nil {
		q.Set("user_id", int64ToString(*params.UserID))
	}
	if params.DisableLinkPreview != nil {
		q.Set("disable_link_preview", strconv.FormatBool(*params.DisableLinkPreview))
	}

	var env postMessageResponseFlexible
	if err := s.client.do(ctx, http.MethodPost, "/messages", q, body, &env); err != nil {
		return nil, err
	}
	if env.Message == nil {
		return nil, errors.New("maxapi: POST /messages: пустой ответ или нет поля message")
	}
	return env.Message, nil
}

func (s *messageService) SendToChat(ctx context.Context, chatID int64, body NewMessageBody) (*Message, error) {
	cid := chatID
	return s.SendMessage(ctx, SendMessageParams{ChatID: &cid}, body)
}

func (s *messageService) SendToUser(ctx context.Context, userID int64, body NewMessageBody) (*Message, error) {
	uid := userID
	return s.SendMessage(ctx, SendMessageParams{UserID: &uid}, body)
}

// ListChatsParams — параметры запроса к GET /chats.
type ListChatsParams struct {
	Count  int   // количество чатов (1–100, по умолчанию 50)
	Marker *int64 // маркер для постраничной навигации
}

// ListChatsResult — результат вызова List (GET /chats).
type ListChatsResult struct {
	Chats  []Chat `json:"chats"`
	Marker *int64 `json:"marker"`
}

func (s *chatService) List(ctx context.Context, params ListChatsParams) (ListChatsResult, error) {
	q := url.Values{}
	if params.Count > 0 {
		q.Set("count", intToString(params.Count))
	}
	if params.Marker != nil {
		q.Set("marker", int64ToString(*params.Marker))
	}

	var res ListChatsResult
	if err := s.client.do(ctx, http.MethodGet, "/chats", q, nil, &res); err != nil {
		return ListChatsResult{}, err
	}
	return res, nil
}

func (s *chatService) Get(ctx context.Context, chatID int64) (*Chat, error) {
	path := "/chats/" + int64ToString(chatID)
	var chat Chat
	if err := s.client.do(ctx, http.MethodGet, path, url.Values{}, nil, &chat); err != nil {
		return nil, err
	}
	return &chat, nil
}

// ChatUpdateBody — тело запроса для PATCH /chats/{chatId}.
type ChatUpdateBody struct {
	Icon   any    `json:"icon,omitempty"`  // совместимо с PhotoAttachmentRequestPayload
	Title  string `json:"title,omitempty"` // новое название чата
	Pin    string `json:"pin,omitempty"`   // message_id для закрепления
	Notify *bool  `json:"notify,omitempty"`
}

func (s *chatService) Update(ctx context.Context, chatID int64, body ChatUpdateBody) (*Chat, error) {
	path := "/chats/" + int64ToString(chatID)
	var chat Chat
	if err := s.client.do(ctx, http.MethodPatch, path, url.Values{}, body, &chat); err != nil {
		return nil, err
	}
	return &chat, nil
}

func (s *chatService) Delete(ctx context.Context, chatID int64) (SimpleResult, error) {
	path := "/chats/" + int64ToString(chatID)
	var res SimpleResult
	if err := s.client.do(ctx, http.MethodDelete, path, url.Values{}, nil, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

// ListMessagesParams — параметры запроса к GET /messages.
// См. https://dev.max.ru/docs-api/methods/GET/messages
type ListMessagesParams struct {
	// Либо ChatID, либо MessageIDs (обязателен один из них).
	ChatID *int64

	// Список message_id через запятую в запросе.
	MessageIDs []string

	// Диапазон по времени (Unix timestamp).
	From *int64
	To   *int64

	// Максимальное количество сообщений (1–100, по умолчанию 50).
	Count int
}

// listMessagesResponse — внутренний тип для декодирования ответа GET /messages.
type listMessagesResponse struct {
	Messages []Message `json:"messages"`
}

func (s *messageService) ListMessages(ctx context.Context, params ListMessagesParams) ([]Message, error) {
	q := url.Values{}

	if params.ChatID != nil {
		q.Set("chat_id", int64ToString(*params.ChatID))
	}

	if len(params.MessageIDs) > 0 {
		q.Set("message_ids", strings.Join(params.MessageIDs, ","))
	}

	// Требуется хотя бы один из параметров chat_id или message_ids.
	if q.Get("chat_id") == "" && q.Get("message_ids") == "" {
		return nil, errors.New("maxapi: either ChatID or MessageIDs must be provided")
	}

	if params.From != nil {
		q.Set("from", int64ToString(*params.From))
	}
	if params.To != nil {
		q.Set("to", int64ToString(*params.To))
	}
	if params.Count > 0 {
		q.Set("count", intToString(params.Count))
	}

	var resp listMessagesResponse
	if err := s.client.do(ctx, http.MethodGet, "/messages", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (s *messageService) GetMessage(ctx context.Context, messageID string) (*Message, error) {
	if strings.TrimSpace(messageID) == "" {
		return nil, errors.New("maxapi: messageID is required")
	}

	var msg Message
	path := "/messages/" + messageID
	if err := s.client.do(ctx, http.MethodGet, path, url.Values{}, nil, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// EditMessageBody — тело запроса для редактирования сообщения (PUT /messages).
// См. https://dev.max.ru/docs-api/methods/PUT/messages
type EditMessageBody struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Link        any          `json:"link,omitempty"` // NewMessageLink / ответ на другое сообщение
	Notify      *bool        `json:"notify,omitempty"`
	Format      string       `json:"format,omitempty"`
}

// SimpleResult описывает стандартный ответ success/message,
// используемый в нескольких методах (PUT /messages, DELETE /messages, POST /answers и т.п.).
type SimpleResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func (s *messageService) EditMessage(ctx context.Context, messageID string, body EditMessageBody) (SimpleResult, error) {
	if strings.TrimSpace(messageID) == "" {
		return SimpleResult{}, errors.New("maxapi: messageID is required")
	}

	q := url.Values{}
	q.Set("message_id", messageID)

	var res SimpleResult
	if err := s.client.do(ctx, http.MethodPut, "/messages", q, body, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID string) (SimpleResult, error) {
	if strings.TrimSpace(messageID) == "" {
		return SimpleResult{}, errors.New("maxapi: messageID is required")
	}

	q := url.Values{}
	q.Set("message_id", messageID)

	var res SimpleResult
	if err := s.client.do(ctx, http.MethodDelete, "/messages", q, nil, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

// CallbackAnswerBody — тело запроса для POST /answers.
// См. https://dev.max.ru/docs-api/methods/POST/answers
type CallbackAnswerBody struct {
	// Message — новое/обновлённое сообщение.
	Message *NewMessageBody `json:"message,omitempty"`
	// Notification — одноразовое уведомление для пользователя.
	Notification *string `json:"notification,omitempty"`
}

func (s *messageService) AnswerCallback(ctx context.Context, callbackID string, body CallbackAnswerBody) (SimpleResult, error) {
	if strings.TrimSpace(callbackID) == "" {
		return SimpleResult{}, errors.New("maxapi: callbackID is required")
	}

	q := url.Values{}
	q.Set("callback_id", callbackID)

	var res SimpleResult
	if err := s.client.do(ctx, http.MethodPost, "/answers", q, body, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

// UpdatesParams — параметры запроса к GET /updates.
type UpdatesParams struct {
	// Limit — максимальноe количество обновлений (1–1000, по умолчанию 100).
	Limit int
	// Timeout — таймаут long polling в секундах (0–90, по умолчанию 30).
	Timeout int
	// Marker — указатель на следующую страницу данных (int64).
	// Если nil, будут получены все новые обновления.
	Marker *int64
	// Types — список типов обновлений (message_created, message_callback, ...).
	Types []string
}

// UpdatesResult — результат вызова GetUpdates.
type UpdatesResult struct {
	Updates []Update
	Marker  *int64
}

// внутренний тип для декодирования ответа /updates
type updatesResponse struct {
	Updates []Update `json:"updates"`
	Marker  *int64   `json:"marker"`
}

// subscriptionService — реализация SubscriptionService.
type subscriptionService struct {
	client *apiClient
}

func (s *subscriptionService) GetUpdates(ctx context.Context, params UpdatesParams) (UpdatesResult, error) {
	q := url.Values{}
	if params.Limit != 0 {
		q.Set("limit", intToString(params.Limit))
	}

	if params.Timeout != 0 {
		q.Set("timeout", intToString(params.Timeout))
	}

	if params.Marker != nil {
		q.Set("marker", int64ToString(*params.Marker))
	}

	if len(params.Types) > 0 {
		q.Set("types", strings.Join(params.Types, ","))
	}

	var resp updatesResponse
	if err := s.client.do(ctx, http.MethodGet, "/updates", q, nil, &resp); err != nil {
		return UpdatesResult{}, err
	}

	return UpdatesResult{
		Updates: resp.Updates,
		Marker:  resp.Marker,
	}, nil
}

// UploadService описывает операции по инициализации загрузки файлов (POST /uploads).
type UploadService interface {
	// InitUpload запрашивает URL для загрузки файла указанного типа.
	// Тип соответствует параметру type: "image", "video", "audio", "file".
	InitUpload(ctx context.Context, uploadType string) (*UploadInitResponse, error)

	// GetVideoInfo получает информацию о видео по videoToken (GET /videos/{videoToken}).
	GetVideoInfo(ctx context.Context, videoToken string) (*VideoInfo, error)
}

// uploadService — реализация UploadService.
type uploadService struct {
	client *apiClient
}

func (s *uploadService) InitUpload(ctx context.Context, uploadType string) (*UploadInitResponse, error) {
	if strings.TrimSpace(uploadType) == "" {
		return nil, errors.New("maxapi: uploadType is required")
	}
	q := url.Values{}
	q.Set("type", uploadType)

	var res UploadInitResponse
	if err := s.client.do(ctx, http.MethodPost, "/uploads", q, nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *uploadService) GetVideoInfo(ctx context.Context, videoToken string) (*VideoInfo, error) {
	if strings.TrimSpace(videoToken) == "" {
		return nil, errors.New("maxapi: videoToken is required")
	}

	path := "/videos/" + videoToken
	var info VideoInfo
	if err := s.client.do(ctx, http.MethodGet, path, url.Values{}, nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ListSubscriptionsResponse — внутренний тип для GET /subscriptions.
type ListSubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// CreateSubscriptionBody — тело запроса для POST /subscriptions.
type CreateSubscriptionBody struct {
	URL         string   `json:"url"`
	UpdateTypes []string `json:"update_types,omitempty"`
	Secret      string   `json:"secret,omitempty"`
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	var resp ListSubscriptionsResponse
	if err := s.client.do(ctx, http.MethodGet, "/subscriptions", url.Values{}, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Subscriptions, nil
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, body CreateSubscriptionBody) (SimpleResult, error) {
	if strings.TrimSpace(body.URL) == "" {
		return SimpleResult{}, errors.New("maxapi: subscription URL is required")
	}
	var res SimpleResult
	if err := s.client.do(ctx, http.MethodPost, "/subscriptions", url.Values{}, body, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, subscriptionURL string) (SimpleResult, error) {
	if strings.TrimSpace(subscriptionURL) == "" {
		return SimpleResult{}, errors.New("maxapi: subscription URL is required")
	}
	q := url.Values{}
	q.Set("url", subscriptionURL)

	var res SimpleResult
	if err := s.client.do(ctx, http.MethodDelete, "/subscriptions", q, nil, &res); err != nil {
		return SimpleResult{}, err
	}
	return res, nil
}

