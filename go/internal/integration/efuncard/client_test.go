package efuncard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gpt-team-api/internal/apperr"
)

func TestRedeemSetsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected bearer header, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"cardId":1,"cardNumber":"4111111111111111","expiryDate":"12/25","code":"CDK-1","status":"active"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	response, err := client.Redeem(context.Background(), "CDK-1")
	if err != nil {
		t.Fatalf("redeem failed: %v", err)
	}

	if !response.Success || response.Data.CardID != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestRedeemMapsConflictStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "already used", http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	if _, err := client.Redeem(context.Background(), "CDK-1"); err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestRedeemMapsConflictJSONMessage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"success":false,"error":"激活码已使用"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	_, err := client.Redeem(context.Background(), "CDK-1")
	if err == nil {
		t.Fatalf("expected conflict error")
	}

	if apperr.Message(err) != "激活码已使用" {
		t.Fatalf("expected parsed conflict message, got %q", apperr.Message(err))
	}
}

func TestQueryCardParsesExpiryMonthYear(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"cardId":24119,"cardNumber":"4462220001292405","expiryMonth":3,"expiryYear":2029,"cvv":"421","status":"CANCELLED"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	response, err := client.QueryCard(context.Background(), "UK-QUERY")
	if err != nil {
		t.Fatalf("query card failed: %v", err)
	}

	if response.Data.ExpiryMonth != 3 || response.Data.ExpiryYear != 2029 {
		t.Fatalf("unexpected expiry month/year: %+v", response.Data)
	}
}

func TestQueryCardParsesPreciseExpiryTime(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"cardId":24119,"cardNumber":"4462220001292405","expiryDate":"03/01","expiryMonth":3,"expiryYear":2029,"expiresAt":"2026-03-16T02:07:48Z","cvv":"421","status":"ACTIVE"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	response, err := client.QueryCard(context.Background(), "UK-QUERY")
	if err != nil {
		t.Fatalf("query card failed: %v", err)
	}

	if response.Data.ExpiresAt != "2026-03-16T02:07:48Z" {
		t.Fatalf("unexpected expiresAt: %+v", response.Data)
	}
}

func TestRedeemSanitizesHTMLBadGatewayError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html><head><title>502 Bad Gateway</title></head><body><center><h1>502 Bad Gateway</h1></center><hr><center>nginx</center></body></html>`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", server.Client())
	_, err := client.Redeem(context.Background(), "CDK-1")
	if err == nil {
		t.Fatalf("expected upstream error")
	}

	if apperr.Status(err) != http.StatusBadGateway {
		t.Fatalf("expected 502 status, got %d", apperr.Status(err))
	}
	if apperr.Message(err) != "上游服务暂时不可用，请稍后重试" {
		t.Fatalf("expected sanitized upstream message, got %q", apperr.Message(err))
	}
}
