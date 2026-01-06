package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"GoBlue/proto/hello"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ctx = context.Background()

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/redis", s.RedisHandler)
	mux.HandleFunc("/proto3", s.ProtoHandler)
	mux.HandleFunc("/proto3/debug", s.ProtoDebugHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-H/Users/abhigyan/Downloads/proto3 eaders", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "ALL GOOD"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Falied to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write resp: %v", err)
	}
}

func (s *Server) RedisHandler(w http.ResponseWriter, r *http.Request) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	val, err := rdb.Incr(ctx, "pageviews").Result()
	if err != nil {
		panic(err)
	}

	resp := map[string]string{
		"route":     "redis",
		"pageviews": fmt.Sprintf("%d", val),
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) ProtoHandler(w http.ResponseWriter, r *http.Request) {
	resp := &hello.HelloResponse{
		Message:   "Hello from protobuf",
		Pageviews: 2,
		Time:      timestamppb.New(time.Now()),
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal protobuf", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) ProtoDebugHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("proto/hello/proto3")
	if err != nil {
		http.Error(w, "failed to read proto file", http.StatusInternalServerError)
		log.Printf("read error: %v", err)
		return
	}

	msg := &hello.HelloResponse{}
	if err := proto.Unmarshal(data, msg); err != nil {
		http.Error(w, "failed to unmarshal proto", http.StatusInternalServerError)
		log.Printf("unmarshal error: %v", err)
		return
	}

	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		http.Error(w, "failed to marshal protojson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
