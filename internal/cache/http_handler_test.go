package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPHandlerReturnsValue(t *testing.T) {
	c := New(10)
	c.Set("user:1", []byte("value-1"), time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/cache?key=user:1", nil)
	rr := httptest.NewRecorder()

	c.HTTPHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "value-1" {
		t.Fatalf("expected value-1, got %s", rr.Body.String())
	}
}

func TestHTTPHandlerMissingKey(t *testing.T) {
	c := New(10)

	req := httptest.NewRequest(http.MethodGet, "/cache", nil)
	rr := httptest.NewRecorder()

	c.HTTPHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestHTTPHandlerNotFound(t *testing.T) {
	c := New(10)

	req := httptest.NewRequest(http.MethodGet, "/cache?key=missing", nil)
	rr := httptest.NewRecorder()

	c.HTTPHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
func TestHTTPHandlerOwnerGet(t *testing.T) {
	c := New(10)
	c.RegisterOwnerLoader(time.Minute, func(key string) ([]byte, error) {
		return []byte("owner-http-value"), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/owner_get?key=user:owner", nil)
	rr := httptest.NewRecorder()

	c.HTTPHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if rr.Body.String() != "owner-http-value" {
		t.Fatalf("expected owner-http-value, got %s", rr.Body.String())
	}
}
