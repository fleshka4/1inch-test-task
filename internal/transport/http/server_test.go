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

	"github.com/fleshka4/1inch-test-task/internal/apperrors"
	"github.com/fleshka4/1inch-test-task/internal/config"
	"github.com/fleshka4/1inch-test-task/internal/service/mock"
)

func TestNilServerNilConfig(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server, err := NewServer(mockService, nil)
	require.Error(t, err)
	require.Nil(t, server)
}

func TestPingHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server, err := NewServer(mockService, &config.Config{})
	require.NoError(t, err)

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

	const (
		pool      = "0x1234567890123456789012345678901234567890"
		src       = "0x1234567890123456789012345678901234567891"
		dst       = "0x1234567890123456789012345678901234567892"
		srcAmount = "1000000000000000000"
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		method         string
		queryParams    map[string]string
		mockSetup      func(*mock.MockService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "success",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			mockSetup: func(ms *mock.MockService) {
				ms.EXPECT().Estimate(gomock.Any(), gomock.Any()).
					Return(big.NewInt(1000000000000000000), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   srcAmount,
		},
		{
			name:   "validation error - missing params",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool": "0x123",
				"src":  "0x456",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "validation error - bad address",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       "invalid",
				"src":        src,
				"dst":        dst,
				"src_amount": "1000",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "validation error - bad src_amount",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": "-1000",
			},
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "service error - invalid argument",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			mockSetup: func(ms *mock.MockService) {
				ms.EXPECT().Estimate(gomock.Any(), gomock.Any()).
					Return(nil, apperrors.ErrInvalidArgument)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "service error - insufficient liquidity",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			mockSetup: func(ms *mock.MockService) {
				ms.EXPECT().Estimate(gomock.Any(), gomock.Any()).
					Return(nil, apperrors.ErrInsufficientLiquidity)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:   "service error - unknown error",
			method: http.MethodGet,
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			mockSetup: func(ms *mock.MockService) {
				ms.EXPECT().Estimate(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unknown error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
		{
			name:           "wrong http method",
			method:         http.MethodPost,
			queryParams:    nil,
			mockSetup:      nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := mock.NewMockService(ctrl)
			server, err := NewServer(mockService, &config.Config{})
			require.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			req := httptest.NewRequest(tt.method, "/estimate", nil)
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Body.Close: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, string(body))
			}

			if tt.expectedStatus == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				require.Equal(t, "text/plain; charset=utf-8", contentType)
			}
		})
	}
}

func TestLogMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockService(ctrl)
	server, err := NewServer(mockService, &config.Config{})
	require.NoError(t, err)

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
	server, err := NewServer(mockService, &config.Config{
		ReadHeaderTimeout: 5 * time.Second,
		GraceTimeout:      5 * time.Second,
	})
	require.NoError(t, err)

	const addr = "localhost:0"

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe(addr)
	}()

	time.Sleep(100 * time.Millisecond)

	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
