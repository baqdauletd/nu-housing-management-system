package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nu-housing-management-system/backend/internal/config"
	"nu-housing-management-system/backend/internal/database"
	"nu-housing-management-system/backend/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	paymentStatusPending = "pending"
	paymentStatusPaid    = "paid"
	paymentStatusExpired = "expired"

	stripeAPIBase = "https://api.stripe.com/v1"
)

type stripeCheckoutSession struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	PaymentIntent string `json:"payment_intent"`
	ExpiresAt     int64  `json:"expires_at"`
}

type stripeWebhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object stripeCheckoutSession `json:"object"`
	} `json:"data"`
}

func GetApplicationPaymentSummary(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		app, ok := authorizeStudentPaymentAccess(c, db)
		if !ok {
			return
		}

		payment, err := getFreshLatestPayment(db, cfg, app.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load payment", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, buildPaymentSummaryResponse(cfg, app, payment))
	}
}

func InitiateApplicationPayment(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		app, ok := authorizeStudentPaymentAccess(c, db)
		if !ok {
			return
		}

		if app.Status != "approved" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only approved applications can be paid"})
			return
		}

		if cfg.StripePaymentAmountKZT <= 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_PAYMENT_AMOUNT_KZT is not configured"})
			return
		}
		if strings.TrimSpace(cfg.StripeSecretKey) == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_SECRET_KEY is not configured"})
			return
		}
		if strings.TrimSpace(cfg.FrontendBaseURL) == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "FRONTEND_BASE_URL is not configured"})
			return
		}

		if existing, ok := getReusablePayment(db, cfg, app.ID); ok {
			c.JSON(http.StatusOK, buildPaymentSummaryResponse(cfg, app, &existing))
			return
		}

		uid := c.GetInt("user_id")
		user, err := database.GetUserByID(db, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load student profile", "details": err.Error()})
			return
		}

		payment, err := createStripePayment(db, cfg, app, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, buildPaymentSummaryResponse(cfg, app, &payment))
	}
}

func SyncStripePayment(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		app, ok := authorizeStudentPaymentAccess(c, db)
		if !ok {
			return
		}

		payment, err := database.GetLatestPaymentByApplication(db, app.ID)
		if err != nil {
			if err == database.ErrPaymentNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load payment", "details": err.Error()})
			return
		}

		if payment.Status == paymentStatusPaid {
			c.JSON(http.StatusOK, buildPaymentSummaryResponse(cfg, app, &payment))
			return
		}

		sessionID := strings.TrimSpace(c.Query("session_id"))
		if sessionID == "" {
			sessionID = strings.TrimSpace(derefPaymentString(payment.ProviderInvoiceID))
		}
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "stripe session id is required"})
			return
		}

		session, err := retrieveStripeCheckoutSession(cfg.StripeSecretKey, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve stripe session", "details": err.Error()})
			return
		}

		if err := applyStripeSessionState(db, payment, session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync payment", "details": err.Error()})
			return
		}

		updated, err := database.GetPaymentByID(db, payment.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload payment", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, buildPaymentSummaryResponse(cfg, app, &updated))
	}
}

func HandleStripeWebhook(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(cfg.StripeWebhookSecret) == "" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "STRIPE_WEBHOOK_SECRET is not configured"})
			return
		}

		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		signature := c.GetHeader("Stripe-Signature")
		if err := verifyStripeSignature(payload, signature, cfg.StripeWebhookSecret); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid stripe signature", "details": err.Error()})
			return
		}

		var event stripeWebhookEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stripe event payload", "details": err.Error()})
			return
		}

		switch event.Type {
		case "checkout.session.completed", "checkout.session.async_payment_succeeded":
			if err := syncPaymentByStripeSessionID(db, event.Data.Object.ID, &event.Data.Object); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment", "details": err.Error()})
				return
			}
		case "checkout.session.expired", "checkout.session.async_payment_failed":
			if err := markPaymentExpiredByStripeSessionID(db, event.Data.Object.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"received": true})
	}
}

