package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// User represents an authenticated user
type User struct {
	AccessToken string
	Email       string
}

func SignInWithEmail(email, password string) (*User, error) {
	if publicClient == nil {
		return nil, fmt.Errorf("supabase client not initialized")
	}

	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", strings.TrimRight(config.URL, "/"))

	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", config.AnonKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: %s", string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		User        struct {
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &User{
		AccessToken: result.AccessToken,
		Email:       result.User.Email,
	}, nil
}

func ValidateSession(sessionToken string) (*User, error) {
	if publicClient == nil {
		return nil, fmt.Errorf("supabase client not initialized")
	}

	url := fmt.Sprintf("%s/auth/v1/user", strings.TrimRight(config.URL, "/"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", config.AnonKey)
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid session")
	}

	var result struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &User{
		AccessToken: sessionToken,
		Email:       result.Email,
	}, nil
}
