package meiguodizhi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gpt-team-api/internal/apperr"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

type ProfileResponse struct {
	FullName string `json:"fullName"`
	Birthday string `json:"birthday"`
	Raw      string `json:"raw"`
}

type refreshRequest struct {
	City   string `json:"city"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

func (c *Client) FetchProfile(ctx context.Context) (ProfileResponse, error) {
	requestBody, err := json.Marshal(refreshRequest{
		City:   "",
		Path:   "/",
		Method: "refresh",
	})
	if err != nil {
		return ProfileResponse{}, apperr.Internal("meiguodizhi_payload_encode_failed", "failed to encode meiguodizhi request", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return ProfileResponse{}, apperr.Internal("meiguodizhi_request_build_failed", "failed to build meiguodizhi request", err)
	}
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://www.meiguodizhi.com")
	request.Header.Set("Referer", "https://www.meiguodizhi.com/")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return ProfileResponse{}, apperr.Upstream("meiguodizhi_request_failed", "failed to call meiguodizhi", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ProfileResponse{}, apperr.Upstream("meiguodizhi_read_failed", "failed to read meiguodizhi response", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return ProfileResponse{}, apperr.New(response.StatusCode, "meiguodizhi_http_error", strings.TrimSpace(string(body)))
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProfileResponse{}, apperr.Upstream("meiguodizhi_decode_failed", "failed to decode meiguodizhi response", err)
	}

	fullName, birthday := extractProfile(payload)
	if fullName == "" || birthday == "" {
		return ProfileResponse{}, apperr.Upstream("meiguodizhi_contract_changed", "meiguodizhi response does not contain full name and birthday", nil)
	}

	return ProfileResponse{
		FullName: fullName,
		Birthday: birthday,
		Raw:      string(body),
	}, nil
}

func extractProfile(payload any) (string, string) {
	var fullName string
	var birthday string
	var fallbackName string

	var walk func(value any)
	walk = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				normalized := normalizeKey(key)
				if fullName == "" && isFullNameKey(normalized) {
					if candidate := extractString(child); candidate != "" {
						fullName = candidate
					}
				}

				if fallbackName == "" && normalized == "name" {
					if candidate := extractString(child); strings.Contains(candidate, " ") {
						fallbackName = candidate
					}
				}

				if birthday == "" && isBirthdayKey(normalized) {
					if candidate := extractString(child); candidate != "" {
						birthday = candidate
					}
				}

				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}

	walk(payload)
	if fullName == "" {
		fullName = fallbackName
	}

	return fullName, birthday
}

func normalizeKey(key string) string {
	replacer := strings.NewReplacer("_", "", "-", "", " ", "")
	return strings.ToLower(replacer.Replace(key))
}

func isFullNameKey(key string) bool {
	switch key {
	case "fullname", "realname", "fullnamecn", "全名", "姓名":
		return true
	default:
		return false
	}
}

func isBirthdayKey(key string) bool {
	switch key {
	case "birthday", "birth", "birthdate", "dateofbirth", "dob", "生日":
		return true
	default:
		return false
	}
}

func extractString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", typed))
	default:
		return ""
	}
}
