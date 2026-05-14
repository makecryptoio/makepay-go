# MakePay Go SDK

Official Go SDK for MakePay server-side integrations. Use it to create crypto
payment links, donation pages, anonymous payment links, subscriptions, POS
terminals, products, Simple Shop storefronts, bookkeeping records, customer
portals, branded domains, checkout URLs, and signed webhook handlers.

## Install

```bash
go get github.com/makecryptoio/makepay-go
```

The package is published through the public Go module index:

```text
https://pkg.go.dev/github.com/makecryptoio/makepay-go
```

The module targets Go 1.22 or newer and uses only the Go standard library.

## Configure

Create a MakePay API key in MakeCrypto and keep the secret on your server only.

```go
package main

import (
	"context"
	"log"
	"os"

	makepay "github.com/makecryptoio/makepay-go"
)

func main() {
	client, err := makepay.NewClient(makepay.ClientOptions{
		KeyID:     os.Getenv("MAKEPAY_KEY_ID"),
		KeySecret: os.Getenv("MAKEPAY_KEY_SECRET"),
		// Optional: override only when MakePay gives you a custom checkout origin.
		CheckoutBaseURL: "https://makepay.io",
	})
	if err != nil {
		log.Fatal(err)
	}

	_ = client
	_ = context.Background()
}
```

The client sends `X-MakeCrypto-Key-Id` and `X-MakeCrypto-Key-Secret` headers to
the MakePay partner API.

## Payment Links

```go
response, err := client.CreatePaymentLink(context.Background(), makepay.PaymentLinkPayload{
	"title":          "Order #1042",
	"description":    "Checkout for order #1042",
	"amount":         "129.99",
	"currency":       "USDT",
	"orderId":        "order_1042",
	"customerEmail":  "buyer@example.com",
	"returnUrl":      "https://merchant.example/orders/1042",
	"successUrl":     "https://merchant.example/orders/1042/success",
	"failureUrl":     "https://merchant.example/orders/1042/pay",
	"expirationTime": "12h",
}, nil)
if err != nil {
	return err
}

log.Printf("created MakePay link: %#v", response["paymentLink"])
```

Read, update, and email existing links:

```go
links, err := client.ListPaymentLinks(ctx, map[string]any{"limit": 50})
detail, err := client.GetPaymentLink(ctx, "PAYMENT_LINK_UID")
updated, err := client.UpdatePaymentLink(ctx, "PAYMENT_LINK_UID", map[string]any{
	"status": "paused",
})
sent, err := client.SendPaymentRequestEmail(ctx, "PAYMENT_LINK_UID", "buyer@example.com")

_, _, _, _ = links, detail, updated, sent
```

## Donations

Donation pages are flexible-amount payment links with a public donation slug.

```go
donation, err := client.CreateDonationLink(ctx, makepay.DonationLinkPayload{
	"title":            "Spring campaign",
	"description":      "Support the 2026 spring fundraiser.",
	"defaultAmountUsd": "25",
	"minimumAmountUsd": "5",
	"donationSlug":     "spring-campaign",
}, nil)
if err != nil {
	return err
}

links, err := client.ListDonationLinks(ctx)
detail, err := client.GetDonationLink(ctx, "DONATION_UID")
updated, err := client.UpdateDonationLink(ctx, "DONATION_UID", map[string]any{
	"status": "paused",
})

_, _, _, _ = donation, links, detail, updated
```

## Anonymous Payment Links

Anonymous links do not use a MakePay API key. They require an explicit
settlement route because MakePay cannot read merchant wallet settings.

```go
response, err := makepay.CreateAnonymousPaymentLink(ctx, makepay.AnonymousPaymentLinkPayload{
	"amount": "25",
	"settlement": map[string]any{
		"currency": "USDT",
		"priorities": []map[string]any{
			{
				"chain":   "ETH",
				"address": "0xYourSettlementWallet",
				"asset":   "ETH.USDT-0xdAC17F958D2ee523a2206206994597C13D831ec7",
			},
		},
	},
	"title":      "Invoice #1042",
	"webhookUrl": "https://merchant.example/webhooks/makepay",
}, makepay.PublicRequestOptions{})
```

