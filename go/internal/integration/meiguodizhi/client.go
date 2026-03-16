package meiguodizhi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gpt-team-api/internal/apperr"
	"gpt-team-api/internal/model"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

type ProfileResponse struct {
	FullName      string `json:"fullName"`
	Birthday      string `json:"birthday"`
	StreetAddress string `json:"streetAddress"`
	District      string `json:"district"`
	City          string `json:"city"`
	State         string `json:"state"`
	StateFull     string `json:"stateFull"`
	ZipCode       string `json:"zipCode"`
	PhoneNumber   string `json:"phoneNumber"`
	Raw           string `json:"raw"`
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

func (c *Client) FetchProfile(ctx context.Context, cardType model.CardType) (ProfileResponse, error) {
	requestPath := profilePathForCardType(cardType)
	requestBody, err := json.Marshal(refreshRequest{
		City:   "",
		Path:   requestPath,
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
	request.Header.Set("Referer", buildReferer(requestPath))

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

	profile := extractProfile(payload)
	if profile.FullName == "" || profile.Birthday == "" {
		return ProfileResponse{}, apperr.Upstream("meiguodizhi_contract_changed", "meiguodizhi response does not contain full name and birthday", nil)
	}

	profile.Raw = string(body)
	return profile, nil
}

func extractProfile(payload any) ProfileResponse {
	state := findFirstString(payload, "state", "province", "region")

	return ProfileResponse{
		FullName:      firstNonEmpty(findFirstString(payload, "full_name", "fullName", "realname", "name", "姓名")),
		Birthday:      firstNonEmpty(findFirstString(payload, "birthday", "birth", "birthdate", "date_of_birth", "dob", "生日")),
		StreetAddress: firstNonEmpty(findFirstString(payload, "address", "street", "streetaddress", "address1", "address_line_1")),
		District:      firstNonEmpty(findFirstString(payload, "district", "districts", "county", "borough", "suburb", "address_alias")),
		City:          firstNonEmpty(findFirstString(payload, "city", "town")),
		State:         state,
		StateFull:     firstNonEmpty(findFirstString(payload, "state_full", "province_full", "region_full"), state),
		ZipCode:       firstNonEmpty(findFirstString(payload, "zip_code", "zip", "zipcode", "postal_code", "postalcode", "postcode")),
		PhoneNumber:   firstNonEmpty(findFirstString(payload, "telephone", "phone", "phone_number", "mobile", "tel")),
	}
}

func findFirstString(value any, keys ...string) string {
	expected := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		expected[normalizeKey(key)] = struct{}{}
	}

	var result string
	var walk func(any)
	walk = func(current any) {
		if result != "" {
			return
		}

		switch typed := current.(type) {
		case map[string]any:
			for key, child := range typed {
				if _, ok := expected[normalizeKey(key)]; ok {
					if candidate := extractString(child); candidate != "" {
						result = candidate
						return
					}
				}

				walk(child)
				if result != "" {
					return
				}
			}
		case []any:
			for _, child := range typed {
				walk(child)
				if result != "" {
					return
				}
			}
		}
	}

	walk(value)
	return result
}

func extractString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func normalizeKey(key string) string {
	replacer := strings.NewReplacer("_", "", "-", "", " ", "")
	return strings.ToLower(replacer.Replace(key))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func profilePathForCardType(cardType model.CardType) string {
	switch cardType {
	case model.CardTypeUK:
		return "/uk-address"
	case model.CardTypeES:
		return "/es-address"
	default:
		return "/"
	}
}

func buildReferer(path string) string {
	return "https://www.meiguodizhi.com" + path
}
