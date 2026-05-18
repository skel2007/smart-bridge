package cloud

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildStringToSign(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	expected := strings.Join([]string{
		"POST",
		contentSHA256(body),
		"",
		"/v1.0/devices",
	}, "\n")

	require.Equal(t, expected, buildStringToSign("post", "/v1.0/devices", body))
}

func TestContentSHA256(t *testing.T) {
	require.Equal(t, emptyContentSHA256, contentSHA256(nil))
	require.Equal(t, "93a23971a914e5eacbf0a8d25154cda309c3c1c72fbb9914d47c60f3cb681588", contentSHA256([]byte(`{"hello":"world"}`)))
}

func TestCanonicalURL(t *testing.T) {
	query := url.Values{}
	query.Set("page_size", "20")
	query.Set("last_id", "abc")

	require.Equal(t, "/v2.0/cloud/thing/device?last_id=abc&page_size=20", canonicalURL("/v2.0/cloud/thing/device", query))
}
