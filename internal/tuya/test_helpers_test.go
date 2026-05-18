package tuya

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
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

type testResponse struct {
	statusCode int
	body       string
}

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

func newTestAPI(t *testing.T, routes map[testRoute]testResponse) (*api, *testAPI) {
	t.Helper()

	testAPI := &testAPI{t: t, routes: routes}
	testAPI.server = httptest.NewServer(http.HandlerFunc(testAPI.handle))
	t.Cleanup(testAPI.server.Close)

	api := newAPI(Credentials{
		Endpoint:     testAPI.server.URL,
		ClientID:     "client",
		ClientSecret: "super-secret",
	})
	api.httpClient = testAPI.server.Client()
	api.now = func() time.Time {
		return time.UnixMilli(1700000000000)
	}
	api.nonce = func() (string, error) {
		return "nonce", nil
	}

	return api, testAPI
}

type testAPI struct {
	t        *testing.T
	server   *httptest.Server
	routes   map[testRoute]testResponse
	mu       sync.Mutex
	requests []*http.Request
	bodies   []string
}

func (api *testAPI) handle(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		api.t.Errorf("read request body: %v", err)
		http.Error(w, "read request body", http.StatusInternalServerError)
		return
	}

	api.mu.Lock()
	api.requests = append(api.requests, req.Clone(req.Context()))
	api.bodies = append(api.bodies, string(body))
	api.mu.Unlock()

	requestRoute := route(req.Method, req.URL.RequestURI())
	resp, ok := api.routes[requestRoute]
	if !ok {
		api.t.Errorf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
		http.Error(w, "unexpected request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.statusCode)
	_, _ = w.Write([]byte(resp.body))
}

func (api *testAPI) requestURIs() []string {
	api.mu.Lock()
	defer api.mu.Unlock()

	uris := make([]string, 0, len(api.requests))
	for _, req := range api.requests {
		uris = append(uris, req.URL.RequestURI())
	}

	return uris
}

func tuyaResult(result string) testResponse {
	return jsonResponse(http.StatusOK, `{"success":true,"result":`+result+`}`)
}

func tuyaError(statusCode int, code, message string) testResponse {
	return jsonResponse(statusCode, `{"success":false,"code":"`+code+`","msg":"`+message+`"}`)
}

func jsonResponse(statusCode int, body string) testResponse {
	return testResponse{
		statusCode: statusCode,
		body:       body,
	}
}
