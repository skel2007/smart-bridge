package tuya

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoSignsRequestHeaders(t *testing.T) {
	client, api := newTestClient(t, map[string]testResponse{
		tokenURI: tuyaResult(`{"access_token":"access-token"}`),
	})

	query := url.Values{}
	query.Set("grant_type", "1")

	var result tuyaTokenResult
	err := client.do(context.Background(), http.MethodGet, tokenPath, query, nil, "", &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", result.AccessToken)
	require.Len(t, api.requests, 1)

	req := api.requests[0]
	require.Equal(t, "client", req.Header.Get("client_id"))
	require.Equal(t, "HMAC-SHA256", req.Header.Get("sign_method"))
	require.Equal(t, "1700000000000", req.Header.Get("t"))
	require.Equal(t, "nonce", req.Header.Get("nonce"))
	require.Empty(t, req.Header.Get("access_token"))
	require.NotEmpty(t, req.Header.Get("sign"))
}

func TestDoSetsAccessTokenHeader(t *testing.T) {
	client, api := newTestClient(t, map[string]testResponse{
		devicesURI: tuyaResult(`[]`),
	})

	query := url.Values{}
	query.Set("page_size", "20")

	var result []tuyaDevice
	err := client.do(context.Background(), http.MethodGet, projectDevices, query, nil, "access-token", &result)

	require.NoError(t, err)
	require.Equal(t, "access-token", api.requests[0].Header.Get("access_token"))
}

func TestDecodeResponseResult(t *testing.T) {
	var result tuyaTokenResult

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
