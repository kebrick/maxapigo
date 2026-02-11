package maxapi

import (
	"fmt"
	"io"
	"net/http"
)

// APIError представляет собой ошибку, возвращённую сервером MAX.
type APIError struct {
	StatusCode int
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("maxapi: API error: status=%d body=%s", e.StatusCode, string(e.Body))
}

func newAPIError(resp *http.Response) (*APIError, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