func authorizeStudentPaymentAccess(c *gin.Context, db *sql.DB) (models.Application, bool) {
	applicationID, err := strconv.Atoi(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid application id"})
		return models.Application{}, false
	}

	app, err := database.GetApplicationByID(db, applicationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return models.Application{}, false
	}

	uid := c.GetInt("user_id")
	if app.StudentID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return models.Application{}, false
	}

	return app, true
}

func buildPaymentSummaryResponse(cfg *config.Config, app models.Application, payment *models.Payment) gin.H {
	response := gin.H{
		"application": gin.H{
			"id":              app.ID,
			"status":          app.Status,
			"major":           app.Major,
			"year":            app.Year,
			"room_preference": app.RoomPreference,
			"submitted_at":    app.SubmittedAt,
		},
		"payable":       app.Status == "approved",
		"provider":      "stripe",
		"merchant_name": cfg.StripeMerchantName,
		"amount_kzt":    cfg.StripePaymentAmountKZT,
		"currency":      strings.ToUpper(cfg.StripePaymentCurrency),
		"payment":       nil,
	}

	if payment == nil {
		return response
	}

	response["payment"] = gin.H{
		"id":                      payment.ID,
		"application_id":          payment.ApplicationID,
		"status":                  payment.Status,
		"amount_kzt":              payment.AmountKZT,
		"currency":                payment.Currency,
		"payment_reference":       payment.PaymentReference,
		"provider_checkout_url":   payment.ProviderCheckoutURL,
		"provider_session_id":     payment.ProviderInvoiceID,
		"provider_transaction_id": payment.ProviderTransactionID,
		"provider_message":        payment.ProviderMessage,
		"customer_phone":          payment.CustomerPhone,
		"initiated_at":            payment.InitiatedAt,
		"paid_at":                 payment.PaidAt,
		"expires_at":              payment.ExpiresAt,
		"created_at":              payment.CreatedAt,
		"updated_at":              payment.UpdatedAt,
		"instructions":            buildStripeInstructions(payment),
	}

	return response
}

func createStripePayment(db *sql.DB, cfg *config.Config, app models.Application, user models.User) (models.Payment, error) {
	reference := fmt.Sprintf("APP-%d-%d", app.ID, time.Now().UTC().Unix())
	successURL := fmt.Sprintf("%s/dashboard/student/payment?applicationId=%d&stripe_session_id={CHECKOUT_SESSION_ID}&payment=success", strings.TrimRight(cfg.FrontendBaseURL, "/"), app.ID)
	cancelURL := fmt.Sprintf("%s/dashboard/student/payment?applicationId=%d&payment=cancelled", strings.TrimRight(cfg.FrontendBaseURL, "/"), app.ID)

	session, err := createStripeCheckoutSession(cfg.StripeSecretKey, stripeCheckoutSessionParams{
		AmountKZT:       cfg.StripePaymentAmountKZT,
		Currency:        cfg.StripePaymentCurrency,
		SuccessURL:      successURL,
		CancelURL:       cancelURL,
		Reference:       reference,
		ApplicationID:   app.ID,
		StudentID:       user.ID,
		StudentEmail:    user.Email,
		ProductName:     cfg.StripeProductName,
		ProductDesc:     cfg.StripePaymentDescription,
		ClientReference: strconv.Itoa(app.ID),
	})
	if err != nil {
		return models.Payment{}, err
	}

	var expiresAt *time.Time
	if session.ExpiresAt > 0 {
		value := time.Unix(session.ExpiresAt, 0).UTC()
		expiresAt = &value
	}

	metadata, err := json.Marshal(gin.H{
		"application_id": app.ID,
		"student_id":     user.ID,
		"student_nu_id":  user.NuID,
		"student_email":  user.Email,
		"stripe_session": session.ID,
	})
	if err != nil {
		return models.Payment{}, err
	}

	message := strings.TrimSpace(cfg.StripePaymentDescription)
	if message == "" {
		message = "NU Housing accommodation payment"
	}

	payment := models.Payment{
		ApplicationID:         app.ID,
		StudentID:             app.StudentID,
		Provider:              "stripe",
		Status:                mapStripePaymentStatus(session.Status, session.PaymentStatus),
		AmountKZT:             cfg.StripePaymentAmountKZT,
		Currency:              strings.ToUpper(cfg.StripePaymentCurrency),
		PaymentReference:      reference,
		ProviderCheckoutURL:   stringPointer(session.URL),
		ProviderInvoiceID:     stringPointer(session.ID),
		ProviderTransactionID: stringPointer(session.PaymentIntent),
		ProviderMessage:       stringPointer(message),
		CustomerPhone:         sanitizePhonePointer(user.Phone),
		Metadata:              metadata,
		ExpiresAt:             expiresAt,
	}

	return database.CreatePayment(db, payment)
}

