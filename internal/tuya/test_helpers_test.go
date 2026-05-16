package tuya

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	tokenURI                = "/v1.0/token?grant_type=1"
	devicesURI              = "/v2.0/cloud/thing/device?page_size=20"
	deviceSpecificationsURI = "/v1.0/devices/device-id/specifications"
	deviceStatusURI         = "/v1.0/devices/device-id/status"
	deviceCommandsURI       = "/v1.0/devices/device-id/commands"
)

type testResponse func() *http.Response

type testRoute struct {
	method string
	uri    string
}

func route(method string, uri string) testRoute {
	return testRoute{method: method, uri: uri}
}

func getRoute(uri string) testRoute {
	return route(http.MethodGet, uri)
}

func postRoute(uri string) testRoute {
	return route(http.MethodPost, uri)
}

func newTestClient(t *testing.T, routes map[testRoute]testResponse) (*Client, *testTuyaAPI) {
	t.Helper()

	api := &testTuyaAPI{t: t, routes: routes}
	client := NewClient(
		Credentials{
			Endpoint:     "https://example.com",
			ClientID:     "client",
			ClientSecret: "super-secret",
		},
		WithHTTPClient(&http.Client{Transport: api}),
		WithNowFunc(func() time.Time {
			return time.UnixMilli(1700000000000)
		}),
		WithNonceFunc(func() (string, error) {
			return "nonce", nil
		}),
	)

	return client, api
}

type testTuyaAPI struct {
	t        *testing.T
	routes   map[testRoute]testResponse
	requests []*http.Request
	bodies   []string
}

func (api *testTuyaAPI) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		api.t.Fatalf("read request body: %v", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	api.requests = append(api.requests, req)
	api.bodies = append(api.bodies, string(body))

	requestRoute := route(req.Method, req.URL.RequestURI())
	resp, ok := api.routes[requestRoute]
	if !ok {
		api.t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
	}

	return resp(), nil
}

func (api *testTuyaAPI) requestURIs() []string {
	uris := make([]string, 0, len(api.requests))
	for _, req := range api.requests {
		uris = append(uris, req.URL.RequestURI())
	}

	return uris
}

func tuyaResult(result string) testResponse {
	return func() *http.Response {
		return jsonResponse(http.StatusOK, `{"success":true,"result":`+result+`}`)
	}
}

func tuyaError(statusCode int, code, message string) testResponse {
	return func() *http.Response {
		return jsonResponse(statusCode, `{"success":false,"code":"`+code+`","msg":"`+message+`"}`)
	}
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
