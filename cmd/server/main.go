package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"net"

	cachepb "github.com/b07505054/gocache/api/proto"
	"github.com/b07505054/gocache/internal/cache"
	"github.com/b07505054/gocache/internal/peer"

	transportgrpc "github.com/b07505054/gocache/internal/transportgrpc"
	"google.golang.org/grpc"
)

func main() {

	addr := flag.String("addr", ":8080", "server listen address")
	self := flag.String("self", "nodeA", "current node name")
	grpcAddr := flag.String("grpc_addr", ":9090", "grpc listen address")
	transport := flag.String("transport", "http", "peer transport: http or grpc")
	peersFlag := flag.String("peers", "", "comma-separated peer mappings in node=url form")
	flag.Parse()

	c := cache.New(3)
	c.RegisterOwnerLoader(30*time.Second, func(k string) ([]byte, error) {
		log.Println("OWNER LOAD for key:", k, "on", *self)
		time.Sleep(200 * time.Millisecond)
		return []byte("db-value-for-" + k + "-from-" + *self), nil
	})

	if *peersFlag != "" {
		picker := peer.NewMapPicker(*self, 50)

		if *transport == "grpc" {
			if err := picker.SetGRPCPeers(parsePeers(*peersFlag)); err != nil {
				log.Fatal(err)
			}
		} else {
			picker.SetHTTPPeers(parsePeers(*peersFlag))
		}

		c.RegisterPeers(picker)
	}

	go func() {
		lis, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			log.Fatal(err)
		}

		grpcServer := grpc.NewServer()
		cachepb.RegisterCacheServiceServer(grpcServer, transportgrpc.NewServer(c))

		log.Println("grpc server running on", *grpcAddr, "as", *self)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	// Internal peer-to-peer endpoint
	http.Handle("/cache", c.HTTPHandler())

	// Public endpoint
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		val, err := c.GetOrLoad(key, 30*time.Second, func(k string) ([]byte, error) {
			log.Println("FALLBACK LOCAL LOAD for key:", k, "on", *self)
			time.Sleep(200 * time.Millisecond)
			return []byte("db-value-for-" + k + "-from-" + *self), nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"node":  *self,
			"key":   key,
			"value": string(val),
		})
	})

	// Optional manual preload endpoint for demo
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")

		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		c.Set(key, []byte(value), 30*time.Second)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"node":   *self,
			"status": "ok",
			"key":    key,
			"value":  value,
		})
	})

	log.Println("server running on", *addr, "as", *self)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func parsePeers(s string) map[string]string {
	peers := make(map[string]string)
	if s == "" {
		return peers
	}

	parts := strings.Split(s, ",")
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		node := strings.TrimSpace(pair[0])
		url := strings.TrimSpace(pair[1])
		if node != "" && url != "" {
			peers[node] = url
		}
	}
	return peers
}
