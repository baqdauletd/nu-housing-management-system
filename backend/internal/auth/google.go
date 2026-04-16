package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

const googleCertsURL = "https://www.googleapis.com/oauth2/v1/certs"

var googleCertCache = struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}{}

func GoogleClientID() (string, error) {
	clientID := strings.TrimSpace(viper.GetString("GOOGLE_CLIENT_ID"))
	if clientID == "" {
		return "", errors.New("missing GOOGLE_CLIENT_ID configuration")
	}

	return clientID, nil
}

func GoogleAllowedDomain() string {
	domain := strings.TrimSpace(strings.ToLower(viper.GetString("GOOGLE_ALLOWED_DOMAIN")))
	if domain == "" {
		return "nu.edu.kz"
	}
	return domain
}

type GoogleTokenInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Audience      string `json:"aud"`
	HostedDomain  string `json:"hd"`
	Issuer        string `json:"iss"`
	ExpiresAt     int64  `json:"exp"`
}

func VerifyGoogleIDToken(idToken string) (GoogleTokenInfo, error) {
	clientID, err := GoogleClientID()
	if err != nil {
		return GoogleTokenInfo{}, err
	}

	if strings.TrimSpace(idToken) == "" {
		return GoogleTokenInfo{}, errors.New("missing Google ID token")
	}

	claims := googleIDTokenClaims{}
	token, err := jwt.ParseWithClaims(idToken, &claims, googleKeyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	if err != nil {
		return GoogleTokenInfo{}, fmt.Errorf("google token verification failed: %w", err)
	}

	if !token.Valid {
		return GoogleTokenInfo{}, errors.New("google token is invalid")
	}

	if !googleAudienceMatches(claims.Audience, clientID) {
		return GoogleTokenInfo{}, errors.New("google token audience mismatch")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return GoogleTokenInfo{}, errors.New("google token issuer mismatch")
	}

	if !claims.EmailVerified {
		return GoogleTokenInfo{}, errors.New("google account email is not verified")
	}

	allowedDomain := GoogleAllowedDomain()
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	hostedDomain := strings.ToLower(strings.TrimSpace(claims.HostedDomain))

	if !strings.HasSuffix(email, "@"+allowedDomain) || hostedDomain != allowedDomain {
		return GoogleTokenInfo{}, errors.New("google account is not in the allowed domain")
	}

	var expiresAt int64
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Unix()
	}

	return GoogleTokenInfo{
		Subject:       claims.Subject,
		Email:         email,
		VerifiedEmail: claims.EmailVerified,
		Audience:      clientID,
		HostedDomain:  hostedDomain,
		Issuer:        claims.Issuer,
		ExpiresAt:     expiresAt,
	}, nil
}

type googleIDTokenClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HostedDomain  string `json:"hd"`
	jwt.RegisteredClaims
}

func googleKeyFunc(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, errors.New("google token missing key id")
	}

	key, err := getGooglePublicKey(kid)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func getGooglePublicKey(kid string) (*rsa.PublicKey, error) {
	now := time.Now()

	googleCertCache.mu.RLock()
	if key := googleCertCache.keys[kid]; key != nil && now.Before(googleCertCache.expiresAt) {
		googleCertCache.mu.RUnlock()
		return key, nil
	}
	googleCertCache.mu.RUnlock()

	if err := refreshGoogleCerts(); err != nil {
		return nil, err
	}

	googleCertCache.mu.RLock()
	defer googleCertCache.mu.RUnlock()

	key := googleCertCache.keys[kid]
	if key == nil {
		return nil, fmt.Errorf("google signing key %q not found", kid)
	}

	return key, nil
}

func refreshGoogleCerts() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(googleCertsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch Google certs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch Google certs: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Google certs response: %w", err)
	}

	var certs map[string]string
	if err := json.Unmarshal(body, &certs); err != nil {
		return fmt.Errorf("failed to decode Google certs response: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(certs))
	for kid, certPEM := range certs {
		key, err := parseGoogleCertificatePEM(certPEM)
		if err != nil {
			return fmt.Errorf("failed to parse Google cert %q: %w", kid, err)
		}
		keys[kid] = key
	}

	googleCertCache.mu.Lock()
	googleCertCache.keys = keys
	googleCertCache.expiresAt = time.Now().Add(parseGoogleCertsMaxAge(resp.Header.Get("Cache-Control")))
	googleCertCache.mu.Unlock()

	return nil
}

func parseGoogleCertificatePEM(certPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("invalid PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate public key is not RSA")
	}

	return publicKey, nil
}

func parseGoogleCertsMaxAge(cacheControl string) time.Duration {
	const fallback = time.Hour

	for _, part := range strings.Split(cacheControl, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "max-age=") {
			continue
		}

		seconds, err := strconv.Atoi(strings.TrimPrefix(part, "max-age="))
		if err != nil || seconds <= 0 {
			return fallback
		}

		return time.Duration(seconds) * time.Second
	}

	return fallback
}

func googleAudienceMatches(audience jwt.ClaimStrings, expected string) bool {
	for _, aud := range audience {
		if aud == expected {
			return true
		}
	}

	return false
}
