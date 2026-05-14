package makepay

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// CreateDonationLink creates a hosted MakePay donation page.
func (c *Client) CreateDonationLink(
	ctx context.Context,
	payload DonationLinkPayload,
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

	body := map[string]any{
		"status":                  status,
		"sendPaymentRequestEmail": sendPaymentRequestEmail,
		"payload":                 payload,
	}
	if payload != nil {
		body["payload"] = copyPayloadWith(payload, "type", "donation")
	}

	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/donations", body, RequestOptions{})
}

// ListDonationLinks returns MakePay donation links visible to the API key.
func (c *Client) ListDonationLinks(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/donations", nil, RequestOptions{})
}

// GetDonationLink returns one MakePay donation link by UID.
func (c *Client) GetDonationLink(ctx context.Context, uid string) (map[string]any, error) {
	if err := assertNonEmpty(uid, "Donation link UID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/donations/"+url.PathEscape(uid), nil, RequestOptions{})
}

// UpdateDonationLink updates a MakePay donation link.
func (c *Client) UpdateDonationLink(ctx context.Context, uid string, updates map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(uid, "Donation link UID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/donations/"+url.PathEscape(uid), updates, RequestOptions{})
}

// ListCustomers returns MakePay customers.
func (c *Client) ListCustomers(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/customers", nil, RequestOptions{})
}

// UpsertCustomer creates or updates a MakePay customer.
func (c *Client) UpsertCustomer(ctx context.Context, payload CustomerPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/customers", payload, RequestOptions{})
}

// CreateCustomerPortal creates a customer portal session.
func (c *Client) CreateCustomerPortal(ctx context.Context, customerID string, payload map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(customerID, "Customer ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/customers/"+url.PathEscape(customerID)+"/portal", payload, RequestOptions{})
}

// ListSubscriptions returns MakePay subscriptions.
func (c *Client) ListSubscriptions(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/subscriptions", nil, RequestOptions{})
}

// CreateSubscription creates a MakePay subscription.
func (c *Client) CreateSubscription(ctx context.Context, payload SubscriptionPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/subscriptions", payload, RequestOptions{})
}

// ListDestinationAssets returns destination assets available to MakePay.
func (c *Client) ListDestinationAssets(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/destination-assets", nil, RequestOptions{})
}

// ListWebhookRequests returns recent MakePay webhook requests.
func (c *Client) ListWebhookRequests(ctx context.Context, query map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/webhook-requests", nil, RequestOptions{Query: query})
}

// ListPosTerminals returns MakePay POS terminals.
func (c *Client) ListPosTerminals(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/pos-terminals", nil, RequestOptions{})
}

// CreatePosTerminal creates a MakePay POS terminal.
func (c *Client) CreatePosTerminal(ctx context.Context, payload PosTerminalPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/pos-terminals", payload, RequestOptions{})
}

// GetPosTerminal returns one MakePay POS terminal.
func (c *Client) GetPosTerminal(ctx context.Context, terminalID string) (map[string]any, error) {
	if err := assertNonEmpty(terminalID, "POS terminal ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/pos-terminals/"+url.PathEscape(terminalID), nil, RequestOptions{})
}

// UpdatePosTerminal updates a MakePay POS terminal.
func (c *Client) UpdatePosTerminal(ctx context.Context, terminalID string, payload PosTerminalPayload) (map[string]any, error) {
	if err := assertNonEmpty(terminalID, "POS terminal ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/pos-terminals/"+url.PathEscape(terminalID), payload, RequestOptions{})
}

// ListProducts returns MakePay Simple Shop products.
func (c *Client) ListProducts(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/products", nil, RequestOptions{})
}

// CreateProduct creates a MakePay Simple Shop product.
func (c *Client) CreateProduct(ctx context.Context, payload ProductPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/products", payload, RequestOptions{})
}

// GetProduct returns one MakePay product.
func (c *Client) GetProduct(ctx context.Context, productID string) (map[string]any, error) {
	if err := assertNonEmpty(productID, "Product ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/products/"+url.PathEscape(productID), nil, RequestOptions{})
}

// UpdateProduct updates a MakePay product.
func (c *Client) UpdateProduct(ctx context.Context, productID string, payload ProductPayload) (map[string]any, error) {
	if err := assertNonEmpty(productID, "Product ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/products/"+url.PathEscape(productID), payload, RequestOptions{})
}

// ListProductDownloads returns downloads for a MakePay product.
func (c *Client) ListProductDownloads(ctx context.Context, productID string) (map[string]any, error) {
	if err := assertNonEmpty(productID, "Product ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop/products/"+url.PathEscape(productID)+"/downloads", nil, RequestOptions{})
}

