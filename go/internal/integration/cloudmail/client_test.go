package cloudmail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListInboxEmailsGeneratesTokenAndQueriesInbox(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/public/genToken":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST for genToken, got %s", r.Method)
			}

			var payload map[string]string
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode genToken body: %v", err)
			}

			if payload["email"] != "admin@mail.example" {
				t.Fatalf("unexpected email: %s", payload["email"])
			}
			if payload["password"] != "secret" {
				t.Fatalf("unexpected password: %s", payload["password"])
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":200,"data":{"token":"jwt-token"}}`))
		case "/api/public/emailList":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST for emailList, got %s", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "jwt-token" {
				t.Fatalf("unexpected auth header: %s", got)
			}

			var payload map[string]any
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode emailList body: %v", err)
			}

			if payload["toEmail"] != "target@mail.example" {
				t.Fatalf("unexpected toEmail: %v", payload["toEmail"])
			}
			if payload["num"] != float64(1) {
				t.Fatalf("unexpected num: %v", payload["num"])
			}
			if payload["timeSort"] != "desc" {
				t.Fatalf("unexpected timeSort: %v", payload["timeSort"])
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"code": 200,
				"data": {
					"records": [
						{
							"id": 88,
							"toEmail": "target@mail.example",
							"sendEmail": "noreply@example.com",
							"sendName": "Example Sender",
							"subject": "Welcome",
							"text": "Hello from Cloudmail",
							"createTime": "2026-03-15 10:20:30"
						}
					]
				}
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "", server.Client())
	items, err := client.ListInboxEmails(context.Background(), "admin@mail.example", "secret", "target@mail.example")
	if err != nil {
		t.Fatalf("ListInboxEmails: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.ID != "88" {
		t.Fatalf("expected id 88, got %s", item.ID)
	}
	if item.Account != "target@mail.example" {
		t.Fatalf("unexpected account: %s", item.Account)
	}
	if item.From != "noreply@example.com" {
		t.Fatalf("unexpected from: %s", item.From)
	}
	if item.Subject != "Welcome" {
		t.Fatalf("unexpected subject: %s", item.Subject)
	}
	if item.Preview != "Hello from Cloudmail" {
		t.Fatalf("unexpected preview: %s", item.Preview)
	}
}

func TestListInboxEmailsRequiresCredentials(t *testing.T) {
	t.Parallel()

	client := NewClient("https://mail.example", "", http.DefaultClient)
	if _, err := client.ListInboxEmails(context.Background(), "", "", "target@mail.example"); err == nil {
		t.Fatalf("expected missing credentials error")
	}
}