## Checkout URLs And Embeds

Use hosted checkout for redirects, or the embed helpers when your frontend
keeps the shopper on the merchant page.

```go
hostedURL, err := client.HostedCheckoutURL("PAYMENT_LINK_UID")
embeddedURL, err := client.EmbeddedCheckoutURL(
	"PAYMENT_LINK_UID",
	"https://merchant.example",
)
donationURL, err := client.HostedDonationURL("spring-campaign")
embeddedDonationURL, err := client.EmbeddedDonationURL(
	"spring-campaign",
	"https://merchant.example",
)

buttonHTML, err := client.EmbedButtonHTML("PAYMENT_LINK_UID", makepay.EmbedSnippetOptions{
	ButtonLabel: "Pay with crypto",
})
iframeHTML, err := client.IframeHTML("PAYMENT_LINK_UID", makepay.EmbedSnippetOptions{
	IframeTitle: "Secure MakePay checkout",
})

_, _, _, _, _, _ = hostedURL, embeddedURL, donationURL, embeddedDonationURL, buttonHTML, iframeHTML
```

## Customers And Subscriptions

```go
customer, err := client.UpsertCustomer(ctx, makepay.CustomerPayload{
	"email":    "buyer@example.com",
	"name":     "Buyer Example",
	"clientId": "crm_123",
})

portal, err := client.CreateCustomerPortal(ctx, "CUSTOMER_ID", map[string]any{
	"returnUrl": "https://merchant.example/account",
})

subscription, err := client.CreateSubscription(ctx, makepay.SubscriptionPayload{
	"amountUsd":            "29",
	"customerEmail":        "buyer@example.com",
	"label":                "Monthly plan",
	"billingIntervalUnit":  "month",
	"billingIntervalCount": 1,
})

_, _, _ = customer, portal, subscription
```

## POS Terminals

```go
terminal, err := client.CreatePosTerminal(ctx, makepay.PosTerminalPayload{
	"name":                "Front counter",
	"pin":                 "1234",
	"allowedAssets":       []string{"ETH.USDT-0xdAC17F958D2ee523a2206206994597C13D831ec7"},
	"emailCollectionMode": "optional_after_deposit",
	"catalogEnabled":      true,
})

terminals, err := client.ListPosTerminals(ctx)
detail, err := client.GetPosTerminal(ctx, "TERMINAL_UID")

_, _, _ = terminal, terminals, detail
```

## Products And Simple Shop

```go
product, err := client.CreateProduct(ctx, makepay.ProductPayload{
	"name":         "Digital guide",
	"productType":  "digital",
	"basePriceUsd": "19",
	"shopSlug":     "digital-guide",
	"images": []map[string]any{
		{"url": "https://merchant.example/guide.png", "alt": "Guide cover"},
	},
	"variants": []map[string]any{
		{"name": "PDF", "priceUsd": "19"},
	},
})

downloads, err := client.CreateProductDownload(ctx, "PRODUCT_UID", map[string]any{
	"fileName":    "guide.pdf",
	"contentType": "application/pdf",
	"url":         "https://merchant.example/downloads/guide.pdf",
})

shop, err := client.UpdateShop(ctx, makepay.ShopPayload{
	"slug":            "merchant-shop",
	"displayCurrency": "USD",
	"checkoutMode":    "hosted",
	"branding":        map[string]any{"accentColor": "#14b8a6"},
})

domain, err := client.UpdateShopDomain(ctx, "shop.merchant.example")
refreshed, err := client.RefreshShopDomain(ctx, nil)
coupon, err := client.CreateShopCoupon(ctx, map[string]any{
	"code":         "SPRING10",
	"discountType": "percent",
	"value":        "10",
})
orders, err := client.ListShopOrders(ctx, map[string]any{"status": "paid", "limit": 25})

_, _, _, _, _, _, _ = product, downloads, shop, domain, refreshed, coupon, orders
```

