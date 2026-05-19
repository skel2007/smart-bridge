package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
)

func (api *API) do(ctx context.Context, method, path string, route string, query url.Values, body []byte, accessToken string, out any) error {
	canonical := canonicalURL(path, query)
	requestURL := api.credentials.Endpoint + canonical

	req, err := retryablehttp.NewRequestWithContext(withRoute(ctx, route), method, requestURL, body)
	if err != nil {
		return fmt.Errorf("create tuya request: %w", err)
	}
	if accessToken != "" {
		req.Header.Set("access_token", accessToken)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := api.signRequest(req.Request); err != nil {
		return err
	}

	resp, err := api.transport.Do(req)
	if err != nil {
		return fmt.Errorf("call tuya api: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read tuya response: %w", err)
	}

	return decodeResponse(resp.StatusCode, data, out)
}

func (api *API) signRequest(req *http.Request) error {
	body, err := requestBody(req)
	if err != nil {
		return fmt.Errorf("read tuya request body for signing: %w", err)
	}

	nonce, err := api.nonce()
	if err != nil {
		return fmt.Errorf("create tuya request nonce: %w", err)
	}

	timestamp := strconv.FormatInt(api.now().UnixMilli(), 10)
	accessToken := req.Header.Get("access_token")
	canonical := req.URL.RequestURI()
	req.Header.Set("client_id", api.credentials.ClientID)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("t", timestamp)
	req.Header.Set("nonce", nonce)
	req.Header.Set("sign", signRequest(api.credentials.ClientSecret, api.credentials.ClientID, accessToken, timestamp, nonce, req.Method, canonical, body))

	return nil
}

func requestBody(req *http.Request) ([]byte, error) {
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = body.Close()
		}()

		return io.ReadAll(body)
	}

	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	return body, nil
}

func decodeResponse(statusCode int, data []byte, out any) error {
	var envelope responseEnvelope
	decodeErr := json.Unmarshal(data, &envelope)
	if statusCode < 200 || statusCode >= 300 {
		if decodeErr == nil {
			return &APIError{StatusCode: statusCode, Code: normalizeCode(envelope.Code), Message: envelope.Msg}
		}

		return &APIError{StatusCode: statusCode}
	}
	if decodeErr != nil {
		return fmt.Errorf("decode tuya response: %w", decodeErr)
	}
	if !envelope.Success {
		return &APIError{StatusCode: statusCode, Code: normalizeCode(envelope.Code), Message: envelope.Msg}
	}
	if out == nil || len(envelope.Result) == 0 || bytes.Equal(envelope.Result, []byte("null")) {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, out); err != nil {
		return fmt.Errorf("decode tuya result: %w", err)
	}

	return nil
}

type responseEnvelope struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Code    json.RawMessage `json:"code"`
	Msg     string          `json:"msg"`
}

func normalizeCode(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	return string(raw)
}
