package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	corsMiddleware := CORS()
	
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			// Log error in real application
			return
		}
	})
	
	handler := corsMiddleware(testHandler)

	tests := []struct {
		name           string
		method         string
		origin         string
		requestHeaders string
		expectedStatus int
		checkHeaders   map[string]string
	}{
		{
			name:           "Simple GET request",
			method:         "GET",
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			origin:         "https://example.com",
			requestHeaders: "Content-Type,Authorization",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name:           "POST request",
			method:         "POST",
			origin:         "https://different.com",
			expectedStatus: http.StatusOK,
			checkHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			for header, expectedValue := range tt.checkHeaders {
				actualValue := rr.Header().Get(header)
				if !containsAllValues(actualValue, expectedValue) {
					t.Errorf("Header %s: expected to contain %s, got %s", header, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	requestIDMiddleware := RequestID()
	
	// Create a test handler that captures the request ID
	var capturedRequestID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = GetRequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	
	handler := requestIDMiddleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	handler.ServeHTTP(rr, req)

	// Check that request ID was set in context
	if capturedRequestID == "" {
		t.Error("Request ID should be set in context")
	}

	// Check that request ID header was set
	headerRequestID := rr.Header().Get("X-Request-ID")
	if headerRequestID == "" {
		t.Error("X-Request-ID header should be set")
	}

	// Check that context and header have the same request ID
	if capturedRequestID != headerRequestID {
		t.Errorf("Context request ID (%s) should match header request ID (%s)", capturedRequestID, headerRequestID)
	}

	// Check that request ID is valid hex string
	if len(capturedRequestID) != 16 {
		t.Errorf("Request ID should be 16 characters long, got %d", len(capturedRequestID))
	}

	for _, char := range capturedRequestID {
		if !isHexChar(char) {
			t.Errorf("Request ID should contain only hex characters, found: %c", char)
		}
	}
}

func TestRequestIDUniqueness(t *testing.T) {
	requestIDMiddleware := RequestID()
	
	var requestIDs []string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := GetRequestIDFromContext(r.Context()); id != "" {
			requestIDs = append(requestIDs, id)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	handler := requestIDMiddleware(testHandler)

	// Generate multiple request IDs
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Check that all request IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range requestIDs {
		if uniqueIDs[id] {
			t.Errorf("Request ID %s is not unique", id)
		}
		uniqueIDs[id] = true
	}

	if len(uniqueIDs) != 100 {
		t.Errorf("Expected 100 unique request IDs, got %d", len(uniqueIDs))
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Test the generateRequestID function directly
	id1 := generateRequestID()
	id2 := generateRequestID()

	// Should be different
	if id1 == id2 {
		t.Error("generateRequestID should produce unique IDs")
	}

	// Should be 16 characters (8 bytes * 2 hex chars per byte)
	if len(id1) != 16 {
		t.Errorf("Expected request ID length 16, got %d", len(id1))
	}
	if len(id2) != 16 {
		t.Errorf("Expected request ID length 16, got %d", len(id2))
	}

	// Should be valid hex
	for _, char := range id1 {
		if !isHexChar(char) {
			t.Errorf("Request ID should contain only hex characters, found: %c in %s", char, id1)
		}
	}
}

func TestMiddlewareChaining(t *testing.T) {
	// Test that both middleware can be chained together
	corsMiddleware := CORS()
	requestIDMiddleware := RequestID()
	
	var capturedRequestID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = GetRequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			// Log error in real application
			return
		}
	})
	
	// Chain the middleware
	handler := corsMiddleware(requestIDMiddleware(testHandler))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that both middleware worked
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check CORS headers
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS headers should be set")
	}

	// Check Request ID
	if capturedRequestID == "" {
		t.Error("Request ID should be set by middleware chain")
	}
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header should be set by middleware chain")
	}
}

// Helper functions
func containsAllValues(actual, expected string) bool {
	if expected == actual {
		return true
	}
	
	// For comma-separated values, check if all expected values are present
	expectedParts := strings.Split(expected, ",")
	for _, part := range expectedParts {
		if !strings.Contains(actual, strings.TrimSpace(part)) {
			return false
		}
	}
	return true
}

func isHexChar(char rune) bool {
	return (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')
} 