## Invoices And Bookkeeping

Bookkeeping APIs manage merchant invoices, expenses, supporting documents, OCR,
and reconciliation links.

```go
created, err := client.CreateBookkeepingInvoice(ctx, makepay.BookkeepingInvoicePayload{
	"title":       "Invoice #1042",
	"currency":    "USD",
	"issueDate":   "2026-05-15",
	"dueDate":     "2026-05-30",
	"counterparty": map[string]any{
		"name":     "Buyer Example",
		"email":    "buyer@example.com",
		"clientId": "crm_123",
	},
	"lineItems": []map[string]any{
		{
			"description": "Implementation services",
			"quantity":    "1",
			"unitAmount":  "500",
			"taxAmount":   "0",
		},
	},
	"metadata": map[string]any{"orderId": "order_1042"},
})

_, err = client.CreateBookkeepingInvoicePaymentLink(ctx, "INVOICE_UID", map[string]any{
	"sendPaymentRequestEmail": true,
})

_, _ = created, err
```

Expenses can be created manually or from wallet activity, then linked back to
payments, transfers, invoices, or uploaded receipts.

```go
expense, err := client.CreateBookkeepingExpense(ctx, makepay.BookkeepingExpensePayload{
	"title":       "Hosting",
	"amount":      "49",
	"currency":    "USD",
	"incurredOn":  "2026-05-15",
	"category":    "Infrastructure",
	"counterparty": map[string]any{"name": "Vendor Example", "type": "vendor"},
})

activityExpense, err := client.CreateBookkeepingExpenseFromActivity(ctx, makepay.BookkeepingExpensePayload{
	"walletActivityEventKey": "CHAIN_EVENT_KEY",
	"category":               "Settlement",
})

reconciliation, err := client.CreateBookkeepingReconciliation(ctx, makepay.BookkeepingReconciliationPayload{
	"invoiceId":        "INVOICE_UID",
	"paymentSessionId": "PAYMENT_SESSION_ID",
	"linkType":         "payment",
})

_, _, _ = expense, activityExpense, reconciliation
```

Document uploads use multipart form data through an `io.Reader`.

```go
file, err := os.Open("receipt.pdf")
if err != nil {
	return err
}
defer file.Close()

uploaded, err := client.UploadBookkeepingDocument(ctx, makepay.BookkeepingDocumentUpload{
	File:         file,
	FileName:     "receipt.pdf",
	DocumentType: "receipt",
	ExpenseID:    "EXPENSE_UID",
})

documents, err := client.ListBookkeepingDocuments(ctx)
download, err := client.GetBookkeepingDocumentDownloadURL(ctx, "DOCUMENT_UID")
ocr, err := client.RunBookkeepingDocumentOCR(ctx, "DOCUMENT_UID")
summary, err := client.GetBookkeepingSummary(ctx)

_, _, _, _, _ = uploaded, documents, download, ocr, summary
```

## Branding And Domains

```go
branding, err := client.UpdateBranding(ctx, makepay.BrandingPayload{
	"brandName":               "Merchant",
	"supportEmail":            "support@merchant.example",
	"brandingBrandColor":      "#111827",
	"brandingAccentColor":     "#14b8a6",
	"paymentLinkTheme":        "system",
	"paymentLinkDomain":       "pay.merchant.example",
	"emailSendingDomain":      "mail.merchant.example",
})

refreshed, err := client.RefreshBrandingDomains(ctx, "all")

_, _ = branding, refreshed
```

## Settings And Operational APIs

```go
settings, err := client.GetSettings(ctx)
updated, err := client.UpdateSettings(ctx, map[string]any{
	"callbackUrl": "https://merchant.example/webhooks/makepay",
})

assets, err := client.ListDestinationAssets(ctx)
webhooks, err := client.ListWebhookRequests(ctx, map[string]any{"limit": 25})

_, _, _, _ = settings, updated, assets, webhooks
```

