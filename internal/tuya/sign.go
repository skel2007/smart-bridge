package tuya

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

const emptyContentSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func signRequest(secret, clientID, accessToken, timestamp, nonce, method, canonicalURL string, body []byte) string {
	stringToSign := buildStringToSign(method, canonicalURL, body)
	payload := clientID + accessToken + timestamp + nonce + stringToSign

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))

	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func buildStringToSign(method, canonicalURL string, body []byte) string {
	return strings.ToUpper(method) + "\n" + contentSHA256(body) + "\n\n" + canonicalURL
}

func contentSHA256(body []byte) string {
	if len(body) == 0 {
		return emptyContentSHA256
	}

	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func canonicalURL(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}

	return path + "?" + query.Encode()
}
