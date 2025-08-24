package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthClient handles authentication API calls
type AuthClient struct {
	BaseURL string
	Client  *http.Client
}

// NewAuthClient creates a new authentication client
func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
	SessionID int `json:"session_id"`
}

// UserProfile represents user profile data
type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastLogin *time.Time `json:"last_login"`
}

// Register creates a new user account
func (c *AuthClient) Register(req RegisterRequest) (*AuthResponse, error) {
	return c.authRequest("POST", "/v1/auth/register", req, nil)
}

// Login authenticates a user and returns JWT token
func (c *AuthClient) Login(req LoginRequest) (*AuthResponse, error) {
	return c.authRequest("POST", "/v1/auth/login", req, nil)
}

// Logout invalidates the current session
func (c *AuthClient) Logout(token string) error {
	_, err := c.authRequest("POST", "/v1/auth/logout", nil, &token)
	return err
}

// GetProfile retrieves current user profile
func (c *AuthClient) GetProfile(token string) (*UserProfile, error) {
	body, err := c.makeRequest("GET", "/v1/auth/profile", nil, &token)
	if err != nil {
		return nil, err
	}

	var profile UserProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &profile, nil
}

// authRequest makes an authentication request and returns AuthResponse
func (c *AuthClient) authRequest(method, endpoint string, payload interface{}, token *string) (*AuthResponse, error) {
	body, err := c.makeRequest(method, endpoint, payload, token)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	return &authResp, nil
}

// makeRequest makes an HTTP request to the API
func (c *AuthClient) makeRequest(method, endpoint string, payload interface{}, token *string) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil && *token != "" {
		req.Header.Set("Authorization", "Bearer "+*token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}