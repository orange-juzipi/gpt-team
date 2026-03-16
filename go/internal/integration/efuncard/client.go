package efuncard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gpt-team-api/internal/apperr"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type RedeemData struct {
	CardID            uint64  `json:"cardId"`
	CardNumber        string  `json:"cardNumber"`
	ExpiryDate        string  `json:"expiryDate"`
	ExpiryMonth       int     `json:"expiryMonth"`
	ExpiryYear        int     `json:"expiryYear"`
	CVV               string  `json:"cvv"`
	Code              string  `json:"code"`
	Status            string  `json:"status"`
	Balance           float64 `json:"balance"`
	CreatedAt         string  `json:"createdAt"`
	ExpiresAt         string  `json:"expiresAt"`
	ExpireAt          string  `json:"expireAt"`
	ExpiredAt         string  `json:"expiredAt"`
	ValidUntil        string  `json:"validUntil"`
	EndAt             string  `json:"endAt"`
	ExpiryTime        string  `json:"expiryTime"`
	BillingAddress    string  `json:"billingAddress"`
	NodeInstructions  string  `json:"nodeInstructions"`
	GroupInstructions string  `json:"groupInstructions"`
}

type QueryData struct {
	CardID            uint64  `json:"cardId"`
	CardNumber        string  `json:"cardNumber"`
	ExpiryDate        string  `json:"expiryDate"`
	ExpiryMonth       int     `json:"expiryMonth"`
	ExpiryYear        int     `json:"expiryYear"`
	CVV               string  `json:"cvv"`
	Code              string  `json:"code"`
	Status            string  `json:"status"`
	Balance           float64 `json:"balance"`
	CreatedAt         string  `json:"createdAt"`
	ExpiresAt         string  `json:"expiresAt"`
	ExpireAt          string  `json:"expireAt"`
	ExpiredAt         string  `json:"expiredAt"`
	ValidUntil        string  `json:"validUntil"`
	EndAt             string  `json:"endAt"`
	ExpiryTime        string  `json:"expiryTime"`
	BillingAddress    string  `json:"billingAddress"`
	NodeInstructions  string  `json:"nodeInstructions"`
	GroupInstructions string  `json:"groupInstructions"`
}

