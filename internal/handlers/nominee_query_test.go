package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseNomineeFilter(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"defaults", "", false},
		{"valid filters", "party=United+Democratic+Alliance&county=Nakuru&heading=x&q=otieno&limit=200&offset=0", false},
		{"limit zero", "limit=0", true},
		{"limit too large", "limit=201", true},
		{"limit not a number", "limit=abc", true},
		{"negative offset", "offset=-1", true},
		{"query too long", "q=" + strings.Repeat("a", 101), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/v1/nominees?"+tc.query, nil)
			_, err := parseNomineeFilter(r)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseNomineeFilter(%q) error = %v, wantErr %v", tc.query, err, tc.wantErr)
			}
		})
	}
}
