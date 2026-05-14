package makepay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestVerifyAndParseWebhook(t *testing.T) {
	body := []byte(`{"event":{"type":"status_changed"}}`)
	secret := "whsec_test"
	timestamp := time.Now().Unix()
	header := signWebhookForTest(body, secret, timestamp)

	if !VerifyWebhook(body, header, secret) {
		t.Fatal("expected signature to verify")
	}
	if VerifyWebhook(body, header, "wrong") {
		t.Fatal("expected wrong secret to fail")
	}

	event, err := ParseWebhook(body, header, secret)
	if err != nil {
		t.Fatal(err)
	}
	if event["event"] == nil {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestVerifyWebhookTolerance(t *testing.T) {
	body := []byte(`{"event":{"type":"status_changed"}}`)
	secret := "whsec_test"
	timestamp := time.Now().Add(-10 * time.Minute).Unix()
	header := signWebhookForTest(body, secret, timestamp)

	if VerifyWebhookWithTolerance(body, header, secret, 5*time.Minute) {
		t.Fatal("expected stale signature to fail")
	}
	if !VerifyWebhookWithTolerance(body, header, secret, 0) {
		t.Fatal("expected disabled tolerance to pass")
	}
}

func signWebhookForTest(body []byte, secret string, timestamp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(body)
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}
