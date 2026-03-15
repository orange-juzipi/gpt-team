package duckmail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListInboxEmailsGeneratesTokenAndQueriesMessages(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST /token, got %s", r.Method)
			}

			var payload map[string]string
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode token body: %v", err)
			}

			if payload["address"] != "box@duckmail.sbs" {
				t.Fatalf("unexpected address: %s", payload["address"])
			}
			if payload["password"] != "secret" {
				t.Fatalf("unexpected password: %s", payload["password"])
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"acc-1","token":"duck-token"}`))
		case "/messages":
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET /messages, got %s", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer duck-token" {
				t.Fatalf("unexpected authorization: %s", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"hydra:member": [
					{
						"id": "msg-1",
						"from": {"name": "Sender", "address": "sender@example.com"},
						"to": [{"name": "You", "address": "box@duckmail.sbs"}],
						"subject": "Duck Subject",
						"createdAt": "2026-03-15T10:20:30Z"
					}
				]
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	items, err := client.ListInboxEmails(context.Background(), "box@duckmail.sbs", "secret", "box@duckmail.sbs")
	if err != nil {
		t.Fatalf("ListInboxEmails: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.ID != "msg-1" {
		t.Fatalf("unexpected id: %s", item.ID)
	}
	if item.Account != "box@duckmail.sbs" {
		t.Fatalf("unexpected account: %s", item.Account)
	}
	if item.From != "sender@example.com" {
		t.Fatalf("unexpected from: %s", item.From)
	}
	if item.Subject != "Duck Subject" {
		t.Fatalf("unexpected subject: %s", item.Subject)
	}
}

func TestCreateAccountUsesAPIKeyWhenProvided(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST /accounts, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer dk_secret" {
			t.Fatalf("unexpected authorization: %s", got)
		}

		var payload map[string]string
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode create account body: %v", err)
		}

		if payload["address"] != "box@duckmail.sbs" {
			t.Fatalf("unexpected address: %s", payload["address"])
		}
		if payload["password"] != "secret123" {
			t.Fatalf("unexpected password: %s", payload["password"])
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"acc-1","address":"box@duckmail.sbs"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	if err := client.CreateAccount(context.Background(), "dk_secret", "box@duckmail.sbs", "secret123"); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
}

func TestListInboxEmailsRequiresCredentials(t *testing.T) {
	t.Parallel()

	client := NewClient("https://api.duckmail.sbs", http.DefaultClient)
	if _, err := client.ListInboxEmails(context.Background(), "", "", "box@duckmail.sbs"); err == nil {
		t.Fatalf("expected missing credentials error")
	}
}
