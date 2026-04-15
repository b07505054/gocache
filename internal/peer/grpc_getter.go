package peer

import (
	"context"
	"fmt"
	"sync"
	"time"

	cachepb "github.com/b07505054/gocache/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCGetter struct {
	addr   string
	mu     sync.Mutex
	conn   *grpc.ClientConn
	client cachepb.CacheServiceClient
}

func NewGRPCGetter(addr string) (*GRPCGetter, error) {
	return &GRPCGetter{
		addr: addr,
	}, nil
}

func (g *GRPCGetter) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.conn != nil {
		err := g.conn.Close()
		g.conn = nil
		g.client = nil
		return err
	}
	return nil
}

func (g *GRPCGetter) ensureClient(ctx context.Context) (cachepb.CacheServiceClient, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.client != nil {
		return g.client, nil
	}

	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		g.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial grpc peer %s: %w", g.addr, err)
	}

	g.conn = conn
	g.client = cachepb.NewCacheServiceClient(conn)
	return g.client, nil
}

func (g *GRPCGetter) Get(ctx context.Context, key string) ([]byte, error) {
	client, err := g.ensureClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetCached(ctx, &cachepb.CacheRequest{Key: key})
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}

func (g *GRPCGetter) OwnerGet(ctx context.Context, key string) ([]byte, error) {
	client, err := g.ensureClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.OwnerGet(ctx, &cachepb.CacheRequest{Key: key})
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}
