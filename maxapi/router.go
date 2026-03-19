package maxapi

import (
	"context"
	"strings"
)

// MessageHandler обрабатывает входящее сообщение.
type MessageHandler func(ctx context.Context, msg *Message) error

// UpdateMiddleware выполняется до роутинга апдейта.
// Если handled == true, Router не будет вызывать другие middleware и хендлеры.
type UpdateMiddleware func(ctx context.Context, upd Update) (handled bool, err error)

// Router предоставляет удобное API для регистрации хендлеров
// по типу апдейта, командам (/start и т.п.) и смайликам/подстрокам.
//
// Использование:
//
//	router := maxapi.NewRouter()
//	router.HandleCommand("/start", func(ctx context.Context, msg *maxapi.Message) error { ... })
//	router.HandleEmoji("😀", func(ctx context.Context, msg *maxapi.Message) error { ... })
//	poller := maxapi.NewLongPoller(client, router.HandleUpdate, maxapi.LongPollerConfig{})
type Router struct {
	typeHandlers map[string]UpdateHandler

	commandHandlers map[string]MessageHandler
	emojiHandlers   map[string]MessageHandler

	defaultMessageHandler MessageHandler

	middlewares []UpdateMiddleware

	// mentionPrefix — строка вида "@username", которая будет
	// отрезана от начала текста сообщения перед разбором команд.
	// Полезно для обработки сообщений из каналов/групп, где
	// сначала идёт тег бота, а затем команда или текст.
	mentionPrefix string
}

// NewRouter создаёт новый роутер.
func NewRouter() *Router {
	return &Router{
		typeHandlers:         make(map[string]UpdateHandler),
		commandHandlers:      make(map[string]MessageHandler),
		emojiHandlers:        make(map[string]MessageHandler),
		defaultMessageHandler: nil,
		middlewares:           nil,
		mentionPrefix:         "",
	}
}

// NewRouterForClient создаёт Router и, при возможности, автоматически
// настраивает обработку упоминаний бота в группах/каналах.
//
// Внутри выполняется client.Bots().Me(ctx); если username получен,
// Router будет резать префикс "@username" в начале текста сообщений.
func NewRouterForClient(ctx context.Context, client Client) *Router {
	r := NewRouter()
	if client == nil {
		return r
	}

	me, err := client.Bots().Me(ctx)
	if err != nil {
		return r
	}
	if me.Username != "" {
		r.SetBotUsername(me.Username)
	}

	return r
}

// Use добавляет middleware, которое выполняется перед всеми хендлерами.
// Middleware может, например, логировать апдейты или делать проверку доступа.
// Если middleware возвращает handled=true, дальнейшая обработка апдейта прекращается.
func (r *Router) Use(mw UpdateMiddleware) {
	if mw == nil {
		return
	}
	r.middlewares = append(r.middlewares, mw)
}

// SetBotUsername задаёт username бота (без "@").
// Тогда Router будет корректно обрабатывать сообщения вида:
// "@my_bot /start" или "@my_bot просто текст" — сначала убирая
// префикс "@my_bot", а затем применяя обычные правила команд/текста.
func (r *Router) SetBotUsername(username string) {
	username = strings.TrimSpace(username)
	if username == "" {
		r.mentionPrefix = ""
		return
	}
	if strings.HasPrefix(username, "@") {
		r.mentionPrefix = username
	} else {
		r.mentionPrefix = "@" + username
	}
}

// HandleUpdateType регистрирует хендлер для конкретного типа апдейта
// (например, maxapi.UpdateTypeMessageCreated, maxapi.UpdateTypeMessageCallback, maxapi.UpdateTypeBotStarted).
func (r *Router) HandleUpdateType(updateType string, h UpdateHandler) {
	if updateType == "" || h == nil {
		return
	}
	r.typeHandlers[updateType] = h
}

// HandleCommand регистрирует хендлер для текстовой команды,
// например "/start", "/help" и т.п.
//
// Команда парсится как первое слово в сообщении, начинающееся с "/".
func (r *Router) HandleCommand(command string, h MessageHandler) {
	if command == "" || h == nil {
		return
	}
	r.commandHandlers[command] = h
}

// HandleEmoji регистрирует хендлер, который срабатывает,
// если текст сообщения содержит указанный смайлик (или любую подстроку).
func (r *Router) HandleEmoji(emoji string, h MessageHandler) {
	if emoji == "" || h == nil {
		return
	}
	r.emojiHandlers[emoji] = h
}

// HandleDefaultMessage задаёт обработчик по умолчанию для всех сообщений,
// которые не попали ни под один другой хендлер.
func (r *Router) HandleDefaultMessage(h MessageHandler) {
	r.defaultMessageHandler = h
}

// HandleUpdate реализует UpdateHandler и маршрутизирует апдейты
// на зарегистрированные хендлеры.
func (r *Router) HandleUpdate(ctx context.Context, upd Update) error {
	// 0. Middleware.
	for _, mw := range r.middlewares {
		handled, err := mw(ctx, upd)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	// Сначала проверяем хендлер по типу апдейта.
	if h := r.typeHandlers[upd.UpdateType]; h != nil {
		return h(ctx, upd)
	}

	// Далее работаем с сообщениями.
	if upd.Message == nil {
		return nil
	}
	msg := upd.Message
	text := strings.TrimSpace(msg.Text())
	if text == "" {
		return nil
	}

	// Если указано имя бота, режем префикс "@bot" в начале сообщения.
	if r.mentionPrefix != "" && strings.HasPrefix(text, r.mentionPrefix) {
		text = strings.TrimSpace(text[len(r.mentionPrefix):])
		if text == "" {
			// Сообщение содержало только упоминание бота.
			return nil
		}
	}

	// 1. Команды (/start, /help и т.п.).
	if strings.HasPrefix(text, "/") {
		cmd := text
		if i := strings.IndexAny(text, " \t\r\n"); i > 0 {
			cmd = text[:i]
		}
		if h := r.commandHandlers[cmd]; h != nil {
			return h(ctx, msg)
		}
	}

	// 2. Хендлеры по смайликам/подстрокам.
	for emoji, h := range r.emojiHandlers {
		if strings.Contains(text, emoji) {
			return h(ctx, msg)
		}
	}

	// 3. Хендлер по умолчанию.
	if r.defaultMessageHandler != nil {
		return r.defaultMessageHandler(ctx, msg)
	}

	return nil
}

