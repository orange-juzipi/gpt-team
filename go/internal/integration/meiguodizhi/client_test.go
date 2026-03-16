package meiguodizhi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gpt-team-api/internal/model"
)

func TestFetchProfileExtractsProfileFields(t *testing.T) {
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
		if got := r.Header.Get("Referer"); got != "https://www.meiguodizhi.com/" {
			t.Fatalf("expected referer header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"address":{"Address":"2350 Monroe Street","Telephone":"713-375-5326","City":"Houston","Zip_Code":"77028","State":"TX","State_Full":"Texas","Full_Name":"Grace Hopper","Birthday":"1906-12-09"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	result, err := client.FetchProfile(context.Background(), model.CardTypeUS)
	if err != nil {
		t.Fatalf("fetch profile: %v", err)
	}

	if result.FullName != "Grace Hopper" || result.Birthday != "1906-12-09" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.StreetAddress != "2350 Monroe Street" || result.City != "Houston" || result.StateFull != "Texas" || result.ZipCode != "77028" || result.PhoneNumber != "713-375-5326" {
		t.Fatalf("expected address fields to be extracted, got %+v", result)
	}
}

func TestFetchProfileUsesRegionSpecificPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		cardType    model.CardType
		wantPath    string
		wantReferer string
	}{
		{name: "uk", cardType: model.CardTypeUK, wantPath: "/uk-address", wantReferer: "https://www.meiguodizhi.com/uk-address"},
		{name: "es", cardType: model.CardTypeES, wantPath: "/es-address", wantReferer: "https://www.meiguodizhi.com/es-address"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}

				var requestBody map[string]string
				if err := json.Unmarshal(body, &requestBody); err != nil {
					t.Fatalf("decode body: %v", err)
				}

				if got := requestBody["path"]; got != testCase.wantPath {
					t.Fatalf("unexpected path: got %q want %q", got, testCase.wantPath)
				}
				if got := r.Header.Get("Referer"); got != testCase.wantReferer {
					t.Fatalf("unexpected referer: got %q want %q", got, testCase.wantReferer)
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"address":{"Full_Name":"Grace Hopper","Birthday":"1906-12-09"}}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, server.Client())
			if _, err := client.FetchProfile(context.Background(), testCase.cardType); err != nil {
				t.Fatalf("fetch profile: %v", err)
			}
		})
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
	if _, err := client.FetchProfile(context.Background(), model.CardTypeUS); err == nil {
		t.Fatalf("expected contract error")
	}
}
