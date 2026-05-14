package makepay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientResourceMethods(t *testing.T) {
	type call struct {
		name   string
		method string
		path   string
		run    func(*Client) (map[string]any, error)
	}

	calls := []call{
		{"create donation", http.MethodPost, "/api/partner/v1/makepay/donations", func(client *Client) (map[string]any, error) {
			return client.CreateDonationLink(context.Background(), DonationLinkPayload{"title": "Campaign"}, nil)
		}},
		{"list customers", http.MethodGet, "/api/partner/v1/makepay/customers", func(client *Client) (map[string]any, error) {
			return client.ListCustomers(context.Background())
		}},
		{"upsert customer", http.MethodPost, "/api/partner/v1/makepay/customers", func(client *Client) (map[string]any, error) {
			return client.UpsertCustomer(context.Background(), CustomerPayload{"email": "buyer@example.com"})
		}},
		{"customer portal", http.MethodPost, "/api/partner/v1/makepay/customers/cus_123/portal", func(client *Client) (map[string]any, error) {
			return client.CreateCustomerPortal(context.Background(), "cus_123", map[string]any{"returnUrl": "https://merchant.example"})
		}},
		{"create subscription", http.MethodPost, "/api/partner/v1/makepay/subscriptions", func(client *Client) (map[string]any, error) {
			return client.CreateSubscription(context.Background(), SubscriptionPayload{"amountUsd": "29"})
		}},
		{"list destination assets", http.MethodGet, "/api/partner/v1/makepay/destination-assets", func(client *Client) (map[string]any, error) {
			return client.ListDestinationAssets(context.Background())
		}},
		{"webhook requests", http.MethodGet, "/api/partner/v1/makepay/webhook-requests", func(client *Client) (map[string]any, error) {
			return client.ListWebhookRequests(context.Background(), map[string]any{"limit": 10})
		}},
		{"create pos terminal", http.MethodPost, "/api/partner/v1/makepay/pos-terminals", func(client *Client) (map[string]any, error) {
			return client.CreatePosTerminal(context.Background(), PosTerminalPayload{"name": "Counter"})
		}},
		{"get pos terminal", http.MethodGet, "/api/partner/v1/makepay/pos-terminals/terminal_123", func(client *Client) (map[string]any, error) {
			return client.GetPosTerminal(context.Background(), "terminal_123")
		}},
		{"create product", http.MethodPost, "/api/partner/v1/makepay/products", func(client *Client) (map[string]any, error) {
			return client.CreateProduct(context.Background(), ProductPayload{"name": "Guide"})
		}},
		{"product downloads", http.MethodGet, "/api/partner/v1/makepay/shop/products/prod_123/downloads", func(client *Client) (map[string]any, error) {
			return client.ListProductDownloads(context.Background(), "prod_123")
		}},
		{"update shop", http.MethodPatch, "/api/partner/v1/makepay/shop", func(client *Client) (map[string]any, error) {
			return client.UpdateShop(context.Background(), ShopPayload{"slug": "merchant"})
		}},
		{"shop builder", http.MethodPut, "/api/partner/v1/makepay/shop/builder", func(client *Client) (map[string]any, error) {
			return client.UpdateShopBuilder(context.Background(), map[string]any{"blocks": []any{}})
		}},
		{"shop domain", http.MethodPut, "/api/partner/v1/makepay/shop/domains", func(client *Client) (map[string]any, error) {
			return client.UpdateShopDomain(context.Background(), "shop.merchant.example")
		}},
		{"archive coupon", http.MethodDelete, "/api/partner/v1/makepay/shop/coupons/coupon_123", func(client *Client) (map[string]any, error) {
			return client.ArchiveShopCoupon(context.Background(), "coupon_123")
		}},
		{"branding domains", http.MethodPost, "/api/partner/v1/makepay/branding/domains/refresh", func(client *Client) (map[string]any, error) {
			return client.RefreshBrandingDomains(context.Background(), "")
		}},
		{"bookkeeping invoice", http.MethodPost, "/api/partner/v1/makepay/bookkeeping/invoices", func(client *Client) (map[string]any, error) {
			return client.CreateBookkeepingInvoice(context.Background(), BookkeepingInvoicePayload{"title": "Invoice"})
		}},
		{"bookkeeping invoice payment link", http.MethodPost, "/api/partner/v1/makepay/bookkeeping/invoices/inv_123/payment-link", func(client *Client) (map[string]any, error) {
			return client.CreateBookkeepingInvoicePaymentLink(context.Background(), "inv_123", map[string]any{"sendPaymentRequestEmail": true})
		}},
		{"bookkeeping expense from activity", http.MethodPost, "/api/partner/v1/makepay/bookkeeping/expenses/from-activity", func(client *Client) (map[string]any, error) {
			return client.CreateBookkeepingExpenseFromActivity(context.Background(), BookkeepingExpensePayload{"walletActivityEventKey": "event_123"})
		}},
		{"bookkeeping document ocr", http.MethodPost, "/api/partner/v1/makepay/bookkeeping/documents/doc_123/ocr", func(client *Client) (map[string]any, error) {
			return client.RunBookkeepingDocumentOCR(context.Background(), "doc_123")
		}},
		{"bookkeeping reconciliation", http.MethodPost, "/api/partner/v1/makepay/bookkeeping/reconciliation", func(client *Client) (map[string]any, error) {
			return client.CreateBookkeepingReconciliation(context.Background(), BookkeepingReconciliationPayload{"invoiceId": "inv_123"})
		}},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != tc.method {
					t.Fatalf("expected %s, got %s", tc.method, request.Method)
				}
				if request.URL.Path != tc.path {
					t.Fatalf("unexpected path: %s", request.URL.Path)
				}
				if request.URL.Query().Get("limit") == "" && strings.Contains(tc.name, "webhook requests") {
					t.Fatalf("query not forwarded: %s", request.URL.RawQuery)
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(`{"ok":true}`))
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

			response, err := tc.run(client)
			if err != nil {
				t.Fatal(err)
			}
			if response["ok"] != true {
				t.Fatalf("unexpected response: %#v", response)
			}
		})
	}
}

