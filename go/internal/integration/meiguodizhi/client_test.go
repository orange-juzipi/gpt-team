package meiguodizhi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchProfileExtractsFullNameAndBirthday(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Origin"); got != "https://www.meiguodizhi.com" {
			t.Fatalf("expected origin header, got %q", got)
		}
		if got := r.Header.Get("Accept-Language"); got != "zh-CN,zh;q=0.9" {
			t.Fatalf("expected accept-language header, got %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var requestBody map[string]string
		if err := json.Unmarshal(body, &requestBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		expected := map[string]string{
			"city":   "",
			"path":   "/",
			"method": "refresh",
		}
		if len(requestBody) != len(expected) {
			t.Fatalf("unexpected body size: %#v", requestBody)
		}
		for key, want := range expected {
			if got := requestBody[key]; got != want {
				t.Fatalf("unexpected body[%q]: got %q want %q", key, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"full_name":"Grace Hopper","birthday":"1906-12-09"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	result, err := client.FetchProfile(context.Background())
	if err != nil {
		t.Fatalf("fetch profile: %v", err)
	}

	if result.FullName != "Grace Hopper" || result.Birthday != "1906-12-09" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestFetchProfileReturnsContractErrorWhenFieldsMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"nickname":"Grace"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if _, err := client.FetchProfile(context.Background()); err == nil {
		t.Fatalf("expected contract error")
	}
}
