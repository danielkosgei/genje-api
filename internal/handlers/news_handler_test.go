package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"jalada/internal/models"
)

// fakeStore records the filter it was asked for and returns canned data.
type fakeStore struct {
	articles   []models.NewsArticle
	total      int
	err        error
	lastFilter models.NewsFilter
	article    *models.NewsArticle
}

func (f *fakeStore) ListArticles(_ context.Context, filter models.NewsFilter) ([]models.NewsArticle, int, error) {
	f.lastFilter = filter
	return f.articles, f.total, f.err
}

func (f *fakeStore) GetArticleByID(_ context.Context, _ uuid.UUID) (*models.NewsArticle, error) {
	return f.article, f.err
}

func (f *fakeStore) ListDataSources(_ context.Context) ([]models.Source, error) {
	return nil, f.err
}

func serveNews(t *testing.T, store NewsStore, method, target string, route func(*NewsHandler) http.HandlerFunc, params map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	if len(params) > 0 {
		rctx := chi.NewRouteContext()
		for k, v := range params {
			rctx.URLParams.Add(k, v)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}
	rec := httptest.NewRecorder()
	route(NewNewsHandler(store))(rec, req)
	return rec
}

func TestListArticlesPassesFiltersAndReturnsPage(t *testing.T) {
	store := &fakeStore{total: 3, articles: []models.NewsArticle{{Title: "A"}}}

	rec := serveNews(t, store, "GET", "/v1/news?category=Politics&q=ruto&limit=5&offset=10",
		func(h *NewsHandler) http.HandlerFunc { return h.ListArticles }, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if store.lastFilter.Category == nil || *store.lastFilter.Category != "Politics" {
		t.Fatalf("category filter not passed through: %+v", store.lastFilter)
	}
	if store.lastFilter.Query != "ruto" || store.lastFilter.Limit != 5 || store.lastFilter.Offset != 10 {
		t.Fatalf("filter = %+v", store.lastFilter)
	}

	var page models.PaginatedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if page.Total != 3 || page.Limit != 5 || page.Offset != 10 || page.HasMore {
		t.Fatalf("page metadata = %+v", page)
	}
}

func TestListArticlesRejectsBadInputBeforeQuerying(t *testing.T) {
	store := &fakeStore{}

	rec := serveNews(t, store, "GET", "/v1/news?limit=500",
		func(h *NewsHandler) http.HandlerFunc { return h.ListArticles }, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if store.lastFilter.Limit != 0 {
		t.Fatal("store was queried despite invalid input")
	}
}

func TestListArticlesReturnsEmptyArrayNotNull(t *testing.T) {
	rec := serveNews(t, &fakeStore{}, "GET", "/v1/news",
		func(h *NewsHandler) http.HandlerFunc { return h.ListArticles }, nil)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(raw["data"]) != "[]" {
		t.Fatalf("data = %s, want []", raw["data"])
	}
}

func TestListArticlesHidesStoreErrors(t *testing.T) {
	rec := serveNews(t, &fakeStore{err: context.DeadlineExceeded}, "GET", "/v1/news",
		func(h *NewsHandler) http.HandlerFunc { return h.ListArticles }, nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestGetArticleStatusCodes(t *testing.T) {
	id := uuid.New().String()
	route := func(h *NewsHandler) http.HandlerFunc { return h.GetArticle }

	if rec := serveNews(t, &fakeStore{}, "GET", "/v1/news/x", route, map[string]string{"id": "not-a-uuid"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid id: status = %d, want 400", rec.Code)
	}
	if rec := serveNews(t, &fakeStore{}, "GET", "/v1/news/"+id, route, map[string]string{"id": id}); rec.Code != http.StatusNotFound {
		t.Fatalf("missing article: status = %d, want 404", rec.Code)
	}

	title := "Found"
	found := &fakeStore{article: &models.NewsArticle{Title: title}}
	if rec := serveNews(t, found, "GET", "/v1/news/"+id, route, map[string]string{"id": id}); rec.Code != http.StatusOK {
		t.Fatalf("found article: status = %d, want 200", rec.Code)
	}
}
