package makepay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// DefaultBaseURL is the MakeCrypto API origin used by the MakePay partner API.
	DefaultBaseURL = "https://www.makecrypto.io"
	// DefaultCheckoutBaseURL is the hosted MakePay checkout origin.
	DefaultCheckoutBaseURL = "https://makepay.io"
	// Version is the SDK version released through the public Go module.
	Version = "0.3.0"
)

// HTTPDoer is implemented by *http.Client and test clients.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientOptions configures a MakePay API client.
type ClientOptions struct {
	BaseURL         string
	CheckoutBaseURL string
	KeyID           string
	KeySecret       string
	HTTPClient      HTTPDoer
}

// Client sends authenticated requests to the MakePay partner API.
type Client struct {
	baseURL         string
	checkoutBaseURL string
	keyID           string
	keySecret       string
	httpClient      HTTPDoer
}

// PaymentLinkPayload is the merchant-defined payment link payload sent to
// MakePay. Fields match the MakePay partner API object model.
type PaymentLinkPayload map[string]any

// CreatePaymentLinkOptions controls payment-link creation behavior.
type CreatePaymentLinkOptions struct {
	Status                  string
	SendPaymentRequestEmail bool
}

// RequestOptions controls low-level API request behavior.
type RequestOptions struct {
	Query map[string]any
}

// PublicRequestOptions controls unauthenticated public MakePay API requests.
type PublicRequestOptions struct {
	BaseURL    string
	HTTPClient HTTPDoer
}

// Error is returned for invalid client configuration or non-2xx API responses.
type Error struct {
	StatusCode   int
	ResponseBody map[string]any
	Body         []byte
	Err          error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("MakePay API request failed with HTTP %d.", e.StatusCode)
	}
	return "MakePay API request failed."
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewClient creates a MakePay API client. KeyID and KeySecret are required and
// should be stored only on the merchant server.
func NewClient(options ClientOptions) (*Client, error) {
	baseURL := normalizeBaseURL(firstNonEmpty(options.BaseURL, DefaultBaseURL))
	checkoutBaseURL := normalizeBaseURL(firstNonEmpty(options.CheckoutBaseURL, DefaultCheckoutBaseURL))

	if strings.TrimSpace(baseURL) == "" {
		return nil, newClientError("MakePay base URL is required.")
	}
	if strings.TrimSpace(options.KeyID) == "" {
		return nil, newClientError("MakePay API key ID is required.")
	}
	if strings.TrimSpace(options.KeySecret) == "" {
		return nil, newClientError("MakePay API key secret is required.")
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:         baseURL,
		checkoutBaseURL: checkoutBaseURL,
		keyID:           options.KeyID,
		keySecret:       options.KeySecret,
		httpClient:      httpClient,
	}, nil
}

// CreatePaymentLink creates a hosted MakePay payment link.
func (c *Client) CreatePaymentLink(
	ctx context.Context,
	payload PaymentLinkPayload,
	options *CreatePaymentLinkOptions,
) (map[string]any, error) {
	status := "active"
	sendPaymentRequestEmail := false
	if options != nil {
		if strings.TrimSpace(options.Status) != "" {
			status = options.Status
		}
		sendPaymentRequestEmail = options.SendPaymentRequestEmail
	}

	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/payment-links", map[string]any{
		"status":                  status,
		"sendPaymentRequestEmail": sendPaymentRequestEmail,
		"payload":                 payload,
	}, RequestOptions{})
}

// ListPaymentLinks returns MakePay payment links visible to the API key.
func (c *Client) ListPaymentLinks(ctx context.Context, query map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/payment-links", nil, RequestOptions{
		Query: query,
	})
}

// GetPaymentLink returns one MakePay payment link by UID.
func (c *Client) GetPaymentLink(ctx context.Context, uid string) (map[string]any, error) {
	if err := assertNonEmpty(uid, "Payment link UID is required."); err != nil {
		return nil, err
	}

	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/payment-links/"+url.PathEscape(uid), nil, RequestOptions{})
}

// UpdatePaymentLink updates a MakePay payment link, such as changing its status.
func (c *Client) UpdatePaymentLink(ctx context.Context, uid string, updates map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(uid, "Payment link UID is required."); err != nil {
		return nil, err
	}

	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/payment-links/"+url.PathEscape(uid), updates, RequestOptions{})
}

