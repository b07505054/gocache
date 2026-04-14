package peer

import (
	"context"
)

type mockGetter struct {
	value []byte
	err   error
}

func (m *mockGetter) Get(ctx context.Context, key string) ([]byte, error) {
	return m.value, m.err
}
