package parser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExternalID(t *testing.T) {
	wb := NewWB()
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"https://www.wildberries.ru/catalog/179978204/detail.aspx", "179978204", true},
		{"  179978204  ", "179978204", true},
		{"179978204", "179978204", true},
		{"https://www.wildberries.ru/catalog/12345/detail.aspx?targetUrl=GP", "12345", true},
		{"совсем не ссылка", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := wb.ExternalID(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("ExternalID(%q) = (%q, %v), хотим (%q, %v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

func TestParseWBFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/wb_detail.json")
	if err != nil {
		t.Fatalf("чтение фикстуры: %v", err)
	}
	info, err := parseWB(body, "179978204")
	if err != nil {
		t.Fatalf("parseWB: %v", err)
	}
	if info.Price != 289900 { // приоритет у price.total (2899.00 ₽)
		t.Errorf("Price = %d, хотим 289900", info.Price)
	}
	if info.Title != "Наушники беспроводные TWS Bluetooth" {
		t.Errorf("Title = %q", info.Title)
	}
	if !info.IsAvailable {
		t.Errorf("IsAvailable = false, хотим true")
	}
	if info.ImageURL == "" {
		t.Errorf("ImageURL пустой")
	}
}

func TestWBFetchOverHTTP(t *testing.T) {
	body, err := os.ReadFile("testdata/wb_detail.json")
	if err != nil {
		t.Fatalf("чтение фикстуры: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("nm") != "179978204" {
			t.Errorf("в запросе нет nm=179978204: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	wb := NewWB()
	wb.baseURL = srv.URL

	info, err := wb.Fetch(context.Background(), "179978204")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if info.Price != 289900 {
		t.Errorf("Price = %d, хотим 289900", info.Price)
	}
}
