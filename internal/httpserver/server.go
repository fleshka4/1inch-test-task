package httpserver

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/fleshka4/inch-test-task/internal/config"
	"github.com/fleshka4/inch-test-task/internal/handlers"
	"github.com/fleshka4/inch-test-task/internal/uniswapv2"
)

// Server initializes and runs the HTTP API.
type Server struct {
	client uniswapv2.PairReader
	mux    *http.ServeMux
	http   *http.Server

	gracefulTimeout time.Duration
}

// New constructs a new HTTP server.
func New(client uniswapv2.PairReader, cfg config.Config) (*Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/estimate", handlers.NewEstimateHandler(client, cfg.RequestTimeout))
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("pong")); err != nil {
			log.Printf("ping write error: %v", err)
		}
	})

	srv := &http.Server{
		Handler:           logMiddleware(mux),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	return &Server{
		client: client,
		mux:    mux,
		http:   srv,

		gracefulTimeout: cfg.ShutdownGrace,
	}, nil
}

// ListenAndServe starts the HTTP server and performs graceful shutdown on SIGINT/SIGTERM.
func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	// Runs server with goroutine.
	go func() {
		log.Printf("listening on %s", addr)
		if err := s.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http.Serve: %v", err)
		}
	}()

	// Waits for stop signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		log.Printf("http.Shutdown: %v", err)
		return errors.Wrap(err, "s.http.Shutdown")
	}

	log.Println("server stopped")
	return nil
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.String(), time.Since(start))
	})
}
