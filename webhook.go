package makepay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is the default allowed timestamp drift for signed
// MakePay webhooks.
const DefaultWebhookTolerance = 5 * time.Minute

// VerifyWebhook verifies the x-makepay-signature header against the exact raw
// request body using DefaultWebhookTolerance.
func VerifyWebhook(rawBody []byte, signatureHeader string, secret string) bool {
	return VerifyWebhookWithTolerance(rawBody, signatureHeader, secret, DefaultWebhookTolerance)
}

// VerifyWebhookWithTolerance verifies the x-makepay-signature header against
// the exact raw request body. A non-positive tolerance disables the timestamp
// freshness check.
func VerifyWebhookWithTolerance(
	rawBody []byte,
	signatureHeader string,
	secret string,
	tolerance time.Duration,
) bool {
	if strings.TrimSpace(signatureHeader) == "" || secret == "" {
		return false
	}

	parts := parseSignatureHeader(signatureHeader)
	timestamp, err := strconv.ParseInt(parts["t"], 10, 64)
	if err != nil || timestamp <= 0 {
		return false
	}

	actual, err := hex.DecodeString(parts["v1"])
	if err != nil || len(actual) == 0 {
		return false
	}

	if tolerance > 0 {
		maxDrift := int64(tolerance / time.Second)
		now := time.Now().Unix()
		if absInt64(now-timestamp) > maxDrift {
			return false
		}
	}

	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write([]byte(strconv.FormatInt(timestamp, 10)))
	expectedMAC.Write([]byte("."))
	expectedMAC.Write(rawBody)
	expected := expectedMAC.Sum(nil)

	return hmac.Equal(expected, actual)
}

// ParseWebhook verifies and decodes a signed MakePay webhook JSON body using
// DefaultWebhookTolerance.
func ParseWebhook(rawBody []byte, signatureHeader string, secret string) (map[string]any, error) {
	return ParseWebhookWithTolerance(rawBody, signatureHeader, secret, DefaultWebhookTolerance)
}

// ParseWebhookWithTolerance verifies and decodes a signed MakePay webhook JSON
// body. A non-positive tolerance disables the timestamp freshness check.
func ParseWebhookWithTolerance(
	rawBody []byte,
	signatureHeader string,
	secret string,
	tolerance time.Duration,
) (map[string]any, error) {
	if !VerifyWebhookWithTolerance(rawBody, signatureHeader, secret, tolerance) {
		return nil, &Error{
			StatusCode: 401,
			Err:        errors.New("Invalid MakePay webhook signature."),
		}
	}

	var event map[string]any
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return nil, &Error{
			StatusCode: 400,
			Err:        errors.New("Invalid MakePay webhook JSON body."),
		}
	}

	return event, nil
}

func parseSignatureHeader(header string) map[string]string {
	parts := map[string]string{}
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && key != "" {
			parts[key] = value
		}
	}
	return parts
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
