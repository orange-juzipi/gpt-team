package cloudmail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/integration/mailbox"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

type genTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type genTokenEnvelope struct {
	Code    *int   `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type emailListRequest struct {
	Num      int    `json:"num"`
	Size     int    `json:"size"`
	Type     int    `json:"type"`
	ToEmail  string `json:"toEmail"`
	TimeSort string `json:"timeSort"`
}

type emailListEnvelope struct {
	Code    *int            `json:"code"`
	Message string          `json:"message"`
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func NewClient(baseURL, apiToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiToken:   strings.TrimSpace(apiToken),
		httpClient: httpClient,
	}
}

func (c *Client) ListInboxEmails(ctx context.Context, authEmail, authPassword, targetEmail string) ([]mailbox.Email, error) {
	authEmail = strings.TrimSpace(authEmail)
	authPassword = strings.TrimSpace(authPassword)
	targetEmail = strings.TrimSpace(targetEmail)

	if authEmail == "" {
		return nil, apperr.BadRequest("cloudmail_email_required", "Cloudmail email is required")
	}
	if authPassword == "" {
		return nil, apperr.BadRequest("cloudmail_password_required", "Cloudmail password is required")
	}
	if targetEmail == "" {
		return nil, apperr.BadRequest("cloudmail_target_email_required", "target email is required")
	}

	token, err := c.generateToken(ctx, authEmail, authPassword)
	if err != nil {
		return nil, err
	}

	return c.queryInboxEmails(ctx, targetEmail, token)
}

func (c *Client) queryInboxEmails(ctx context.Context, targetEmail, token string) ([]mailbox.Email, error) {
	requestBody, err := json.Marshal(emailListRequest{
		Num:      1,
		Size:     50,
		Type:     0,
		ToEmail:  targetEmail,
		TimeSort: "desc",
	})
	if err != nil {
		return nil, apperr.Internal("cloudmail_payload_encode_failed", "failed to encode Cloudmail email query", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/public/emailList",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, apperr.Internal("cloudmail_request_build_failed", "failed to build Cloudmail request", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, apperr.Upstream("cloudmail_request_failed", "failed to call Cloudmail", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, apperr.Upstream("cloudmail_read_failed", "failed to read Cloudmail response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, mapHTTPError(statusCodeFromResponse(response.StatusCode), body)
	}

	var envelope emailListEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, apperr.Upstream("cloudmail_decode_failed", "failed to decode Cloudmail response", err)
	}

	if envelope.Code != nil && *envelope.Code != 0 && *envelope.Code != http.StatusOK {
		message := strings.TrimSpace(firstNonEmpty(envelope.Message, envelope.Msg))
		if message == "" {
			message = "Cloudmail returned an error"
		}
		return nil, apperr.Upstream("cloudmail_response_failed", message, nil)
	}

	items, err := decodeEmailItems(envelope.Data)
	if err != nil {
		return nil, err
	}

	emails := make([]mailbox.Email, 0, len(items))
	for _, item := range items {
		emails = append(emails, normalizeEmail(item, targetEmail))
	}

	return emails, nil
}

func (c *Client) generateToken(ctx context.Context, email, password string) (string, error) {
	if email == "" {
		return "", apperr.BadRequest("cloudmail_email_required", "Cloudmail email is required")
	}
	if password == "" {
		return "", apperr.BadRequest("cloudmail_password_required", "Cloudmail password is required")
	}

	requestBody, err := json.Marshal(genTokenRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", apperr.Internal("cloudmail_payload_encode_failed", "failed to encode Cloudmail token request", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/public/genToken",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return "", apperr.Internal("cloudmail_request_build_failed", "failed to build Cloudmail request", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", apperr.Upstream("cloudmail_request_failed", "failed to call Cloudmail", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", apperr.Upstream("cloudmail_read_failed", "failed to read Cloudmail response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return "", mapHTTPError(statusCodeFromResponse(response.StatusCode), body)
	}

	var envelope genTokenEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", apperr.Upstream("cloudmail_decode_failed", "failed to decode Cloudmail response", err)
	}

	if envelope.Code != nil && *envelope.Code != 0 && *envelope.Code != http.StatusOK {
		message := strings.TrimSpace(firstNonEmpty(envelope.Message, envelope.Msg))
		if message == "" {
			message = "Cloudmail returned an error"
		}
		return "", apperr.Upstream("cloudmail_token_failed", message, nil)
	}

	token := strings.TrimSpace(envelope.Data.Token)
	if token == "" {
		return "", apperr.Upstream("cloudmail_token_missing", "Cloudmail token was empty", nil)
	}

	return token, nil
}

func decodeEmailItems(raw json.RawMessage) ([]map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var directItems []map[string]any
	if err := json.Unmarshal(raw, &directItems); err == nil {
		return directItems, nil
	}

	var container map[string]any
	if err := json.Unmarshal(raw, &container); err != nil {
		return nil, apperr.Upstream("cloudmail_decode_failed", "failed to decode Cloudmail email list", err)
	}

	for _, key := range []string{"list", "rows", "records", "items", "data"} {
		value, exists := container[key]
		if !exists {
			continue
		}

		items, err := remarshalItems(value)
		if err != nil {
			continue
		}
		return items, nil
	}

	return nil, nil
}

func remarshalItems(value any) ([]map[string]any, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, apperr.Upstream("cloudmail_decode_failed", "failed to decode Cloudmail email list", err)
	}

	var items []map[string]any
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, apperr.Upstream("cloudmail_decode_failed", "failed to decode Cloudmail email list", err)
	}

	return items, nil
}

func normalizeEmail(item map[string]any, requestedAccount string) mailbox.Email {
	subject := asString(firstValue(item, "subject", "title"))
	text := asString(firstValue(item, "text", "contentText", "plainText", "summary"))
	html := asString(firstValue(item, "content", "html", "body"))

	return mailbox.Email{
		ID:         asString(firstValue(item, "emailId", "mailId", "id")),
		Account:    firstNonEmpty(asString(firstValue(item, "toEmail", "name", "email", "mailbox", "address")), strings.TrimSpace(requestedAccount)),
		From:       asString(firstValue(item, "sendEmail", "from", "fromEmail")),
		FromName:   asString(firstValue(item, "sendName", "fromName")),
		Subject:    subject,
		Preview:    summarizeText(firstNonEmpty(text, html, subject)),
		ReceivedAt: asString(firstValue(item, "createTime", "receivedAt", "time", "sendTime")),
	}
}

func firstValue(item map[string]any, keys ...string) any {
	for _, key := range keys {
		value, exists := item[key]
		if exists {
			return value
		}
	}
	return nil
}

func summarizeText(value string) string {
	cleaned := htmlTagPattern.ReplaceAllString(value, " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	if len(cleaned) > 180 {
		return cleaned[:180] + "..."
	}
	return cleaned
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%.0f", typed)
		}
		return fmt.Sprintf("%v", typed)
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	default:
		return ""
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

func mapHTTPError(status int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(status)
	}

	code := fmt.Sprintf("cloudmail_http_%d", status)
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

func statusCodeFromResponse(status int) int {
	if status == 0 {
		return http.StatusBadGateway
	}
	return status
}
