package validate

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pool      = "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	src       = "0x6B175474E89094C44Da98b954EedeAC495271d0F"
	dst       = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	srcAmount = "1000000000000000000"
)

func TestEstimateRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		queryParams    map[string]string
		method         string
		expectedStatus int
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name: "valid request",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: 0,
			wantErr:        assert.NoError,
		},
		{
			name: "wrong http method",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			wantErr:        assert.Error,
		},
		{
			name: "missing pool parameter",
			queryParams: map[string]string{
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "missing src parameter",
			queryParams: map[string]string{
				"pool":       pool,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "missing dst parameter",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "missing src_amount parameter",
			queryParams: map[string]string{
				"pool": pool,
				"src":  src,
				"dst":  dst,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "invalid pool address format",
			queryParams: map[string]string{
				"pool":       "invalid_address",
				"src":        src,
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "invalid src address format",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        "invalid_address",
				"dst":        dst,
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "invalid dst address format",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        "invalid_address",
				"src_amount": srcAmount,
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "invalid src_amount format",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": "not_a_number",
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "zero src_amount",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": "0",
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "negative src_amount",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": "-100",
			},
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			wantErr:        assert.Error,
		},
		{
			name: "very large src_amount",
			queryParams: map[string]string{
				"pool":       pool,
				"src":        src,
				"dst":        dst,
				"src_amount": "99999999999999999999999999999999999999999999999999",
			},
			method:         http.MethodGet,
			expectedStatus: 0,
			wantErr:        assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/estimate", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			result, status, err := EstimateRequestValidate(req)

			tt.wantErr(t, err)
			require.Equal(t, tt.expectedStatus, status)

			if result != nil {
				require.Equal(t, common.HexToAddress(tt.queryParams["pool"]), result.Pool)
				require.Equal(t, common.HexToAddress(tt.queryParams["src"]), result.Src)
				require.Equal(t, common.HexToAddress(tt.queryParams["dst"]), result.Dst)

				expectedAmount, ok := new(big.Int).SetString(tt.queryParams["src_amount"], 10)
				require.True(t, ok)
				require.Equal(t, expectedAmount, result.SrcAmount)
				require.True(t, result.SrcAmount.Sign() > 0)
			}
		})
	}
}

func TestEstimateRequestValidate_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty query parameters", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/estimate", nil)
		result, status, err := EstimateRequestValidate(req)

		require.Error(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Nil(t, result)
	})

	t.Run("different http methods", func(t *testing.T) {
		t.Parallel()

		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/estimate", nil)
				q := req.URL.Query()
				q.Add("pool", pool)
				q.Add("src", src)
				q.Add("dst", dst)
				q.Add("src_amount", "100")
				req.URL.RawQuery = q.Encode()

				result, status, err := EstimateRequestValidate(req)

				require.Error(t, err)
				require.Equal(t, http.StatusMethodNotAllowed, status)
				require.Nil(t, result)
			})
		}
	})
}
