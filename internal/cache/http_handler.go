package cache

import (
	"errors"
	"net/http"
)

// HTTPHandler exposes cache reads over HTTP.
func (c *Cache) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		val, ok := c.Get(key)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(val)
	})

	mux.HandleFunc("/owner_get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		val, err := c.ResolveAsOwner(key)
		if err != nil {
			if errors.Is(err, ErrNoOwnerLoader) {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(val)
	})

	return mux
}
