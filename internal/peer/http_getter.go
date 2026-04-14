package peer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPGetter fetches cache values from a remote HTTP peer.
type HTTPGetter struct {
	baseURL string
	client  *http.Client
}

// NewHTTPGetter creates a new HTTPGetter.
func NewHTTPGetter(baseURL string) *HTTPGetter {
	return &HTTPGetter{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Get fetches a value from the remote peer cache only.
func (h *HTTPGetter) Get(ctx context.Context, key string) ([]byte, error) {
	return h.doGet(ctx, "/cache", key)
}

// OwnerGet asks the owner node to resolve the key with load-on-miss semantics.
func (h *HTTPGetter) OwnerGet(ctx context.Context, key string) ([]byte, error) {
	return h.doGet(ctx, "/owner_get", key)
}

func (h *HTTPGetter) doGet(ctx context.Context, path string, key string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s", h.baseURL, path, url.QueryEscape(key))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote peer returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
