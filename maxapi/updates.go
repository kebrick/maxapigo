package maxapi

import (
	"context"
	"time"
)

// UpdateHandler — функция-обработчик обновлений.
type UpdateHandler func(ctx context.Context, upd Update) error

// LongPoller инкапсулирует логику long polling (GET /updates в цикле).
type LongPoller struct {
	client  Client
	handler UpdateHandler

	// Marker — указатель на следующую страницу данных.
	marker *int64

	// Задержка между попытками при ошибке.
	errorDelay time.Duration
}

// LongPollerConfig задаёт настройки long polling.
type LongPollerConfig struct {
	// Marker — начальный маркер, если нужно продолжить с определённого места.
	Marker *int64
	// ErrorDelay — задержка между запросами при ошибке.
	ErrorDelay time.Duration
}

// NewLongPoller создаёт новый LongPoller.
func NewLongPoller(client Client, handler UpdateHandler, cfg LongPollerConfig) *LongPoller {
	delay := cfg.ErrorDelay
	if delay == 0 {
		delay = 3 * time.Second
	}
	return &LongPoller{
		client:     client,
		handler:    handler,
		marker:     cfg.Marker,
		errorDelay: delay,
	}
}

// Run запускает бесконечный цикл получения обновлений.
func (p *LongPoller) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// timeout по умолчанию в API — 30 с; limit — 100 (см. GET /updates).
		res, err := p.client.Subscriptions().GetUpdates(ctx, UpdatesParams{
			Marker:  p.marker,
			Timeout: 30,
			Limit:   100,
		})
		if err != nil {
			// При ошибке просто ждём и повторяем.
			time.Sleep(p.errorDelay)
			continue
		}

		for _, upd := range res.Updates {
			if err := p.handler(ctx, upd); err != nil {
				// Ошибку обработки здесь намеренно игнорируем,
				// чтобы не останавливать общий цикл.
			}
		}

		if res.Marker != nil {
			p.marker = res.Marker
		}
	}
}