type BillingTransaction struct {
	ID              string  `json:"id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Merchant        string  `json:"merchant"`
	MerchantName    string  `json:"merchantName"`
	MerchantCity    string  `json:"merchantCity"`
	MerchantCountry string  `json:"merchantCountry"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"createdAt"`
	Date            string  `json:"date"`
}

type BillingData struct {
	CardID            uint64               `json:"cardId"`
	Code              string               `json:"code"`
	LastFour          string               `json:"lastFour"`
	Total             int                  `json:"total"`
	Transactions      []BillingTransaction `json:"transactions"`
	TotalSpent        float64              `json:"totalSpent"`
	RemainingBalance  float64              `json:"remainingBalance"`
	SettledCount      int                  `json:"settledCount"`
	SettledAmount     float64              `json:"settledAmount"`
	BillingAddress    string               `json:"billingAddress"`
	NodeInstructions  string               `json:"nodeInstructions"`
	GroupInstructions string               `json:"groupInstructions"`
}

type ThreeDSVerification struct {
	OTP        string `json:"otp"`
	Merchant   string `json:"merchant"`
	Amount     string `json:"amount"`
	ReceivedAt string `json:"receivedAt"`
}

type ThreeDSData struct {
	CardID        uint64                `json:"cardId"`
	Code          string                `json:"code"`
	Verifications []ThreeDSVerification `json:"verifications"`
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *Client) Redeem(ctx context.Context, code string) (APIResponse[RedeemData], error) {
	request, err := c.newJSONRequest(ctx, http.MethodPost, "/api/external/redeem", map[string]any{
		"code": code,
	})
	if err != nil {
		return APIResponse[RedeemData]{}, err
	}

	return doRequest[APIResponse[RedeemData]](c.httpClient, request)
}

func (c *Client) QueryCard(ctx context.Context, code string) (APIResponse[QueryData], error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/api/external/cards/query/"+urlEscape(code), nil)
	if err != nil {
		return APIResponse[QueryData]{}, err
	}

	return doRequest[APIResponse[QueryData]](c.httpClient, request)
}

func (c *Client) Billing(ctx context.Context, code string) (APIResponse[BillingData], error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/api/external/billing/"+urlEscape(code), nil)
	if err != nil {
		return APIResponse[BillingData]{}, err
	}

	return doRequest[APIResponse[BillingData]](c.httpClient, request)
}

func (c *Client) ThreeDS(ctx context.Context, code string, minutes int) (APIResponse[ThreeDSData], error) {
	request, err := c.newJSONRequest(ctx, http.MethodPost, "/api/external/3ds/verify", map[string]any{
		"code":    code,
		"minutes": minutes,
	})
	if err != nil {
		return APIResponse[ThreeDSData]{}, err
	}

	return doRequest[APIResponse[ThreeDSData]](c.httpClient, request)
}

func (c *Client) newJSONRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, apperr.Internal("efuncard_payload_encode_failed", "failed to encode efuncard payload", err)
	}

	request, err := c.newRequest(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if c.apiKey == "" {
		return nil, apperr.Internal("efuncard_api_key_missing", "EFuncard API key is not configured", nil)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, apperr.Internal("efuncard_request_build_failed", "failed to build efuncard request", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Accept", "application/json")

	return request, nil
}

func doRequest[T any](httpClient *http.Client, request *http.Request) (T, error) {
	var zero T

	response, err := httpClient.Do(request)
	if err != nil {
		return zero, apperr.Upstream("efuncard_request_failed", "failed to call efuncard", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return zero, apperr.Upstream("efuncard_read_failed", "failed to read efuncard response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return zero, mapHTTPError(response.StatusCode, body)
	}

	var payload T
	if err := json.Unmarshal(body, &payload); err != nil {
		return zero, apperr.Upstream("efuncard_decode_failed", "failed to decode efuncard response", err)
	}

	return payload, nil
}

func mapHTTPError(status int, body []byte) error {
	message := sanitizeHTTPErrorMessage(status, extractErrorMessage(body))
	if message == "" {
		message = http.StatusText(status)
	}

	code := fmt.Sprintf("efuncard_http_%d", status)
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

func sanitizeHTTPErrorMessage(status int, message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return ""
	}

	if looksLikeHTMLResponse(trimmed) || looksLikeGatewayError(trimmed) {
		if status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout {
			return "上游服务暂时不可用，请稍后重试"
		}
		if status >= http.StatusInternalServerError {
			return "上游服务响应异常，请稍后重试"
		}
	}

	return trimmed
}

func looksLikeHTMLResponse(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	return strings.Contains(lower, "<html") ||
		strings.Contains(lower, "<!doctype html") ||
		strings.Contains(lower, "<body") ||
		strings.Contains(lower, "<head") ||
		strings.Contains(lower, "</html>")
}

func looksLikeGatewayError(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	return strings.Contains(lower, "bad gateway") ||
		strings.Contains(lower, "gateway timeout") ||
		strings.Contains(lower, "service unavailable") ||
		strings.Contains(lower, "nginx")
}

func extractErrorMessage(body []byte) string {
	message := strings.TrimSpace(string(body))
	if message == "" {
		return ""
	}

	var payload struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return message
	}

	switch {
	case strings.TrimSpace(payload.Error) != "":
		return strings.TrimSpace(payload.Error)
	case strings.TrimSpace(payload.Message) != "":
		return strings.TrimSpace(payload.Message)
	case strings.TrimSpace(payload.Msg) != "":
		return strings.TrimSpace(payload.Msg)
	default:
		return message
	}
}

func urlEscape(value string) string {
	return url.PathEscape(value)
}
