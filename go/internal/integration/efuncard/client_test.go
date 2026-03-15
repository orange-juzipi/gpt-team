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
