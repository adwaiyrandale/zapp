package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrMissingKey  = errors.New("idempotency key is required")
	ErrKeyExpired  = errors.New("idempotency key has expired")
	ErrKeyConflict = errors.New("idempotency key already used with different request")
)

type Store interface {
	Get(ctx context.Context, key string) (*Response, error)
	Set(ctx context.Context, key string, response *Response, ttl time.Duration) error
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

type Idempotency struct {
	store Store
	ttl   time.Duration
}

func New(store Store, ttl time.Duration) *Idempotency {
	return &Idempotency{
		store: store,
		ttl:   ttl,
	}
}

func (i *Idempotency) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			key = r.URL.Query().Get("idempotency_key")
		}

		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		stored, err := i.store.Get(ctx, key)
		if err == nil && stored != nil {
			for k, v := range stored.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(stored.StatusCode)
			w.Write(stored.Body)
			return
		}

		rec := &responseRecorder{
			ResponseWriter: w,
			body:           []byte{},
			statusCode:     http.StatusOK,
			headers:        make(map[string]string),
		}

		next.ServeHTTP(rec, r)

		if rec.statusCode < 200 || rec.statusCode >= 300 {
			return
		}

		for k := range w.Header() {
			rec.headers[k] = w.Header().Get(k)
		}

		response := &Response{
			StatusCode: rec.statusCode,
			Body:       rec.body,
			Headers:    rec.headers,
		}

		i.store.Set(ctx, key, response, i.ttl)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	body       []byte
	statusCode int
	headers    map[string]string
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) Get(ctx context.Context, key string) (*Response, error) {
	data, err := s.client.Get(ctx, "idempotency:"+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *RedisStore) Set(ctx context.Context, key string, response *Response, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, "idempotency:"+key, data, ttl).Err()
}

func Middleware(store Store, ttl time.Duration) func(http.Handler) http.Handler {
	idempotency := New(store, ttl)
	return idempotency.Middleware
}

func ExtractKey(r *http.Request) string {
	if key := r.Header.Get("Idempotency-Key"); key != "" {
		return key
	}
	return r.URL.Query().Get("idempotency_key")
}

func RequireKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := ExtractKey(r)
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func KeyFromCtx(r *http.Request) string {
	return chi.URLParam(r, "idempotency_key")
}

func InvalidKeyError() error {
	return fmt.Errorf("%w: invalid key format", ErrMissingKey)
}
