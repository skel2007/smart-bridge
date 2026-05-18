package cloud

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoSignsRequestHeaders(t *testing.T) {
	api, recorder := newTestAPI(t,
		get(tokenURI, tuyaResult(`{"access_token":"access-token"}`)),
	)

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tokenResult
	err := api.do(context.Background(), http.MethodGet, tokenPath, query, nil, "", &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", result.AccessToken)
	require.Equal(t, 1, recorder.requestCount())

	req := recorder.request(0)
	require.Equal(t, "client", req.Header.Get("client_id"))
	require.Equal(t, "HMAC-SHA256", req.Header.Get("sign_method"))
	require.Equal(t, "1700000000000", req.Header.Get("t"))
	require.Equal(t, "nonce", req.Header.Get("nonce"))
	require.Empty(t, req.Header.Get("access_token"))
	require.NotEmpty(t, req.Header.Get("sign"))
}

func TestDoRetriesRetryableHTTPStatusWithFreshSignature(t *testing.T) {
	api, recorder := newTestAPI(t,
		get(tokenURI,
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
			tuyaResult(`{"access_token":"access-token"}`),
		),
	)
	nonce := 0
	api.nonce = func() (string, error) {
		nonce++
		return "nonce-" + strconv.Itoa(nonce), nil
	}

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tokenResult
	err := api.do(context.Background(), http.MethodGet, tokenPath, query, nil, "", &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", result.AccessToken)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, tokenURI, recorder.request(1).URL.RequestURI())
	require.Equal(t, "nonce-1", recorder.request(0).Header.Get("nonce"))
	require.Equal(t, "nonce-2", recorder.request(1).Header.Get("nonce"))
	require.NotEqual(t, recorder.request(0).Header.Get("sign"), recorder.request(1).Header.Get("sign"))
}

func TestDoStopsRetryingAfterMaxAttempts(t *testing.T) {
	api, recorder := newTestAPI(t,
		get(tokenURI,
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
			retryableTuyaError(http.StatusServiceUnavailable, "SYSTEM_ERROR", "try later"),
		),
	)

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tokenResult
	err := api.do(context.Background(), http.MethodGet, tokenPath, query, nil, "", &result)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
	require.Equal(t, tokenURI, recorder.request(0).URL.RequestURI())
	require.Equal(t, tokenURI, recorder.request(1).URL.RequestURI())
	require.Equal(t, tokenURI, recorder.request(2).URL.RequestURI())
	require.Equal(t, tokenURI, recorder.request(3).URL.RequestURI())
	require.Equal(t, tokenURI, recorder.request(4).URL.RequestURI())
}

func TestDoSetsAccessTokenHeader(t *testing.T) {
	api, recorder := newTestAPI(t,
		get(devicesURI, tuyaResult(`[]`)),
	)

	query := url.Values{}
	query.Set("page_size", "20")

	var result []Device
	err := api.do(context.Background(), http.MethodGet, projectDevices, query, nil, "access-token", &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", recorder.request(0).Header.Get("access_token"))
}

func TestSignRequestPreservesBodyWithoutGetBody(t *testing.T) {
	api, _ := newTestAPI(t)
	req, err := http.NewRequest(http.MethodPost, "https://tuya.test/v1.0/devices/device-id/commands", io.NopCloser(strings.NewReader(`{"commands":[]}`)))
	require.NoError(t, err)

	err = api.signRequest(req)

	require.NoError(t, err)
	require.NotEmpty(t, req.Header.Get("sign"))

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{"commands":[]}`, string(body))
}

func TestDecodeResponseResult(t *testing.T) {
	var result tokenResult

	err := decodeResponse(http.StatusOK, []byte(`{"success":true,"result":{"access_token":"access-token"}}`), &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", result.AccessToken)
}

func TestDecodeResponseReturnsTuyaErrorResponse(t *testing.T) {
	err := decodeResponse(http.StatusOK, []byte(`{"success":false,"code":1106,"msg":"permission denied"}`), nil)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusOK, apiErr.StatusCode)
	require.Equal(t, "1106", apiErr.Code)
	require.Equal(t, "permission denied", apiErr.Message)
}

func TestDecodeResponseReturnsHTTPError(t *testing.T) {
	err := decodeResponse(http.StatusServiceUnavailable, []byte(`{"success":false,"code":"SYSTEM_ERROR","msg":"try later"}`), nil)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
	require.Equal(t, "SYSTEM_ERROR", apiErr.Code)
	require.Equal(t, "try later", apiErr.Message)
}
