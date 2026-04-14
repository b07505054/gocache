package peer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPGetterGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key != "user:1" {
			http.Error(w, "unexpected key", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("peer-value"))
	}))
	defer server.Close()

	getter := NewHTTPGetter(server.URL)

	val, err := getter.Get(context.Background(), "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val) != "peer-value" {
		t.Fatalf("expected peer-value, got %s", val)
	}
}

func TestHTTPGetterNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	getter := NewHTTPGetter(server.URL)

	_, err := getter.Get(context.Background(), "missing-key")
	if err == nil {
		t.Fatalf("expected error for non-200 response")
	}
}
func TestHTTPGetterOwnerGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/owner_get" {
			http.Error(w, "unexpected path", http.StatusBadRequest)
			return
		}
		key := r.URL.Query().Get("key")
		if key != "user:owner" {
			http.Error(w, "unexpected key", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte("owner-value"))
	}))
	defer server.Close()

	getter := NewHTTPGetter(server.URL)

	val, err := getter.OwnerGet(context.Background(), "user:owner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val) != "owner-value" {
		t.Fatalf("expected owner-value, got %s", val)
	}
}
