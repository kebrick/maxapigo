package maxapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://platform-api.max.ru"
	defaultTimeout = 10 * time.Second
)

// Client описывает высокоуровневый интерфейс для работы с API MAX.
type Client interface {
	// Bots возвращает сервис работы с ботом.
	Bots() BotService
	// Messages возвращает сервис работы с сообщениями.
	Messages() MessageService
	// Chats возвращает сервис работы с групповыми чатами.
	Chats() ChatService
	// Subscriptions возвращает сервис работы с подписками и обновлениями.
	Subscriptions() SubscriptionService
	// Uploads возвращает сервис инициализации загрузок и информации о медиа.
	Uploads() UploadService
}

// Config задаёт параметры клиента.
type Config struct {
	// Token — токен бота, передаётся в заголовке Authorization: <token>.
	Token string

	// BaseURL — базовый URL API. По умолчанию https://platform-api.max.ru.
	BaseURL string

	// HTTPClient — кастомный http.Client; если nil — создаётся со стандартным таймаутом.
	HTTPClient *http.Client

	// Logger — интерфейс логгера. Если nil — логи отключены.
	Logger Logger

	// Debug включает расширенное логирование запросов/ответов.
	Debug bool
}

// apiClient — внутренняя реализация Client.
type apiClient struct {
	cfg        Config
	httpClient *http.Client
	baseURL    *url.URL
	logger     Logger

	bots          *botService
	messages      *messageService
	chats         *chatService
	subscriptions *subscriptionService
	uploads       *uploadService
}

// NewClient создаёт новый экземпляр клиента.
func NewClient(cfg Config) (Client, error) {
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, errors.New("maxapi: token is required")
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("maxapi: invalid base URL: %w", err)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	logger := cfg.Logger
	if logger == nil {
		logger = &noopLogger{}
	}

	c := &apiClient{
		cfg:        cfg,
		httpClient: httpClient,
		baseURL:    u,
		logger:     logger,
	}

	c.bots = &botService{client: c}
	c.messages = &messageService{client: c}
	c.chats = &chatService{client: c}
	c.subscriptions = &subscriptionService{client: c}
	c.uploads = &uploadService{client: c}

	return c, nil
}

func (c *apiClient) Bots() BotService {
	return c.bots
}

func (c *apiClient) Messages() MessageService {
	return c.messages
}

func (c *apiClient) Chats() ChatService {
	return c.chats
}

func (c *apiClient) Subscriptions() SubscriptionService {
	return c.subscriptions
}

func (c *apiClient) Uploads() UploadService {
	return c.uploads
}

// do выполняет HTTP-запрос к API MAX.
func (c *apiClient) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	u := *c.baseURL
	u.Path = strings.TrimRight(c.baseURL.Path, "/") + path
	u.RawQuery = query.Encode()

	var reqBodyReader *strings.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("maxapi: encode request: %w", err)
		}
		reqBodyReader = strings.NewReader(string(data))
	} else {
		reqBodyReader = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBodyReader)
	if err != nil {
		return fmt.Errorf("maxapi: create request: %w", err)
	}

	req.Header.Set("Authorization", c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	if c.cfg.Debug {
		c.logger.Debugf(ctx, "maxapi: %s %s", method, u.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("maxapi: do request: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	if c.cfg.Debug {
		c.logger.Debugf(ctx, "maxapi: %s %s -> %d (%s)", method, u.String(), resp.StatusCode, elapsed)
	}

	if resp.StatusCode >= 400 {
		apiErr, readErr := newAPIError(resp)
		if readErr != nil {
			return fmt.Errorf("maxapi: read error response: %w", readErr)
		}
		return apiErr
	}

	if out == nil {
		return nil
	}

	// Читаем тело ответа целиком, чтобы при отладке можно было вывести его.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("maxapi: read response body: %w", err)
	}

	if c.cfg.Debug {
		c.logger.Debugf(ctx, "maxapi: response body: %s", string(data))
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("maxapi: decode response: %w", err)
	}

	return nil
}

