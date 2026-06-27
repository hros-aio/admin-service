package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hros/admin-service/internal/application/interfaces"
)

// SSOProviderConfig holds configuration details for a single Identity Provider.
type SSOProviderConfig struct {
	ClientID    string
	RedirectURL string
	AuthURL     string
	Scopes      []string
}

// InitiateSSOInput represents the input for initiating the SSO flow.
type InitiateSSOInput struct {
	Provider string
}

// InitiateSSOOutput represents the output of initiating the SSO flow.
type InitiateSSOOutput struct {
	RedirectURL string
}

// InitiateSSOUseCase orchestrates the workflow for starting the SSO flow.
type InitiateSSOUseCase struct {
	stateCache interfaces.SSOStateCache
	providers  map[string]SSOProviderConfig
}

// NewInitiateSSOUseCase creates a new InitiateSSOUseCase.
func NewInitiateSSOUseCase(
	stateCache interfaces.SSOStateCache,
	providers map[string]SSOProviderConfig,
) *InitiateSSOUseCase {
	return &InitiateSSOUseCase{
		stateCache: stateCache,
		providers:  providers,
	}
}

// Execute handles generating state variables, caching them, and constructing the redirect URL.
func (uc *InitiateSSOUseCase) Execute(ctx context.Context, input InitiateSSOInput) (InitiateSSOOutput, error) {
	if input.Provider == "" {
		return InitiateSSOOutput{}, errors.New("provider cannot be empty")
	}

	providerConfig, exists := uc.providers[strings.ToLower(input.Provider)]
	if !exists {
		return InitiateSSOOutput{}, fmt.Errorf("unsupported or unconfigured provider: %s", input.Provider)
	}

	// Validate config parameters
	if providerConfig.ClientID == "" || providerConfig.RedirectURL == "" || providerConfig.AuthURL == "" {
		return InitiateSSOOutput{}, fmt.Errorf("invalid configuration for provider %s: missing ClientID, RedirectURL, or AuthURL", input.Provider)
	}

	// Generate state (32 bytes of entropy)
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return InitiateSSOOutput{}, fmt.Errorf("generate secure state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Generate nonce (32 bytes of entropy)
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return InitiateSSOOutput{}, fmt.Errorf("generate secure nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Store state and nonce in cache with a 10-minute TTL
	const stateTTL = 10 * time.Minute
	if err := uc.stateCache.StoreState(ctx, state, nonce, stateTTL); err != nil {
		return InitiateSSOOutput{}, fmt.Errorf("store sso state: %w", err)
	}

	// Construct redirect URL
	u, err := url.Parse(providerConfig.AuthURL)
	if err != nil {
		return InitiateSSOOutput{}, fmt.Errorf("parse provider auth url: %w", err)
	}

	q := u.Query()
	q.Set("client_id", providerConfig.ClientID)
	q.Set("redirect_uri", providerConfig.RedirectURL)
	q.Set("response_type", "code")
	if len(providerConfig.Scopes) > 0 {
		q.Set("scope", strings.Join(providerConfig.Scopes, " "))
	} else {
		q.Set("scope", "openid email profile")
	}
	q.Set("state", state)
	q.Set("nonce", nonce)

	u.RawQuery = q.Encode()

	return InitiateSSOOutput{
		RedirectURL: u.String(),
	}, nil
}