// CreateProductDownload creates a download for a MakePay product.
func (c *Client) CreateProductDownload(ctx context.Context, productID string, payload map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(productID, "Product ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/shop/products/"+url.PathEscape(productID)+"/downloads", payload, RequestOptions{})
}

// GetShop returns the MakePay Simple Shop.
func (c *Client) GetShop(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop", nil, RequestOptions{})
}

// UpdateShop updates the MakePay Simple Shop.
func (c *Client) UpdateShop(ctx context.Context, payload ShopPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/shop", payload, RequestOptions{})
}

// GetShopBuilder returns Simple Shop builder blocks and settings.
func (c *Client) GetShopBuilder(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop/builder", nil, RequestOptions{})
}

// UpdateShopBuilder updates Simple Shop builder blocks and settings.
func (c *Client) UpdateShopBuilder(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPut, "/api/partner/v1/makepay/shop/builder", payload, RequestOptions{})
}

// GetShopDomain returns Simple Shop custom-domain status.
func (c *Client) GetShopDomain(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop/domains", nil, RequestOptions{})
}

// UpdateShopDomain updates the Simple Shop custom domain. Pass a string,
// nil, or a map with a domain field.
func (c *Client) UpdateShopDomain(ctx context.Context, input any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPut, "/api/partner/v1/makepay/shop/domains", normalizeDomainInput(input), RequestOptions{})
}

// RefreshShopDomain refreshes Simple Shop domain verification.
func (c *Client) RefreshShopDomain(ctx context.Context, input any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/shop/domains", normalizeDomainInput(input), RequestOptions{})
}

// ListShopCoupons returns Simple Shop coupons.
func (c *Client) ListShopCoupons(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop/coupons", nil, RequestOptions{})
}

// CreateShopCoupon creates a Simple Shop coupon.
func (c *Client) CreateShopCoupon(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/shop/coupons", payload, RequestOptions{})
}

// UpdateShopCoupon updates a Simple Shop coupon.
func (c *Client) UpdateShopCoupon(ctx context.Context, couponUID string, payload map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(couponUID, "Shop coupon UID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/shop/coupons/"+url.PathEscape(couponUID), payload, RequestOptions{})
}

// ArchiveShopCoupon archives a Simple Shop coupon.
func (c *Client) ArchiveShopCoupon(ctx context.Context, couponUID string) (map[string]any, error) {
	if err := assertNonEmpty(couponUID, "Shop coupon UID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodDelete, "/api/partner/v1/makepay/shop/coupons/"+url.PathEscape(couponUID), nil, RequestOptions{})
}

// ListShopOrders returns Simple Shop orders.
func (c *Client) ListShopOrders(ctx context.Context, query map[string]any) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/shop/orders", nil, RequestOptions{Query: query})
}

// GetBranding returns MakePay branding and domain settings.
func (c *Client) GetBranding(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/branding", nil, RequestOptions{})
}

// UpdateBranding updates MakePay branding and domain settings.
func (c *Client) UpdateBranding(ctx context.Context, payload BrandingPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/branding", payload, RequestOptions{})
}

// RefreshBrandingDomains refreshes payment-link and email-sending domain checks.
func (c *Client) RefreshBrandingDomains(ctx context.Context, kind string) (map[string]any, error) {
	if strings.TrimSpace(kind) == "" {
		kind = "all"
	}
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/branding/domains/refresh", map[string]any{"kind": kind}, RequestOptions{})
}

// GetBookkeepingSummary returns bookkeeping dashboard data.
func (c *Client) GetBookkeepingSummary(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping", nil, RequestOptions{})
}

// ListBookkeepingInvoices returns bookkeeping invoices.
func (c *Client) ListBookkeepingInvoices(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/invoices", nil, RequestOptions{})
}

// CreateBookkeepingInvoice creates a bookkeeping invoice.
func (c *Client) CreateBookkeepingInvoice(ctx context.Context, payload BookkeepingInvoicePayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/invoices", payload, RequestOptions{})
}

// GetBookkeepingInvoice returns one bookkeeping invoice.
func (c *Client) GetBookkeepingInvoice(ctx context.Context, invoiceID string) (map[string]any, error) {
	if err := assertNonEmpty(invoiceID, "Bookkeeping invoice ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/invoices/"+url.PathEscape(invoiceID), nil, RequestOptions{})
}

// UpdateBookkeepingInvoice updates a bookkeeping invoice.
func (c *Client) UpdateBookkeepingInvoice(ctx context.Context, invoiceID string, payload BookkeepingInvoicePayload) (map[string]any, error) {
	if err := assertNonEmpty(invoiceID, "Bookkeeping invoice ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/bookkeeping/invoices/"+url.PathEscape(invoiceID), payload, RequestOptions{})
}

