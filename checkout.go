package makepay

import (
	"html"
	"net/url"
	"strings"
)

// CheckoutURLOptions customizes generated MakePay checkout URLs.
type CheckoutURLOptions struct {
	BaseURL      string
	ParentOrigin string
}

// EmbedSnippetOptions customizes generated MakePay HTML snippets.
type EmbedSnippetOptions struct {
	BaseURL      string
	ParentOrigin string
	ButtonLabel  string
	IframeTitle  string
}

// HostedCheckoutURL returns a full-page hosted checkout URL for a payment UID.
func (c *Client) HostedCheckoutURL(paymentUID string) (string, error) {
	return BuildHostedCheckoutURL(paymentUID, CheckoutURLOptions{
		BaseURL: c.checkoutBaseURL,
	})
}

// HostedDonationURL returns a full-page hosted donation URL for a donation slug.
func (c *Client) HostedDonationURL(donationSlug string) (string, error) {
	return BuildHostedDonationURL(donationSlug, CheckoutURLOptions{
		BaseURL: c.checkoutBaseURL,
	})
}

// EmbeddedCheckoutURL returns an iframe checkout URL for a payment UID.
func (c *Client) EmbeddedCheckoutURL(paymentUID string, parentOrigin string) (string, error) {
	return BuildEmbeddedCheckoutURL(paymentUID, CheckoutURLOptions{
		BaseURL:      c.checkoutBaseURL,
		ParentOrigin: parentOrigin,
	})
}

// EmbeddedDonationURL returns an iframe donation URL for a donation slug.
func (c *Client) EmbeddedDonationURL(donationSlug string, parentOrigin string) (string, error) {
	return BuildEmbeddedDonationURL(donationSlug, CheckoutURLOptions{
		BaseURL:      c.checkoutBaseURL,
		ParentOrigin: parentOrigin,
	})
}

// ModalScriptURL returns the MakePay modal script URL.
func (c *Client) ModalScriptURL() string {
	return BuildModalScriptURL(CheckoutURLOptions{
		BaseURL: c.checkoutBaseURL,
	})
}

// EmbedButtonHTML returns a small static HTML button snippet that loads the
// MakePay modal script.
func (c *Client) EmbedButtonHTML(paymentUID string, options EmbedSnippetOptions) (string, error) {
	options.BaseURL = c.checkoutBaseURL
	return BuildEmbedButtonHTML(paymentUID, options)
}

// IframeHTML returns a static iframe snippet for embedded MakePay checkout.
func (c *Client) IframeHTML(paymentUID string, options EmbedSnippetOptions) (string, error) {
	options.BaseURL = c.checkoutBaseURL
	return BuildIframeHTML(paymentUID, options)
}

// BuildHostedCheckoutURL returns a full-page hosted checkout URL for a payment UID.
func BuildHostedCheckoutURL(paymentUID string, options CheckoutURLOptions) (string, error) {
	if err := assertNonEmpty(paymentUID, "Payment link UID is required."); err != nil {
		return "", err
	}

	return buildURL(firstNonEmpty(options.BaseURL, DefaultCheckoutBaseURL), "/payment/"+url.PathEscape(paymentUID), nil)
}

// BuildHostedDonationURL returns a full-page hosted donation URL for a donation slug.
func BuildHostedDonationURL(donationSlug string, options CheckoutURLOptions) (string, error) {
	if err := assertNonEmpty(donationSlug, "Donation slug is required."); err != nil {
		return "", err
	}

	return buildURL(firstNonEmpty(options.BaseURL, DefaultCheckoutBaseURL), "/donations/"+url.PathEscape(donationSlug), nil)
}

// BuildEmbeddedCheckoutURL returns an iframe checkout URL for a payment UID.
func BuildEmbeddedCheckoutURL(paymentUID string, options CheckoutURLOptions) (string, error) {
	if err := assertNonEmpty(paymentUID, "Payment link UID is required."); err != nil {
		return "", err
	}

	query := url.Values{}
	if strings.TrimSpace(options.ParentOrigin) != "" {
		query.Set("parentOrigin", options.ParentOrigin)
	}

	return buildURL(firstNonEmpty(options.BaseURL, DefaultCheckoutBaseURL), "/embed/payment/"+url.PathEscape(paymentUID), query)
}

// BuildEmbeddedDonationURL returns an iframe donation URL for a donation slug.
func BuildEmbeddedDonationURL(donationSlug string, options CheckoutURLOptions) (string, error) {
	if err := assertNonEmpty(donationSlug, "Donation slug is required."); err != nil {
		return "", err
	}

	query := url.Values{}
	if strings.TrimSpace(options.ParentOrigin) != "" {
		query.Set("parentOrigin", options.ParentOrigin)
	}

	return buildURL(firstNonEmpty(options.BaseURL, DefaultCheckoutBaseURL), "/embed/donations/"+url.PathEscape(donationSlug), query)
}

// BuildModalScriptURL returns the MakePay modal script URL.
func BuildModalScriptURL(options CheckoutURLOptions) string {
	value, _ := buildURL(firstNonEmpty(options.BaseURL, DefaultCheckoutBaseURL), "/modal/makepay.js", nil)
	return value
}

// BuildEmbedButtonHTML returns a small static HTML button snippet that loads the
// MakePay modal script.
func BuildEmbedButtonHTML(paymentUID string, options EmbedSnippetOptions) (string, error) {
	if err := assertNonEmpty(paymentUID, "Payment link UID is required."); err != nil {
		return "", err
	}

	buttonLabel := firstNonEmpty(options.ButtonLabel, "Pay with crypto")
	return strings.Join([]string{
		`<script src="` + html.EscapeString(BuildModalScriptURL(CheckoutURLOptions{BaseURL: options.BaseURL})) + `"></script>`,
		`<button type="button" data-makepay-payment-link="` + html.EscapeString(paymentUID) + `">`,
		"  " + html.EscapeString(buttonLabel),
		"</button>",
	}, "\n"), nil
}

// BuildIframeHTML returns a static iframe snippet for embedded MakePay checkout.
func BuildIframeHTML(paymentUID string, options EmbedSnippetOptions) (string, error) {
	if err := assertNonEmpty(paymentUID, "Payment link UID is required."); err != nil {
		return "", err
	}

	src, err := BuildEmbeddedCheckoutURL(paymentUID, CheckoutURLOptions{
		BaseURL:      options.BaseURL,
		ParentOrigin: options.ParentOrigin,
	})
	if err != nil {
		return "", err
	}

	title := firstNonEmpty(options.IframeTitle, "MakePay checkout")
	return strings.Join([]string{
		"<iframe",
		`  title="` + html.EscapeString(title) + `"`,
		`  src="` + html.EscapeString(src) + `"`,
		`  style="width:100%;min-height:720px;border:0;border-radius:12px;"`,
		`  allow="clipboard-read; clipboard-write"`,
		"></iframe>",
	}, "\n"), nil
}

func buildURL(baseURL string, path string, query url.Values) (string, error) {
	parsed, err := url.Parse(normalizeBaseURL(baseURL))
	if err != nil {
		return "", err
	}

	joinedPath := strings.TrimRight(parsed.Path, "/") + path
	if unescapedPath, err := url.PathUnescape(joinedPath); err == nil && unescapedPath != joinedPath {
		parsed.Path = unescapedPath
		parsed.RawPath = joinedPath
	} else {
		parsed.Path = joinedPath
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