// SendPaymentRequestEmail asks MakePay to email a payment request for a link.
func (c *Client) SendPaymentRequestEmail(ctx context.Context, uid string, email string) (map[string]any, error) {
	if err := assertNonEmpty(uid, "Payment link UID is required."); err != nil {
		return nil, err
	}

	body := map[string]any{}
	if strings.TrimSpace(email) != "" {
		body["email"] = email
	}

	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/payment-links/"+url.PathEscape(uid)+"/send-request-email", body, RequestOptions{})
}

// GetSettings returns the MakePay merchant settings for the API key.
func (c *Client) GetSettings(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/settings", nil, RequestOptions{})
}

// UpdateSettings updates MakePay merchant settings for the API key.
func (c *Client) UpdateSettings(ctx context.Context, settings map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPut, "/api/partner/v1/makepay/settings", settings, RequestOptions{})
}

// Request sends an authenticated MakePay partner API request and decodes the
// JSON object response.
func (c *Client) Request(
	ctx context.Context,
	method string,
	path string,
	body any,
	options RequestOptions,
) (map[string]any, error) {
	var reader io.Reader
	contentType := ""
	if body != nil && method != http.MethodGet {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, newClientError("Unable to encode MakePay request body as JSON.")
		}
		reader = bytes.NewReader(payload)
		contentType = "application/json"
	}

	return c.requestWithReader(ctx, method, path, reader, contentType, options)
}

func (c *Client) requestWithReader(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	contentType string,
	options RequestOptions,
) (map[string]any, error) {
	if c == nil {
		return nil, newClientError("MakePay client is required.")
	}

	requestURL, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	query := requestURL.Query()
	for key, value := range options.Query {
		if value != nil {
			query.Set(key, fmt.Sprint(value))
		}
	}
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MakePayGo/"+Version)
	req.Header.Set("X-MakeCrypto-Key-Id", c.keyID)
	req.Header.Set("X-MakeCrypto-Key-Secret", c.keySecret)
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	decoded := decodeObject(rawBody)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &Error{
			StatusCode:   response.StatusCode,
			ResponseBody: decoded,
			Body:         rawBody,
			Err:          errors.New(readErrorMessage(decoded, response.StatusCode)),
		}
	}

	return decoded, nil
}

// CreateAnonymousPaymentLink creates a public MakePay payment link without API
// credentials. Anonymous links must include an explicit settlement route.
func CreateAnonymousPaymentLink(
	ctx context.Context,
	payload AnonymousPaymentLinkPayload,
	options PublicRequestOptions,
) (map[string]any, error) {
	baseURL := normalizeBaseURL(firstNonEmpty(options.BaseURL, DefaultBaseURL))
	if strings.TrimSpace(baseURL) == "" {
		return nil, newClientError("MakePay base URL is required.")
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	requestURL, err := url.Parse(baseURL + "/api/partner/v1/makepay/payment-links")
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, newClientError("Unable to encode MakePay request body as JSON.")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MakePayGo/"+Version)

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	decoded := decodeObject(rawBody)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &Error{
			StatusCode:   response.StatusCode,
			ResponseBody: decoded,
			Body:         rawBody,
			Err:          errors.New(readErrorMessage(decoded, response.StatusCode)),
		}
	}

	return decoded, nil
}

// CreateAnonymousMakePayPaymentLink is an alias for CreateAnonymousPaymentLink.
func CreateAnonymousMakePayPaymentLink(
	ctx context.Context,
	payload AnonymousPaymentLinkPayload,
	options PublicRequestOptions,
) (map[string]any, error) {
	return CreateAnonymousPaymentLink(ctx, payload, options)
}

func decodeObject(rawBody []byte) map[string]any {
	if len(rawBody) == 0 {
		return map[string]any{}
	}

	var decoded map[string]any
	if err := json.Unmarshal(rawBody, &decoded); err != nil {
		return map[string]any{}
	}

	return decoded
}

func readErrorMessage(decoded map[string]any, status int) string {
	if message, ok := decoded["error"].(string); ok && message != "" {
		return message
	}

	return fmt.Sprintf("MakePay API request failed with HTTP %d.", status)
}

func assertNonEmpty(value string, message string) error {
	if strings.TrimSpace(value) == "" {
		return newClientError(message)
	}
	return nil
}

func newClientError(message string) *Error {
	return &Error{
		StatusCode: 400,
		Err:        errors.New(message),
	}
}

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func normalizeBaseURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/")
}