func createStripeCheckoutSession(secretKey string, params stripeCheckoutSessionParams) (stripeCheckoutSession, error) {
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", params.SuccessURL)
	form.Set("cancel_url", params.CancelURL)
	form.Set("currency", params.Currency)
	form.Set("customer_email", params.StudentEmail)
	form.Set("client_reference_id", params.ClientReference)
	form.Set("payment_method_types[0]", "card")
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", params.Currency)
	form.Set("line_items[0][price_data][unit_amount]", strconv.Itoa(params.AmountKZT*100))
	form.Set("line_items[0][price_data][product_data][name]", nonEmpty(params.ProductName, "NU Housing payment"))
	form.Set("line_items[0][price_data][product_data][description]", nonEmpty(params.ProductDesc, "NU Housing accommodation payment"))
	form.Set("metadata[payment_reference]", params.Reference)
	form.Set("metadata[application_id]", strconv.Itoa(params.ApplicationID))
	form.Set("metadata[student_id]", strconv.Itoa(params.StudentID))

	request, err := http.NewRequest(http.MethodPost, stripeAPIBase+"/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return stripeCheckoutSession{}, err
	}

	request.SetBasicAuth(secretKey, "")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return stripeCheckoutSession{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return stripeCheckoutSession{}, err
	}

	if response.StatusCode >= 400 {
		return stripeCheckoutSession{}, fmt.Errorf("stripe create checkout session failed: %s", compactStripeError(body))
	}

	var session stripeCheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		return stripeCheckoutSession{}, err
	}
	return session, nil
}

func retrieveStripeCheckoutSession(secretKey, sessionID string) (stripeCheckoutSession, error) {
	request, err := http.NewRequest(http.MethodGet, stripeAPIBase+"/checkout/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		return stripeCheckoutSession{}, err
	}

	request.SetBasicAuth(secretKey, "")
	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return stripeCheckoutSession{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return stripeCheckoutSession{}, err
	}
	if response.StatusCode >= 400 {
		return stripeCheckoutSession{}, fmt.Errorf("stripe retrieve checkout session failed: %s", compactStripeError(body))
	}

	var session stripeCheckoutSession
	if err := json.Unmarshal(body, &session); err != nil {
		return stripeCheckoutSession{}, err
	}
	return session, nil
}

func verifyStripeSignature(payload []byte, headerValue, secret string) error {
	headerValue = strings.TrimSpace(headerValue)
	if headerValue == "" {
		return fmt.Errorf("missing Stripe-Signature header")
	}

	var timestamp string
	var signatures []string
	for _, part := range strings.Split(headerValue, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestamp = value
		case "v1":
			signatures = append(signatures, value)
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return fmt.Errorf("malformed Stripe-Signature header")
	}

	issuedAt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid stripe signature timestamp")
	}
	if delta := time.Since(time.Unix(issuedAt, 0)); delta > 5*time.Minute || delta < -5*time.Minute {
		return fmt.Errorf("stripe signature timestamp outside tolerance")
	}

	signedPayload := append([]byte(timestamp+"."), payload...)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(signedPayload)
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, signature := range signatures {
		if hmac.Equal([]byte(signature), []byte(expected)) {
			return nil
		}
	}

	return fmt.Errorf("no matching stripe signature")
}

