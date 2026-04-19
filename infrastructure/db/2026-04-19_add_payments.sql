CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    application_id INT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    student_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    amount_kzt INT NOT NULL CHECK (amount_kzt > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'KZT',
    payment_reference VARCHAR(64) NOT NULL UNIQUE,
    provider_checkout_url TEXT,
    provider_invoice_id TEXT,
    provider_transaction_id TEXT,
    provider_message TEXT,
    customer_phone VARCHAR(20),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    initiated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_application_id ON payments(application_id);
CREATE INDEX idx_payments_student_id ON payments(student_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_application_created_at ON payments(application_id, created_at DESC, id DESC);
