package http

import (
	"bytes"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fleshka4/1inch-test-task/internal/config"
	"github.com/fleshka4/1inch-test-task/internal/service"
	"github.com/fleshka4/1inch-test-task/internal/service/mock"
)

func TestPingHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server := NewServer(mockService, config.Config{})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	resp := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Body.Close: %v", err)
		}
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "pong", string(body))
}

func TestEstimateHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server := NewServer(mockService, config.Config{})

	t.Run("success", func(t *testing.T) {
		expectedAmount := big.NewInt(1000000000000000000)
		mockService.EXPECT().
			Estimate(gomock.Any(), gomock.Any()).
			Return(expectedAmount, nil)

		req := httptest.NewRequest("GET", "/estimate?pool=0x1234567890123456789012345678901234567890&"+
			"src=0x1234567890123456789012345678901234567891&dst=0x1234567890123456789012345678901234567892"+
			"&src_amount=1000000000000000000", nil)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			require.Equal(t, "text/plain; charset=utf-8", contentType)
		}

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedAmount.String(), string(body))
	})

	t.Run("validation error - missing params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/estimate?pool=0x123&src=0x456", nil)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error - bad address", func(t *testing.T) {
		req := httptest.NewRequest(
			"GET",
			"/estimate?"+
				"pool=invalid&"+
				"src=0x1234567890123456789012345678901234567891&"+
				"dst=0x1234567890123456789012345678901234567892&"+
				"src_amount=1000",
			nil,
		)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error - bad src_amount", func(t *testing.T) {
		req := httptest.NewRequest(
			"GET",
			"/estimate?"+
				"pool=0x1234567890123456789012345678901234567890&"+
				"src=0x1234567890123456789012345678901234567891&"+
				"dst=0x1234567890123456789012345678901234567892&"+
				"src_amount=-1000",
			nil,
		)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	testServiceError := func(t *testing.T, serviceError error, expectedStatusCode int) {
		mockService.EXPECT().
			Estimate(gomock.Any(), gomock.Any()).
			Return(nil, serviceError)

		req := httptest.NewRequest(
			"GET",
			"/estimate?"+
				"pool=0x1234567890123456789012345678901234567890&"+
				"src=0x1234567890123456789012345678901234567891&"+
				"dst=0x1234567890123456789012345678901234567892&"+
				"src_amount=1000000000000000000",
			nil,
		)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, expectedStatusCode, resp.StatusCode)
	}

	t.Run("service error - invalid argument", func(t *testing.T) {
		testServiceError(t, service.ErrInvalidArgument, http.StatusBadRequest)
	})

	t.Run("service error - insufficient liquidity", func(t *testing.T) {
		testServiceError(t, service.ErrInsufficientLiquidity, http.StatusBadRequest)
	})

	t.Run("service error - pair read failed", func(t *testing.T) {
		testServiceError(t, service.ErrPairRead, http.StatusBadGateway)
	})

	t.Run("service error - unknown error", func(t *testing.T) {
		testServiceError(t, errors.New("unknown error"), http.StatusInternalServerError)
	})

	t.Run("wrong http method", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("POST", "/estimate", nil)
		w := httptest.NewRecorder()

		server.mux.ServeHTTP(w, req)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Body.Close: %v", err)
			}
		}(resp.Body)

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func TestLogMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server := NewServer(mockService, config.Config{})

	var logOutput bytes.Buffer

	originalLogger := log.Writer()
	log.SetOutput(&logOutput)
	defer log.SetOutput(originalLogger)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler := server.logMiddleware(server.mux)
	handler.ServeHTTP(w, req)

	logContent := logOutput.String()
	require.Contains(t, logContent, "GET /ping")
}

func TestServer_ListenAndServe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server := NewServer(mockService, config.Config{
		ReadHeaderTimeout: 5 * time.Second,
		GraceTimeout:      5 * time.Second,
	})

	const addr = "localhost:0"

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe(addr)
	}()

	time.Sleep(100 * time.Millisecond)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