func syncPaymentByStripeSessionID(db *sql.DB, sessionID string, session *stripeCheckoutSession) error {
	payment, err := database.GetPaymentByProviderInvoiceID(db, sessionID)
	if err != nil {
		if err == database.ErrPaymentNotFound {
			return nil
		}
		return err
	}

	currentSession := session
	if currentSession == nil {
		currentSession = &stripeCheckoutSession{ID: sessionID}
	}
	return applyStripeSessionState(db, payment, *currentSession)
}

func markPaymentExpiredByStripeSessionID(db *sql.DB, sessionID string) error {
	payment, err := database.GetPaymentByProviderInvoiceID(db, sessionID)
	if err != nil {
		if err == database.ErrPaymentNotFound {
			return nil
		}
		return err
	}
	return database.UpdatePaymentStatus(db, payment.ID, paymentStatusExpired)
}

func applyStripeSessionState(db *sql.DB, payment models.Payment, session stripeCheckoutSession) error {
	status := mapStripePaymentStatus(session.Status, session.PaymentStatus)
	switch status {
	case paymentStatusPaid:
		return database.MarkPaymentPaid(db, payment.ID, session.PaymentIntent, session.ID, session.PaymentStatus, time.Now().UTC())
	case paymentStatusExpired:
		return database.UpdatePaymentStatus(db, payment.ID, paymentStatusExpired)
	default:
		return database.UpdatePaymentStripeSession(db, payment.ID, session.ID, session.URL, session.PaymentIntent, status, session.PaymentStatus, stripeExpiresAtPointer(session.ExpiresAt))
	}
}

func getReusablePayment(db *sql.DB, cfg *config.Config, applicationID int) (models.Payment, bool) {
	payment, err := getFreshLatestPayment(db, cfg, applicationID)
	if err != nil || payment == nil {
		return models.Payment{}, false
	}

	if payment.Status == paymentStatusPending {
		return *payment, true
	}
	return models.Payment{}, false
}

func getFreshLatestPayment(db *sql.DB, cfg *config.Config, applicationID int) (*models.Payment, error) {
	payment, err := database.GetLatestPaymentByApplication(db, applicationID)
	if err != nil {
		if err == database.ErrPaymentNotFound {
			return nil, nil
		}
		return nil, err
	}

	if payment.Provider == "stripe" && strings.TrimSpace(derefPaymentString(payment.ProviderInvoiceID)) != "" && payment.Status != paymentStatusPaid {
		session, err := retrieveStripeCheckoutSession(cfg.StripeSecretKey, derefPaymentString(payment.ProviderInvoiceID))
		if err == nil {
			if err := applyStripeSessionState(db, payment, session); err != nil {
				return nil, err
			}
			updated, reloadErr := database.GetPaymentByID(db, payment.ID)
			if reloadErr == nil {
				payment = updated
			}
		}
	}

	if payment.ExpiresAt != nil && payment.Status != paymentStatusPaid && payment.ExpiresAt.Before(time.Now().UTC()) {
		if err := database.UpdatePaymentStatus(db, payment.ID, paymentStatusExpired); err != nil {
			return nil, err
		}
		payment.Status = paymentStatusExpired
	}

	return &payment, nil
}

func buildStripeInstructions(payment *models.Payment) string {
	if payment == nil {
		return ""
	}
	return "You will be redirected to Stripe Checkout in test mode. Use a Stripe test card, complete the payment, and you will return here automatically."
}

func mapStripePaymentStatus(sessionStatus, paymentStatus string) string {
	if paymentStatus == "paid" {
		return paymentStatusPaid
	}
	if sessionStatus == "expired" {
		return paymentStatusExpired
	}
	return paymentStatusPending
}

func compactStripeError(body []byte) string {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Error.Message) != "" {
		return payload.Error.Message
	}
	return strings.TrimSpace(string(bytes.TrimSpace(body)))
}

func stripeExpiresAtPointer(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	parsed := time.Unix(value, 0).UTC()
	return &parsed
}

func derefPaymentString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPointer(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func sanitizePhonePointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

type stripeCheckoutSessionParams struct {
	AmountKZT       int
	Currency        string
	SuccessURL      string
	CancelURL       string
	Reference       string
	ApplicationID   int
	StudentID       int
	StudentEmail    string
	ProductName     string
	ProductDesc     string
	ClientReference string
}
