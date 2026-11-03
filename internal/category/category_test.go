package category

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		url      string
		title    string
		fallback string
		want     string
	}{
		{"label wins over url, title and fallback", "Sports", "https://example.com/politics/x", "Ruto speaks", "Business", Sports},
		{"upper case label is matched", "NATIONAL NEWS", "", "", "", News},
		{"africa label becomes world", "Africa", "", "", "", World},
		{"label keyword inside phrase", "Kenya Politics Today", "", "", "", Politics},
		{"url path used when label is missing", "", "https://www.standardmedia.co.ke/politics/article/123/x", "", "", Politics},
		{"domain name is not a section", "", "https://citizen.digital/article/123", "", "", News},
		{"title used when label and url give nothing", "", "https://example.com/a/1", "Premier League in limbo", "", Sports},
		{"title beats the feed default", "", "https://example.com/a/1", "Ruto blames opposition", "Entertainment", Politics},
		{"no matching title falls back to the feed default", "", "https://example.com/a/1", "Gunmen kill 24 in attacks", "Entertainment", Entertainment},
		{"feed default used when nothing else matches", "", "https://www.tuko.co.ke/rss/x", "", "Entertainment", Entertainment},
		{"politician names are politics", "", "", "Maurice Ogeta recalls Raila", "", Politics},
		{"unmatched everything falls back to news", "Zzz", "https://example.com/zzz", "Qqq", "", News},
		{"empty everything falls back to news", "", "", "", "", News},
		{"technology keywords", "Gadgets and Tech", "", "", "", Technology},
		{"ambiguous label follows rule order", "Digital Economy", "", "", "", Business},
		{"health keywords", "Public Health", "", "", "", Health},
		{"business keywords", "Markets", "", "", "", Business},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.label, tt.url, tt.title, tt.fallback); got != tt.want {
				t.Fatalf("Normalize(%q, %q, %q, %q) = %q, want %q", tt.label, tt.url, tt.title, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestNormalizeAlwaysReturnsACanonicalCategory(t *testing.T) {
	inputs := []string{"", "x", "Sports", "Africa", "junk 123", "https://example.com/a/b"}
	valid := map[string]bool{}
	for _, c := range All {
		valid[c] = true
	}
	for _, in := range inputs {
		if got := Normalize(in, in, in, in); !valid[got] {
			t.Fatalf("Normalize(%q) = %q, which is not a canonical category", in, got)
		}
	}
}
