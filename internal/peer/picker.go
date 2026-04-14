package peer

import "context"

// Getter defines the behavior for fetching cache data from a peer node.
type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// OwnerGetter defines the behavior for asking the owner node
// to resolve a key with load-on-miss semantics.
type OwnerGetter interface {
	Getter
	OwnerGet(ctx context.Context, key string) ([]byte, error)
}

// Picker defines the behavior for selecting the appropriate peer for a key.
type Picker interface {
	PickPeer(key string) (Getter, bool)
}
