package models

import "time"

type Payment struct {
	ID                    int        `json:"id"`
	ApplicationID         int        `json:"application_id"`
	StudentID             int        `json:"student_id"`
	Provider              string     `json:"provider"`
	Status                string     `json:"status"`
	AmountKZT             int        `json:"amount_kzt"`
	Currency              string     `json:"currency"`
	PaymentReference      string     `json:"payment_reference"`
	ProviderCheckoutURL   *string    `json:"provider_checkout_url,omitempty"`
	ProviderInvoiceID     *string    `json:"provider_invoice_id,omitempty"`
	ProviderTransactionID *string    `json:"provider_transaction_id,omitempty"`
	ProviderMessage       *string    `json:"provider_message,omitempty"`
	CustomerPhone         *string    `json:"customer_phone,omitempty"`
	Metadata              []byte     `json:"metadata,omitempty"`
	InitiatedAt           time.Time  `json:"initiated_at"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
