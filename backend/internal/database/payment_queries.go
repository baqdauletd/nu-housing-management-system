package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
	"time"
)

var ErrPaymentNotFound = errors.New("payment not found")

func CreatePayment(db *sql.DB, payment models.Payment) (models.Payment, error) {
	query := `
		INSERT INTO payments (
			application_id,
			student_id,
			provider,
			status,
			amount_kzt,
			currency,
			payment_reference,
			provider_checkout_url,
			provider_message,
			customer_phone,
			metadata,
			initiated_at,
			expires_at,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, NOW(), $12, NOW(), NOW()
		)
		RETURNING
			id, application_id, student_id, provider, status, amount_kzt, currency, payment_reference,
			provider_checkout_url, provider_invoice_id, provider_transaction_id, provider_message, customer_phone,
			metadata, initiated_at, paid_at, expires_at, created_at, updated_at
	`

	var created models.Payment
	err := db.QueryRow(
		query,
		payment.ApplicationID,
		payment.StudentID,
		payment.Provider,
		payment.Status,
		payment.AmountKZT,
		payment.Currency,
		payment.PaymentReference,
		derefOptionalString(payment.ProviderCheckoutURL),
		derefOptionalString(payment.ProviderMessage),
		derefOptionalString(payment.CustomerPhone),
		payment.Metadata,
		payment.ExpiresAt,
	).Scan(
		&created.ID,
		&created.ApplicationID,
		&created.StudentID,
		&created.Provider,
		&created.Status,
		&created.AmountKZT,
		&created.Currency,
		&created.PaymentReference,
		&created.ProviderCheckoutURL,
		&created.ProviderInvoiceID,
		&created.ProviderTransactionID,
		&created.ProviderMessage,
		&created.CustomerPhone,
		&created.Metadata,
		&created.InitiatedAt,
		&created.PaidAt,
		&created.ExpiresAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return created, err
	}

	db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'create', 'payment', $2)`, payment.StudentID, created.ID)
	return created, nil
}

func GetLatestPaymentByApplication(db *sql.DB, applicationID int) (models.Payment, error) {
	query := `
		SELECT
			id, application_id, student_id, provider, status, amount_kzt, currency, payment_reference,
			provider_checkout_url, provider_invoice_id, provider_transaction_id, provider_message, customer_phone,
			metadata, initiated_at, paid_at, expires_at, created_at, updated_at
		FROM payments
		WHERE application_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`
	return scanPayment(db.QueryRow(query, applicationID))
}

func GetPaymentByID(db *sql.DB, paymentID int) (models.Payment, error) {
	query := `
		SELECT
			id, application_id, student_id, provider, status, amount_kzt, currency, payment_reference,
			provider_checkout_url, provider_invoice_id, provider_transaction_id, provider_message, customer_phone,
			metadata, initiated_at, paid_at, expires_at, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	return scanPayment(db.QueryRow(query, paymentID))
}

func GetPaymentByReference(db *sql.DB, reference string) (models.Payment, error) {
	query := `
		SELECT
			id, application_id, student_id, provider, status, amount_kzt, currency, payment_reference,
			provider_checkout_url, provider_invoice_id, provider_transaction_id, provider_message, customer_phone,
			metadata, initiated_at, paid_at, expires_at, created_at, updated_at
		FROM payments
		WHERE payment_reference = $1
	`
	return scanPayment(db.QueryRow(query, reference))
}

func GetPaymentByProviderInvoiceID(db *sql.DB, providerInvoiceID string) (models.Payment, error) {
	query := `
		SELECT
			id, application_id, student_id, provider, status, amount_kzt, currency, payment_reference,
			provider_checkout_url, provider_invoice_id, provider_transaction_id, provider_message, customer_phone,
			metadata, initiated_at, paid_at, expires_at, created_at, updated_at
		FROM payments
		WHERE provider_invoice_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`
	return scanPayment(db.QueryRow(query, providerInvoiceID))
}

func UpdatePaymentStatus(db *sql.DB, paymentID int, status string) error {
	_, err := db.Exec(
		`UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2`,
		status,
		paymentID,
	)
	return err
}

func UpdatePaymentStripeSession(
	db *sql.DB,
	paymentID int,
	sessionID string,
	checkoutURL string,
	transactionID string,
	status string,
	providerMessage string,
	expiresAt *time.Time,
) error {
	query := `
		UPDATE payments
		SET
			provider_invoice_id = COALESCE(NULLIF($1, ''), provider_invoice_id),
			provider_checkout_url = COALESCE(NULLIF($2, ''), provider_checkout_url),
			provider_transaction_id = COALESCE(NULLIF($3, ''), provider_transaction_id),
			status = COALESCE(NULLIF($4, ''), status),
			provider_message = COALESCE(NULLIF($5, ''), provider_message),
			expires_at = COALESCE($6, expires_at),
			updated_at = NOW()
		WHERE id = $7
	`
	_, err := db.Exec(query, sessionID, checkoutURL, transactionID, status, providerMessage, expiresAt, paymentID)
	return err
}

func MarkPaymentPaid(
	db *sql.DB,
	paymentID int,
	transactionID string,
	invoiceID string,
	providerMessage string,
	paidAt time.Time,
) error {
	query := `
		UPDATE payments
		SET
			status = 'paid',
			provider_transaction_id = NULLIF($1, ''),
			provider_invoice_id = NULLIF($2, ''),
			provider_message = COALESCE(NULLIF($3, ''), provider_message),
			paid_at = $4,
			updated_at = NOW()
		WHERE id = $5
	`
	_, err := db.Exec(query, transactionID, invoiceID, providerMessage, paidAt, paymentID)
	return err
}

func scanPayment(scanner interface {
	Scan(dest ...any) error
}) (models.Payment, error) {
	var payment models.Payment
	err := scanner.Scan(
		&payment.ID,
		&payment.ApplicationID,
		&payment.StudentID,
		&payment.Provider,
		&payment.Status,
		&payment.AmountKZT,
		&payment.Currency,
		&payment.PaymentReference,
		&payment.ProviderCheckoutURL,
		&payment.ProviderInvoiceID,
		&payment.ProviderTransactionID,
		&payment.ProviderMessage,
		&payment.CustomerPhone,
		&payment.Metadata,
		&payment.InitiatedAt,
		&payment.PaidAt,
		&payment.ExpiresAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return payment, ErrPaymentNotFound
	}
	return payment, err
}

func derefOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
