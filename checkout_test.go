package makepay

import (
	"strings"
	"testing"
)

func TestCheckoutURLs(t *testing.T) {
	hosted, err := BuildHostedCheckoutURL("pay_123", CheckoutURLOptions{
		BaseURL: "https://pay.example/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if hosted != "https://pay.example/payment/pay_123" {
		t.Fatalf("unexpected hosted URL: %s", hosted)
	}

	embedded, err := BuildEmbeddedCheckoutURL("pay_123", CheckoutURLOptions{
		BaseURL:      "https://pay.example/",
		ParentOrigin: "https://merchant.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if embedded != "https://pay.example/embed/payment/pay_123?parentOrigin=https%3A%2F%2Fmerchant.example" {
		t.Fatalf("unexpected embedded URL: %s", embedded)
	}

	donation, err := BuildHostedDonationURL("spring campaign", CheckoutURLOptions{
		BaseURL: "https://pay.example/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if donation != "https://pay.example/donations/spring%20campaign" {
		t.Fatalf("unexpected donation URL: %s", donation)
	}

	embeddedDonation, err := BuildEmbeddedDonationURL("spring", CheckoutURLOptions{
		BaseURL:      "https://pay.example/",
		ParentOrigin: "https://merchant.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if embeddedDonation != "https://pay.example/embed/donations/spring?parentOrigin=https%3A%2F%2Fmerchant.example" {
		t.Fatalf("unexpected embedded donation URL: %s", embeddedDonation)
	}

	if got := BuildModalScriptURL(CheckoutURLOptions{BaseURL: "https://pay.example/"}); got != "https://pay.example/modal/makepay.js" {
		t.Fatalf("unexpected modal script URL: %s", got)
	}
}

func TestCheckoutHTMLSnippets(t *testing.T) {
	button, err := BuildEmbedButtonHTML(`pay_"<&`, EmbedSnippetOptions{
		ButtonLabel: "Pay <now>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(button, `data-makepay-payment-link="pay_&#34;&lt;&amp;"`) {
		t.Fatalf("payment UID was not escaped: %s", button)
	}
	if !strings.Contains(button, "Pay &lt;now&gt;") {
		t.Fatalf("button label was not escaped: %s", button)
	}

	iframe, err := BuildIframeHTML("pay_123", EmbedSnippetOptions{
		IframeTitle: "Secure checkout",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(iframe, `src="https://makepay.io/embed/payment/pay_123"`) {
		t.Fatalf("unexpected iframe snippet: %s", iframe)
	}
}
