// Package category maps the free-form section labels that news feeds use onto a
// small fixed set of categories, so the API can filter reliably.
package category

import "strings"

// Canonical categories. Every article is stored with exactly one of these.
const (
	Business      = "Business"
	Politics      = "Politics"
	Sports        = "Sports"
	Entertainment = "Entertainment"
	Technology    = "Technology"
	Health        = "Health"
	World         = "World"
	News          = "News"
)

// All lists every canonical category, in display order.
var All = []string{Business, Politics, Sports, Entertainment, Technology, Health, World, News}

// rule maps keywords to a category. Earlier rules win when a label matches several.
type rule struct {
	category string
	keywords []string
}

var rules = []rule{
	{Politics, []string{"politic", "election", "parliament", "government", "county", "governor", "ruto", "raila", "uhuru", "gachagua", "kalonzo", "odinga", "matiang", "cabinet", "senator"}},
	{Business, []string{"business", "econom", "financ", "money", "market", "trade", "compan"}},
	{Sports, []string{"sport", "football", "soccer", "rugby", "cricket", "athlet", "premier league", "league", "fifa", "messi", "goal", "match", "tournament", "championship", "world cup"}},
	{Entertainment, []string{"entertain", "celebrit", "music", "movie", "film", "showbiz", "lifestyle"}},
	{Technology, []string{"tech", "digital", "innovation", "gadget", "startup", "science"}},
	{Health, []string{"health", "medic", "hospital", "wellness"}},
	{World, []string{"world", "africa", "international", "global", "abroad"}},
	{News, []string{"news", "national", "kenya", "local", "general", "crime", "courts"}},
}

// Normalize returns the canonical category for an article.
//
// The article's own label is tried first (for example "NATIONAL NEWS" or "Sports"), then the path
// of its URL (for example standardmedia.co.ke/politics/article/...), then its title. Feed-level
// defaults are the weakest signal: a section feed often carries stories from other sections.
// Anything still unmatched becomes News.
func Normalize(label, articleURL, title, feedDefault string) string {
	if c := match(label); c != "" {
		return c
	}
	if c := match(urlPath(articleURL)); c != "" {
		return c
	}
	if c := match(title); c != "" {
		return c
	}
	if c := match(feedDefault); c != "" {
		return c
	}
	return News
}

// match returns the first canonical category whose keyword appears in s, or "" if none does.
func match(s string) string {
	lower := strings.ToLower(s)
	if strings.TrimSpace(lower) == "" {
		return ""
	}
	for _, r := range rules {
		for _, kw := range r.keywords {
			if strings.Contains(lower, kw) {
				return r.category
			}
		}
	}
	return ""
}

// urlPath returns the path of rawURL without the scheme and host, so keywords in a domain
// name such as "citizen.digital" cannot be mistaken for a section.
func urlPath(rawURL string) string {
	s := rawURL
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.Index(s, "/"); i >= 0 {
		return s[i:]
	}
	return ""
}
