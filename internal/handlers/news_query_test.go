package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func parse(t *testing.T, query string) (limit, offset int, err error) {
	t.Helper()
	r := httptest.NewRequest("GET", "/v1/news?"+query, nil)
	f, err := parseNewsFilter(r)
	return f.Limit, f.Offset, err
}

func TestParseNewsFilterPagination(t *testing.T) {
	tests := []struct {
		query      string
		wantLimit  int
		wantOffset int
		wantErr    bool
	}{
		{"", 20, 0, false},
		{"limit=1&offset=0", 1, 0, false},
		{"limit=100&offset=40", 100, 40, false},
		{"limit=0", 0, 0, true},
		{"limit=101", 0, 0, true},
		{"limit=abc", 0, 0, true},
		{"offset=-1", 0, 0, true},
		{"offset=x", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			limit, offset, err := parse(t, tt.query)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if limit != tt.wantLimit || offset != tt.wantOffset {
				t.Fatalf("limit/offset = %d/%d, want %d/%d", limit, offset, tt.wantLimit, tt.wantOffset)
			}
		})
	}
}

func TestParseNewsFilterCategoryIsCanonicalised(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/news?category=sPoRtS", nil)
	f, err := parseNewsFilter(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Category == nil || *f.Category != "Sports" {
		t.Fatalf("category = %v, want Sports", f.Category)
	}
}

func TestParseNewsFilterRejectsUnknownCategory(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/news?category=Gardening", nil)
	_, err := parseNewsFilter(r)
	if err == nil || !strings.Contains(err.Error(), "Business") {
		t.Fatalf("expected an error listing valid categories, got %v", err)
	}
}

func TestParseNewsFilterQuery(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/news?q=++ruto+%25+", nil)
	f, err := parseNewsFilter(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Query != "ruto %" {
		t.Fatalf("query = %q, want trimmed %q", f.Query, "ruto %")
	}

	long := strings.Repeat("a", newsMaxQueryLen+1)
	r = httptest.NewRequest("GET", "/v1/news?q="+long, nil)
	if _, err := parseNewsFilter(r); err == nil {
		t.Fatal("expected an error for an over-long query")
	}
}

func TestParseNewsFilterUUIDsAndFlags(t *testing.T) {
	bad := []string{"/v1/news?source_id=nope", "/v1/news?politician_id=nope", "/v1/news?election_related=yes"}
	for _, u := range bad {
		if _, err := parseNewsFilter(httptest.NewRequest("GET", u, nil)); err == nil {
			t.Fatalf("%s: expected an error", u)
		}
	}

	f, err := parseNewsFilter(httptest.NewRequest("GET", "/v1/news?election_related=true", nil))
	if err != nil || f.ElectionRelated == nil || !*f.ElectionRelated {
		t.Fatalf("election_related=true should set the flag, got %v, %v", f.ElectionRelated, err)
	}
}
