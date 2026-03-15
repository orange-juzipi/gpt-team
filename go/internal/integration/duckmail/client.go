package duckmail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/mailbox"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type createAccountRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

type tokenRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

type tokenResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type messageListEnvelope struct {
	Items []messageItem `json:"hydra:member"`
}

type messageItem struct {
	ID        string        `json:"id"`
	From      messageSender `json:"from"`
	To        []messageTo   `json:"to"`
	Subject   string        `json:"subject"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
}

type messageSender struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type messageTo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type messageDetail struct {
	Text string   `json:"text"`
	HTML []string `json:"html"`
}

type errorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) CreateAccount(ctx context.Context, apiKey, address, password string) error {
	address = strings.TrimSpace(address)
	password = strings.TrimSpace(password)
	apiKey = strings.TrimSpace(apiKey)

	if address == "" {
		return apperr.BadRequest("duckmail_address_required", "DuckMail address is required")
	}
	if password == "" {
		return apperr.BadRequest("duckmail_password_required", "DuckMail password is required")
	}

	body, err := json.Marshal(createAccountRequest{
		Address:  address,
		Password: password,
	})
	if err != nil {
		return apperr.Internal("duckmail_payload_encode_failed", "failed to encode DuckMail account request", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/accounts",
		bytes.NewReader(body),
	)
	if err != nil {
		return apperr.Internal("duckmail_request_build_failed", "failed to build DuckMail request", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return apperr.Upstream("duckmail_request_failed", "failed to call DuckMail", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return apperr.Upstream("duckmail_read_failed", "failed to read DuckMail response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return mapHTTPError(response.StatusCode, responseBody)
	}

	return nil
}

func (c *Client) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	authEmail = strings.TrimSpace(authEmail)
	authPassword = strings.TrimSpace(authPassword)
	targetEmail = strings.TrimSpace(targetEmail)

	if authEmail == "" {
		return nil, apperr.BadRequest("duckmail_email_required", "DuckMail email is required")
	}
	if authPassword == "" {
		return nil, apperr.BadRequest("duckmail_password_required", "DuckMail password is required")
	}
	if targetEmail == "" {
		return nil, apperr.BadRequest("duckmail_target_email_required", "target email is required")
	}

	token, err := c.generateToken(ctx, authEmail, authPassword)
	if err != nil {
		return nil, err
	}

	return c.queryMessages(ctx, token, targetEmail)
}

func (c *Client) generateToken(ctx context.Context, address, password string) (string, error) {
	body, err := json.Marshal(tokenRequest{
		Address:  address,
		Password: password,
	})
	if err != nil {
		return "", apperr.Internal("duckmail_payload_encode_failed", "failed to encode DuckMail token request", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/token",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", apperr.Internal("duckmail_request_build_failed", "failed to build DuckMail request", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", apperr.Upstream("duckmail_request_failed", "failed to call DuckMail", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", apperr.Upstream("duckmail_read_failed", "failed to read DuckMail response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return "", mapHTTPError(response.StatusCode, responseBody)
	}

	var payload tokenResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return "", apperr.Upstream("duckmail_decode_failed", "failed to decode DuckMail token response", err)
	}

	token := strings.TrimSpace(payload.Token)
	if token == "" {
		return "", apperr.Upstream("duckmail_token_missing", "DuckMail token was empty", nil)
	}

	return token, nil
}

func (c *Client) queryMessages(ctx context.Context, token, targetEmail string) ([]mailbox.Email, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/messages", http.NoBody)
	if err != nil {
		return nil, apperr.Internal("duckmail_request_build_failed", "failed to build DuckMail request", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, apperr.Upstream("duckmail_request_failed", "failed to call DuckMail", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, apperr.Upstream("duckmail_read_failed", "failed to read DuckMail response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, mapHTTPError(response.StatusCode, responseBody)
	}

	var payload messageListEnvelope
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, apperr.Upstream("duckmail_decode_failed", "failed to decode DuckMail message list", err)
	}

	items := make([]mailbox.Email, 0, len(payload.Items))
	for _, item := range payload.Items {
		items = append(items, mailbox.Email{
			ID:         strings.TrimSpace(item.ID),
			Account:    resolveAccount(item.To, targetEmail),
			From:       strings.TrimSpace(item.From.Address),
			FromName:   strings.TrimSpace(item.From.Name),
			Subject:    strings.TrimSpace(item.Subject),
			Preview:    strings.TrimSpace(item.Subject),
			ReceivedAt: firstNonEmpty(strings.TrimSpace(item.CreatedAt), strings.TrimSpace(item.UpdatedAt)),
		})
	}

	return items, nil
}

func resolveAccount(values []messageTo, fallback string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value.Address); trimmed != "" {
			return trimmed
		}
	}

	return strings.TrimSpace(fallback)
}

func mapHTTPError(status int, body []byte) error {
	message := strings.TrimSpace(string(body))

	var payload errorEnvelope
	if err := json.Unmarshal(body, &payload); err == nil {
		message = firstNonEmpty(payload.Message, payload.Error, message)
	}

	if message == "" {
		message = http.StatusText(status)
	}

	code := fmt.Sprintf("duckmail_http_%d", status)
	switch status {
	case http.StatusNotFound:
		return apperr.NotFound(code, message)
	case http.StatusConflict:
		return apperr.Conflict(code, message)
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return apperr.BadRequest(code, message)
	case http.StatusUnauthorized, http.StatusForbidden:
		return apperr.New(status, code, message)
	default:
		return apperr.Upstream(code, message, nil)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
