package cloud

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	tokenURI          = "/v1.0/token?grant_type=1"
	refreshTokenURI   = "/v1.0/token/refresh-token"
	devicesURI        = "/v2.0/cloud/thing/device?page_size=20"
	deviceCommandsURI = "/v1.0/devices/device-id/commands"
)

type testResponse struct {
	statusCode int
	headers    map[string]string
	body       string
}

type requestRecorder struct {
	t       *testing.T
	mu      sync.Mutex
	routes  map[testRouteKey][]testResponse
	records []*http.Request
}

type testRoute struct {
	key       testRouteKey
	responses []testResponse
}

type testRouteKey struct {
	method string
	uri    string
}

func get(uri string, responses ...testResponse) testRoute {
	return testRoute{key: testRouteKey{method: http.MethodGet, uri: uri}, responses: responses}
}

func post(uri string, responses ...testResponse) testRoute {
	return testRoute{key: testRouteKey{method: http.MethodPost, uri: uri}, responses: responses}
}

func newTestAPI(t *testing.T, routes ...testRoute) (*API, *requestRecorder) {
	t.Helper()

	recorder := &requestRecorder{
		t:      t,
		routes: make(map[testRouteKey][]testResponse),
	}
	for _, route := range routes {
		recorder.routes[route.key] = append(recorder.routes[route.key], route.responses...)
	}

	server := httptest.NewServer(recorder)
	t.Cleanup(server.Close)

	api := NewAPI(Credentials{
		Endpoint:     server.URL,
		ClientID:     "client",
		ClientSecret: "super-secret",
	})
	api.now = func() time.Time {
		return time.UnixMilli(1700000000000)
	}
	api.nonce = func() (string, error) {
		return "nonce", nil
	}

	return api, recorder
}

func (recorder *requestRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !recorder.recordRequest(req) {
		http.Error(w, "read request body", http.StatusInternalServerError)
		return
	}

	response, ok := recorder.nextResponse(req)
	if !ok {
		http.Error(w, "unexpected request", http.StatusInternalServerError)
		return
	}

	response.writeTo(w)
}

func (recorder *requestRecorder) recordRequest(req *http.Request) bool {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		recorder.t.Errorf("read request body: %v", err)
		return false
	}

	request := req.Clone(req.Context())
	request.Header = req.Header.Clone()
	request.Body = io.NopCloser(bytes.NewReader(body))

	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.records = append(recorder.records, request)

	return true
}

func (recorder *requestRecorder) nextResponse(req *http.Request) (testResponse, bool) {
	route := testRouteKey{method: req.Method, uri: req.URL.RequestURI()}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	responses := recorder.routes[route]
	if len(responses) == 0 {
		recorder.t.Errorf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
		return testResponse{}, false
	}

	response := responses[0]
	recorder.routes[route] = responses[1:]

	return response, true
}

func (recorder *requestRecorder) request(index int) *http.Request {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return recorder.records[index]
}

func (recorder *requestRecorder) requestCount() int {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return len(recorder.records)
}

func tuyaResult(result string) testResponse {
	return jsonResponse(http.StatusOK, `{"success":true,"result":`+result+`}`)
}

func tuyaToken(accessToken, refreshToken string, expireTime int64) testResponse {
	return tuyaResult(`{"access_token":"` + accessToken + `","refresh_token":"` + refreshToken + `","expire_time":` + strconv.FormatInt(expireTime, 10) + `}`)
}

func tuyaError(statusCode int, code, message string) testResponse {
	return jsonResponse(statusCode, `{"success":false,"code":"`+code+`","msg":"`+message+`"}`)
}

func retryableTuyaError(statusCode int, code, message string) testResponse {
	resp := tuyaError(statusCode, code, message)
	resp.headers = map[string]string{
		"Retry-After": "0",
	}

	return resp
}

func jsonResponse(statusCode int, body string) testResponse {
	return testResponse{
		statusCode: statusCode,
		body:       body,
	}
}

func (response testResponse) writeTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	for name, value := range response.headers {
		w.Header().Set(name, value)
	}
	w.WriteHeader(response.statusCode)
	_, _ = w.Write([]byte(response.body))
}
