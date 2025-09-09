package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/fleshka4/1inch-test-task/internal/config"
	"github.com/fleshka4/1inch-test-task/internal/service"
)

// Server represents the HTTP transport layer.
type Server struct {
	est service.Service
	mux *http.ServeMux

	graceTimeout      time.Duration
	readHeaderTimeout time.Duration
	requestTimeout    time.Duration
}

// NewServer creates a new HTTP server with registered routes.
func NewServer(est service.Service, cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	s := &Server{
		est: est,
		mux: http.NewServeMux(),

		graceTimeout:      cfg.GraceTimeout,
		readHeaderTimeout: cfg.ReadHeaderTimeout,
		requestTimeout:    cfg.RequestTimeout,
	}

	s.mux.HandleFunc("/estimate", s.handleEstimate)
	s.mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("pong")); err != nil {
			log.Printf("ping write error: %v", err)
		}
	})

	return s, nil
}

// ListenAndServe starts the HTTP server and enables graceful shutdown.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.logMiddleware(s.mux),
		ReadHeaderTimeout: s.readHeaderTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("http server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Block until a signal is received.
	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.graceTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "srv.Shutdown")
	}
	log.Println("server stopped gracefully")
	return nil
}

// logMiddleware logs each HTTP request and the time taken to process it.
func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.String(), time.Since(start))
	})
}