// CreateBookkeepingInvoicePaymentLink creates a MakePay payment link for an invoice.
func (c *Client) CreateBookkeepingInvoicePaymentLink(ctx context.Context, invoiceID string, options map[string]any) (map[string]any, error) {
	if err := assertNonEmpty(invoiceID, "Bookkeeping invoice ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/invoices/"+url.PathEscape(invoiceID)+"/payment-link", options, RequestOptions{})
}

// ListBookkeepingExpenses returns bookkeeping expenses.
func (c *Client) ListBookkeepingExpenses(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/expenses", nil, RequestOptions{})
}

// CreateBookkeepingExpense creates a bookkeeping expense.
func (c *Client) CreateBookkeepingExpense(ctx context.Context, payload BookkeepingExpensePayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/expenses", payload, RequestOptions{})
}

// CreateBookkeepingExpenseFromActivity creates an expense from wallet activity.
func (c *Client) CreateBookkeepingExpenseFromActivity(ctx context.Context, payload BookkeepingExpensePayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/expenses/from-activity", payload, RequestOptions{})
}

// GetBookkeepingExpense returns one bookkeeping expense.
func (c *Client) GetBookkeepingExpense(ctx context.Context, expenseID string) (map[string]any, error) {
	if err := assertNonEmpty(expenseID, "Bookkeeping expense ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/expenses/"+url.PathEscape(expenseID), nil, RequestOptions{})
}

// UpdateBookkeepingExpense updates a bookkeeping expense.
func (c *Client) UpdateBookkeepingExpense(ctx context.Context, expenseID string, payload BookkeepingExpensePayload) (map[string]any, error) {
	if err := assertNonEmpty(expenseID, "Bookkeeping expense ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPatch, "/api/partner/v1/makepay/bookkeeping/expenses/"+url.PathEscape(expenseID), payload, RequestOptions{})
}

// ListBookkeepingDocuments returns bookkeeping documents.
func (c *Client) ListBookkeepingDocuments(ctx context.Context) (map[string]any, error) {
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/documents", nil, RequestOptions{})
}

// UploadBookkeepingDocument uploads a supporting bookkeeping document.
func (c *Client) UploadBookkeepingDocument(ctx context.Context, input BookkeepingDocumentUpload) (map[string]any, error) {
	if input.File == nil {
		return nil, newClientError("Bookkeeping document file is required.")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileName := firstNonEmpty(input.FileName, "document")
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, input.File); err != nil {
		return nil, err
	}
	writeMultipartField(writer, "documentType", input.DocumentType)
	writeMultipartField(writer, "invoiceId", input.InvoiceID)
	writeMultipartField(writer, "expenseId", input.ExpenseID)
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return c.requestWithReader(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/documents", &body, writer.FormDataContentType(), RequestOptions{})
}

// GetBookkeepingDocumentDownloadURL returns a temporary document download URL.
func (c *Client) GetBookkeepingDocumentDownloadURL(ctx context.Context, documentID string) (map[string]any, error) {
	if err := assertNonEmpty(documentID, "Bookkeeping document ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodGet, "/api/partner/v1/makepay/bookkeeping/documents/"+url.PathEscape(documentID)+"/download", nil, RequestOptions{})
}

// RunBookkeepingDocumentOCR runs OCR for a bookkeeping document.
func (c *Client) RunBookkeepingDocumentOCR(ctx context.Context, documentID string) (map[string]any, error) {
	if err := assertNonEmpty(documentID, "Bookkeeping document ID is required."); err != nil {
		return nil, err
	}
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/documents/"+url.PathEscape(documentID)+"/ocr", map[string]any{}, RequestOptions{})
}

// CreateBookkeepingReconciliation creates a bookkeeping reconciliation link.
func (c *Client) CreateBookkeepingReconciliation(ctx context.Context, payload BookkeepingReconciliationPayload) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/partner/v1/makepay/bookkeeping/reconciliation", payload, RequestOptions{})
}

func copyPayloadWith(payload map[string]any, key string, value any) map[string]any {
	copied := make(map[string]any, len(payload)+1)
	for payloadKey, payloadValue := range payload {
		copied[payloadKey] = payloadValue
	}
	copied[key] = value
	return copied
}

func normalizeDomainInput(input any) any {
	switch value := input.(type) {
	case string:
		return map[string]any{"domain": value}
	case nil:
		return map[string]any{"domain": nil}
	default:
		return value
	}
}

func writeMultipartField(writer *multipart.Writer, key string, value string) {
	if strings.TrimSpace(value) != "" {
		_ = writer.WriteField(key, value)
	}
}
