package transportgrpc

import (
	"context"

	cachepb "github.com/b07505054/gocache/api/proto"
	"github.com/b07505054/gocache/internal/cache"
	"github.com/b07505054/gocache/internal/peer"
)

type Server struct {
	cachepb.UnimplementedCacheServiceServer
	cache *cache.Cache
}

func NewServer(c *cache.Cache) *Server {
	return &Server{cache: c}
}

func (s *Server) GetCached(ctx context.Context, req *cachepb.CacheRequest) (*cachepb.CacheResponse, error) {
	val, ok := s.cache.Get(req.GetKey())
	if !ok {
		return nil, peer.ErrNotFound
	}

	return &cachepb.CacheResponse{
		Value: val,
	}, nil
}

func (s *Server) OwnerGet(ctx context.Context, req *cachepb.CacheRequest) (*cachepb.CacheResponse, error) {
	val, err := s.cache.ResolveAsOwner(req.GetKey())
	if err != nil {
		return nil, err
	}

	return &cachepb.CacheResponse{
		Value: val,
	}, nil
}