## Verify Webhooks

Read the exact raw body before parsing JSON.

```go
func handleMakePayWebhook(writer http.ResponseWriter, request *http.Request) {
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "invalid body", http.StatusBadRequest)
		return
	}

	event, err := makepay.ParseWebhook(
		rawBody,
		request.Header.Get("x-makepay-signature"),
		os.Getenv("MAKEPAY_WEBHOOK_SECRET"),
	)
	if err != nil {
		http.Error(writer, "invalid signature", http.StatusUnauthorized)
		return
	}

	if event["event"] != nil {
		// Update your local order status.
	}

	writer.WriteHeader(http.StatusOK)
}
```

Use `VerifyWebhook` when you only need a boolean result.

## Method Coverage

| Area            | SDK methods |
| --------------- | ----------- |
| Payment links   | `CreatePaymentLink`, `ListPaymentLinks`, `GetPaymentLink`, `UpdatePaymentLink`, `SendPaymentRequestEmail` |
| Donations       | `CreateDonationLink`, `ListDonationLinks`, `GetDonationLink`, `UpdateDonationLink` |
| Anonymous links | `CreateAnonymousPaymentLink` |
| Checkout        | hosted, embedded, modal, button, iframe, and donation URL helpers |
| Customers       | `ListCustomers`, `UpsertCustomer`, `CreateCustomerPortal` |
| Subscriptions   | `ListSubscriptions`, `CreateSubscription` |
| POS terminals   | `ListPosTerminals`, `CreatePosTerminal`, `GetPosTerminal`, `UpdatePosTerminal` |
| Products        | `ListProducts`, `CreateProduct`, `GetProduct`, `UpdateProduct`, `ListProductDownloads`, `CreateProductDownload` |
| Simple Shop     | `GetShop`, `UpdateShop`, `GetShopBuilder`, `UpdateShopBuilder`, domain, coupon, and order methods |
| Bookkeeping     | summary, invoice, expense, document upload/OCR, and reconciliation methods |
| Branding        | `GetBranding`, `UpdateBranding`, `RefreshBrandingDomains` |
| Operations      | `GetSettings`, `UpdateSettings`, `ListDestinationAssets`, `ListWebhookRequests` |
| Webhooks        | `VerifyWebhook`, `ParseWebhook` |

## Payload And Response Models

The SDK keeps request payloads open-ended with `map[string]any` aliases because
several MakePay surfaces are configurable and continue to gain fields. Send
camelCase keys for new integrations. Some API routes may accept snake_case for
compatibility, but camelCase is the stable SDK convention.

Model conventions:

- Use strings for decimal money values when precision matters, for example
  `"129.99"` instead of `129.99`.
- Dates are ISO strings. Date-only fields, such as invoice `issueDate`, should
  use `YYYY-MM-DD`.
- IDs are usually public `uid` values. Bookkeeping detail endpoints accept an
  internal UUID or public UID.
- API methods return decoded JSON objects as `map[string]any` so production can
  add response fields without breaking Go consumers.

## Error Handling

API calls return `*makepay.Error` for API responses outside the 2xx range. It
includes the HTTP status, decoded JSON response body, and raw response bytes.

```go
response, err := client.GetPaymentLink(ctx, "PAYMENT_LINK_UID")
if err != nil {
	var makePayError *makepay.Error
	if errors.As(err, &makePayError) {
		log.Println(makePayError.StatusCode, makePayError.ResponseBody)
	}
	return err
}

_ = response
```

## Source Layout

The canonical monorepo source lives in `apps/plugins/go-sdk`. The public
repository at `https://github.com/makecryptoio/makepay-go` mirrors only the SDK
files so pkg.go.dev and Go users can install or inspect it without the full
MakeCrypto workspace.