func TestCreateDonationLinkAddsDonationType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}

		payload, ok := body["payload"].(map[string]any)
		if !ok {
			t.Fatalf("missing payload: %#v", body)
		}
		if payload["type"] != "donation" {
			t.Fatalf("expected donation type, got %#v", payload["type"])
		}
		if body["sendPaymentRequestEmail"] != true {
			t.Fatalf("expected email flag, got %#v", body["sendPaymentRequestEmail"])
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(ClientOptions{BaseURL: server.URL, KeyID: "mk_test", KeySecret: "mksec_test"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateDonationLink(context.Background(), DonationLinkPayload{"title": "Campaign"}, &CreatePaymentLinkOptions{
		SendPaymentRequestEmail: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUploadBookkeepingDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		reader, err := request.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}

		fields := map[string]string{}
		fileSeen := false
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			value, _ := io.ReadAll(part)
			if part.FormName() == "file" {
				fileSeen = part.FileName() == "receipt.pdf" && string(value) == "pdf bytes"
				continue
			}
			fields[part.FormName()] = string(value)
		}

		if !fileSeen {
			t.Fatal("expected uploaded file")
		}
		if fields["documentType"] != "receipt" || fields["expenseId"] != "exp_123" {
			t.Fatalf("unexpected multipart fields: %#v", fields)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(ClientOptions{BaseURL: server.URL, KeyID: "mk_test", KeySecret: "mksec_test"})
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.UploadBookkeepingDocument(context.Background(), BookkeepingDocumentUpload{
		File:         strings.NewReader("pdf bytes"),
		FileName:     "receipt.pdf",
		DocumentType: "receipt",
		ExpenseID:    "exp_123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response["ok"] != true {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestUploadBookkeepingDocumentRequiresFile(t *testing.T) {
	client, err := NewClient(ClientOptions{KeyID: "mk_test", KeySecret: "mksec_test"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.UploadBookkeepingDocument(context.Background(), BookkeepingDocumentUpload{})
	if err == nil {
		t.Fatal("expected file validation error")
	}
}

func TestCreateAnonymousPaymentLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-MakeCrypto-Key-Id") != "" {
			t.Fatal("anonymous request should not send API key headers")
		}
		if request.URL.Path != "/api/partner/v1/makepay/payment-links" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["amount"] != "25" {
			t.Fatalf("unexpected body: %#v", body)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	response, err := CreateAnonymousPaymentLink(context.Background(), AnonymousPaymentLinkPayload{
		"amount": "25",
	}, PublicRequestOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if response["ok"] != true {
		t.Fatalf("unexpected response: %#v", response)
	}
}
