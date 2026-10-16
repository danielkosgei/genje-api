package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"jalada/internal/models"
)

const (
	nomineeDefaultLimit = 50
	nomineeMaxLimit     = 200
	nomineeMaxQueryLen  = 100
)

// parseNomineeFilter reads GET /v1/nominees. Bad input is rejected with a reason, the same as news.
func parseNomineeFilter(r *http.Request) (models.NomineeFilter, error) {
	q := r.URL.Query()
	f := models.NomineeFilter{Limit: nomineeDefaultLimit}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > nomineeMaxLimit {
			return f, fmt.Errorf("limit must be a whole number from 1 to %d", nomineeMaxLimit)
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

	f.Party = strings.TrimSpace(q.Get("party"))
	f.County = strings.TrimSpace(q.Get("county"))
	f.Heading = strings.TrimSpace(q.Get("heading"))
	f.Query = strings.TrimSpace(q.Get("q"))
	if len([]rune(f.Query)) > nomineeMaxQueryLen {
		return f, fmt.Errorf("q must be at most %d characters", nomineeMaxQueryLen)
	}
	return f, nil
}
