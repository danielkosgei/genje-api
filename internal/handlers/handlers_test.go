package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"genje-api/internal/models"
)

func TestHealthHandler(t *testing.T) {
	handler := &Handler{}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.Health(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("handler returned wrong status: got %v want %v", response.Status, "ok")
	}

	if response.Version != "1.0.0" {
		t.Errorf("handler returned wrong version: got %v want %v", response.Version, "1.0.0")
	}
} 