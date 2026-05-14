# MakePay Go SDK

Official Go SDK for MakePay server-side integrations. Use it to create hosted
payment links, read and update links, manage MakePay settings, generate hosted
and embedded checkout URLs, and verify signed webhooks.

## Install

```bash
go get github.com/makecryptoio/makepay-go
```

The package is published through the public Go module index:

```text
https://pkg.go.dev/github.com/makecryptoio/makepay-go
```

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

## Create a hosted checkout link

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

Send a MakePay payment request email during creation:

```go
_, err := client.CreatePaymentLink(ctx, payload, &makepay.CreatePaymentLinkOptions{
	SendPaymentRequestEmail: true,
})
```

## Hosted and embedded checkout

Use hosted checkout URLs for redirects, or the embed helpers when your frontend
keeps the shopper on the merchant page.

```go
hostedURL, err := client.HostedCheckoutURL("PAYMENT_LINK_UID")
if err != nil {
	return err
}

embeddedURL, err := client.EmbeddedCheckoutURL(
	"PAYMENT_LINK_UID",
	"https://merchant.example",
)
if err != nil {
	return err
}

buttonHTML, err := client.EmbedButtonHTML("PAYMENT_LINK_UID", makepay.EmbedSnippetOptions{
	ButtonLabel: "Pay with crypto",
})
if err != nil {
	return err
}

iframeHTML, err := client.IframeHTML("PAYMENT_LINK_UID", makepay.EmbedSnippetOptions{
	IframeTitle: "Secure MakePay checkout",
})
if err != nil {
	return err
}

log.Println(hostedURL, embeddedURL, buttonHTML, iframeHTML)
```

## Read and update payment links

```go
links, err := client.ListPaymentLinks(ctx, map[string]any{"limit": 50})
detail, err := client.GetPaymentLink(ctx, "PAYMENT_LINK_UID")
updated, err := client.UpdatePaymentLink(ctx, "PAYMENT_LINK_UID", map[string]any{
	"status": "paused",
})
sent, err := client.SendPaymentRequestEmail(ctx, "PAYMENT_LINK_UID", "buyer@example.com")

_, _, _, _ = links, detail, updated, sent
```

## Settings

```go
settings, err := client.GetSettings(ctx)
updated, err := client.UpdateSettings(ctx, map[string]any{
	"callbackUrl": "https://merchant.example/webhooks/makepay",
})

_, _ = settings, updated
```

## Verify webhooks

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

## Error handling

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

## Source layout

The canonical monorepo source lives in `apps/plugins/go-sdk`. The public
repository at `https://github.com/makecryptoio/makepay-go` mirrors only the SDK
files so pkg.go.dev and Go users can install or inspect it without the full
MakeCrypto workspace.
