package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"jalada/internal/category"
	"jalada/internal/models"
)

const (
	newsDefaultLimit = 20
	newsMaxLimit     = 100
	newsMaxQueryLen  = 100
)

// parseNewsFilter reads the query string of GET /v1/news. Bad input is reported, not ignored, so a
// client that misspells a parameter learns about it instead of silently getting unfiltered results.
func parseNewsFilter(r *http.Request) (models.NewsFilter, error) {
	q := r.URL.Query()
	f := models.NewsFilter{Limit: newsDefaultLimit}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > newsMaxLimit {
			return f, fmt.Errorf("limit must be a whole number from 1 to %d", newsMaxLimit)
		}
		f.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return f, fmt.Errorf("offset must be a whole number of 0 or more")
		}
		f.Offset = n
	}

	if v := q.Get("source_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, fmt.Errorf("source_id must be a UUID")
		}
		f.SourceID = &id
	}
	if v := q.Get("politician_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, fmt.Errorf("politician_id must be a UUID")
		}
		f.PoliticianID = &id
	}

	if v := q.Get("election_related"); v != "" {
		switch v {
		case "true":
			t := true
			f.ElectionRelated = &t
		case "false":
			// Explicit false means "no restriction", which is the default
		default:
			return f, fmt.Errorf("election_related must be true or false")
		}
	}

	if v := strings.TrimSpace(q.Get("category")); v != "" {
		canonical, ok := canonicalCategory(v)
		if !ok {
			return f, fmt.Errorf("category must be one of: %s", strings.Join(category.All, ", "))
		}
		f.Category = &canonical
	}

	if v := strings.TrimSpace(q.Get("q")); v != "" {
		if len([]rune(v)) > newsMaxQueryLen {
			return f, fmt.Errorf("q must be at most %d characters", newsMaxQueryLen)
		}
		f.Query = v
	}

	return f, nil
}

// canonicalCategory matches a category name ignoring case, returning the stored spelling.
func canonicalCategory(name string) (string, bool) {
	for _, c := range category.All {
		if strings.EqualFold(c, name) {
			return c, true
		}
	}
	return "", false
}
