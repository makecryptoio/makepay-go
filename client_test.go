package makepay

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCreatePaymentLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", request.Method)
		}
		if request.URL.Path != "/api/partner/v1/makepay/payment-links" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if got := request.Header.Get("X-MakeCrypto-Key-Id"); got != "mk_test" {
			t.Fatalf("unexpected key id header: %s", got)
		}
		if got := request.Header.Get("X-MakeCrypto-Key-Secret"); got != "mksec_test" {
			t.Fatalf("unexpected key secret header: %s", got)
		}
		if got := request.Header.Get("User-Agent"); got != "MakePayGo/"+Version {
			t.Fatalf("unexpected user agent: %s", got)
		}

		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["status"] != "active" {
			t.Fatalf("unexpected status: %#v", body["status"])
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"ok":true,"paymentLink":{"uid":"pay_123","publicUrl":"https://makepay.io/payment/pay_123"}}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(ClientOptions{
		BaseURL:   server.URL,
		KeyID:     "mk_test",
		KeySecret: "mksec_test",
	})
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.CreatePaymentLink(context.Background(), PaymentLinkPayload{
		"amount":   "12.50",
		"currency": "USDT",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if response["ok"] != true {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestClientQueryAndError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("limit") != "20" {
			t.Fatalf("query not forwarded: %s", request.URL.RawQuery)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusPaymentRequired)
		_, _ = writer.Write([]byte(`{"error":"payment required"}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(ClientOptions{
		BaseURL:   server.URL,
		KeyID:     "mk_test",
		KeySecret: "mksec_test",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.ListPaymentLinks(context.Background(), map[string]any{"limit": 20})
	if err == nil {
		t.Fatal("expected an API error")
	}

	var makePayError *Error
	if !errors.As(err, &makePayError) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if makePayError.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("unexpected status: %d", makePayError.StatusCode)
	}
	if makePayError.ResponseBody["error"] != "payment required" {
		t.Fatalf("unexpected response body: %#v", makePayError.ResponseBody)
	}
}

func TestClientRequiresKeys(t *testing.T) {
	_, err := NewClient(ClientOptions{})
	if err == nil {
		t.Fatal("expected configuration error")
	}

	var makePayError *Error
	if !errors.As(err, &makePayError) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if makePayError.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", makePayError.StatusCode)
	}
}
