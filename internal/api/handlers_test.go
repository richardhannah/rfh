package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Skip testing healthHandler since it requires real DB connection
// Focus on testing utility functions instead

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		expectCode int
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			expectCode: http.StatusOK,
		},
		{
			name:       "created response",
			status:     http.StatusCreated,
			data:       map[string]string{"id": "123"},
			expectCode: http.StatusCreated,
		},
		{
			name:       "nil data",
			status:     http.StatusOK,
			data:       nil,
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.status, tt.data)

			if w.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			// Verify JSON is valid
			var result interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		expectCode int
	}{
		{
			name:       "bad request",
			status:     http.StatusBadRequest,
			message:    "invalid input",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "not found",
			status:     http.StatusNotFound,
			message:    "resource not found",
			expectCode: http.StatusNotFound,
		},
		{
			name:       "internal server error",
			status:     http.StatusInternalServerError,
			message:    "something went wrong",
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeError(w, tt.status, tt.message)

			if w.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			var errorResponse map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
			}

			if errorResponse["error"] != tt.message {
				t.Errorf("expected error message %q, got %v", tt.message, errorResponse["error"])
			}
		})
	}
}

// Skip handler tests that require database connections
// These would need proper integration tests with a test database