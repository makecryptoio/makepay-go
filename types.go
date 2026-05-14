package makepay

import "io"

// DonationLinkPayload is the merchant-defined donation page payload.
type DonationLinkPayload map[string]any

// BrandingPayload updates MakePay branding and custom-domain settings.
type BrandingPayload map[string]any

// CustomerPayload creates or updates a MakePay customer.
type CustomerPayload map[string]any

// SubscriptionPayload creates a MakePay subscription.
type SubscriptionPayload map[string]any

// PosTerminalPayload creates or updates a MakePay POS terminal.
type PosTerminalPayload map[string]any

// ProductPayload creates or updates a MakePay product.
type ProductPayload map[string]any

// ShopPayload updates a MakePay Simple Shop storefront.
type ShopPayload map[string]any

// AnonymousPaymentLinkPayload creates a public payment link without API keys.
type AnonymousPaymentLinkPayload map[string]any

// BookkeepingCounterpartyPayload creates invoice or expense counterparties.
type BookkeepingCounterpartyPayload map[string]any

// BookkeepingInvoiceLineItemPayload creates invoice line items.
type BookkeepingInvoiceLineItemPayload map[string]any

// BookkeepingInvoicePayload creates or updates bookkeeping invoices.
type BookkeepingInvoicePayload map[string]any

// BookkeepingExpensePayload creates or updates bookkeeping expenses.
type BookkeepingExpensePayload map[string]any

// BookkeepingReconciliationPayload links invoices, expenses, payments, and
// wallet activity for bookkeeping reconciliation.
type BookkeepingReconciliationPayload map[string]any

// BookkeepingDocumentUpload uploads a supporting bookkeeping document.
type BookkeepingDocumentUpload struct {
	File         io.Reader
	FileName     string
	DocumentType string
	InvoiceID    string
	ExpenseID    string
}
