// @title Zapp Payment Gateway API
// @version 1.0
// @description Payment Gateway with Payments, Settlements, and Ledger
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/adwaiyrandale/zapp
// @contact.email support@zapp.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
	port   string
}

func NewServer() *Server {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-API-Key, Idempotency-Key")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	// Swagger
	// r.Get("/swagger/*", swagger.Handler())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Proxy to payment service
	r.Route("/api/v1/payments", func(r chi.Router) {
		r.Handle("/*", proxyTo("http://localhost:8082"))
	})

	// Proxy to settlement service
	r.Route("/api/v1/settlements", func(r chi.Router) {
		r.Handle("/*", proxyTo("http://localhost:8083"))
	})

	// Proxy to ledger service
	r.Route("/api/v1/ledger", func(r chi.Router) {
		r.Handle("/*", proxyTo("http://localhost:8081"))
	})

	return &Server{
		router: r,
		port:   getEnv("PORT", "8080"),
	}
}

func proxyTo(target string) http.Handler {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	return proxy
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) Run() error {
	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Println("Shutting down server...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("API Gateway starting on port %s", s.port)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", s.port)
	return srv.ListenAndServe()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	server := NewServer()

	if err := server.Run(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
