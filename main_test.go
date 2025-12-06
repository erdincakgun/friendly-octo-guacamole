package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// Server & responseWriter Tests
// =============================================================================

func TestNewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.menuItems == nil {
		t.Fatal("menuItems map is nil")
	}

	expectedCount := 5
	if len(server.menuItems) != expectedCount {
		t.Errorf("expected %d menu items, got %d", expectedCount, len(server.menuItems))
	}

	// Verify all expected IDs exist
	expectedIDs := []string{"1", "2", "3", "4", "5"}
	for _, id := range expectedIDs {
		if _, exists := server.menuItems[id]; !exists {
			t.Errorf("expected menu item with ID %q to exist", id)
		}
	}

	// Verify a specific menu item
	item, exists := server.menuItems["1"]
	if !exists {
		t.Fatal("menu item '1' not found")
	}
	if item.Name != "Margherita Pizza" {
		t.Errorf("expected name 'Margherita Pizza', got %q", item.Name)
	}
	if item.Price != 12.99 {
		t.Errorf("expected price 12.99, got %f", item.Price)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// Initially should be OK (default)
	if rw.statusCode != http.StatusOK {
		t.Errorf("expected initial statusCode %d, got %d", http.StatusOK, rw.statusCode)
	}

	// Write a different status
	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("expected statusCode %d after WriteHeader, got %d", http.StatusNotFound, rw.statusCode)
	}

	// Verify it was also written to the underlying ResponseWriter
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected underlying recorder code %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestResponseWriter_MultipleWriteHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	rw.WriteHeader(http.StatusBadRequest) // Second call

	// Our wrapper should record the last call
	if rw.statusCode != http.StatusBadRequest {
		t.Errorf("expected statusCode %d, got %d", http.StatusBadRequest, rw.statusCode)
	}
}

// =============================================================================
// writeJSON Tests
// =============================================================================

func TestWriteJSON_Success(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]string{"message": "hello"}
	err := writeJSON(rec, http.StatusOK, data)

	if err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("expected message 'hello', got %q", result["message"])
	}
}

func TestWriteJSON_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name   string
		status int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Accepted", http.StatusAccepted},
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			data := map[string]int{"status": tc.status}

			err := writeJSON(rec, tc.status, data)

			if err != nil {
				t.Fatalf("writeJSON returned error: %v", err)
			}

			if rec.Code != tc.status {
				t.Errorf("expected status %d, got %d", tc.status, rec.Code)
			}
		})
	}
}

func TestWriteJSON_ComplexData(t *testing.T) {
	rec := httptest.NewRecorder()

	data := MenuItem{
		ID:          "test-1",
		Name:        "Test Item",
		Price:       9.99,
		Available:   true,
		Description: "A test item",
		Restaurant:  "Test Restaurant",
		Category:    "Test",
		PrepTime:    10,
	}

	err := writeJSON(rec, http.StatusOK, data)
	if err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	var result MenuItem
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.ID != "test-1" {
		t.Errorf("expected ID 'test-1', got %q", result.ID)
	}
	if result.Price != 9.99 {
		t.Errorf("expected Price 9.99, got %f", result.Price)
	}
}

func TestWriteJSON_UnmarshalableData(t *testing.T) {
	rec := httptest.NewRecorder()

	// Channels cannot be marshaled to JSON
	data := make(chan int)
	err := writeJSON(rec, http.StatusOK, data)

	if err == nil {
		t.Error("expected error for unmarshalable data, got nil")
	}

	// Should fallback to 500 Internal Server Error
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on marshal error, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// =============================================================================
// writeError Tests
// =============================================================================

func TestWriteError_BasicError(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, http.StatusBadRequest, "Invalid input")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if result["error"] != "Bad Request" {
		t.Errorf("expected error 'Bad Request', got %q", result["error"])
	}

	if result["message"] != "Invalid input" {
		t.Errorf("expected message 'Invalid input', got %q", result["message"])
	}
}

func TestWriteError_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		status        int
		expectedError string
	}{
		{http.StatusBadRequest, "Bad Request"},
		{http.StatusNotFound, "Not Found"},
		{http.StatusInternalServerError, "Internal Server Error"},
		{http.StatusUnauthorized, "Unauthorized"},
		{http.StatusForbidden, "Forbidden"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedError, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeError(rec, tc.status, "test message")

			if rec.Code != tc.status {
				t.Errorf("expected status %d, got %d", tc.status, rec.Code)
			}

			var result map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if result["error"] != tc.expectedError {
				t.Errorf("expected error %q, got %q", tc.expectedError, result["error"])
			}
		})
	}
}

// =============================================================================
// healthHandler Tests
// =============================================================================

func TestHealthHandler_Success(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %q", result["status"])
	}

	if result["timestamp"] == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestHealthHandler_ReturnsValidTimestamp(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.healthHandler(rec, req)

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Timestamp should be in RFC3339 format
	timestamp := result["timestamp"]
	if !strings.Contains(timestamp, "T") || !strings.Contains(timestamp, ":") {
		t.Errorf("timestamp %q doesn't appear to be RFC3339 format", timestamp)
	}
}

// =============================================================================
// menuHandler Tests
// =============================================================================

func TestMenuHandler_SuccessPath(t *testing.T) {
	server := NewServer()

	// Run multiple times to get at least one success (90% success rate)
	var gotSuccess bool
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
		rec := httptest.NewRecorder()

		server.menuHandler(rec, req)

		if rec.Code == http.StatusOK {
			gotSuccess = true

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Check count field
			count, ok := result["count"].(float64) // JSON numbers are float64
			if !ok {
				t.Error("expected 'count' field in response")
			}
			if int(count) != 5 {
				t.Errorf("expected count 5, got %v", count)
			}

			// Check menu_items field
			menuItems, ok := result["menu_items"].([]interface{})
			if !ok {
				t.Error("expected 'menu_items' array in response")
			}
			if len(menuItems) != 5 {
				t.Errorf("expected 5 menu items, got %d", len(menuItems))
			}

			break
		}
	}

	if !gotSuccess {
		t.Error("menuHandler never returned success after 50 attempts (extremely unlikely)")
	}
}

func TestMenuHandler_FailurePath(t *testing.T) {
	server := NewServer()

	// Run multiple times to get at least one failure (10% failure rate)
	var gotFailure bool
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
		rec := httptest.NewRecorder()

		server.menuHandler(rec, req)

		if rec.Code == http.StatusInternalServerError {
			gotFailure = true

			var result map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal error response: %v", err)
			}

			if result["error"] != "Internal Server Error" {
				t.Errorf("expected error 'Internal Server Error', got %q", result["error"])
			}

			expectedMsg := "Failed to fetch menu items from restaurant database"
			if result["message"] != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, result["message"])
			}

			break
		}
	}

	if !gotFailure {
		t.Error("menuHandler never returned failure after 100 attempts (extremely unlikely with 10% failure rate)")
	}
}

func TestMenuHandler_ResponseStructure(t *testing.T) {
	server := NewServer()

	// Get a successful response
	var rec *httptest.ResponseRecorder
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/menu", nil)
		rec = httptest.NewRecorder()
		server.menuHandler(rec, req)
		if rec.Code == http.StatusOK {
			break
		}
	}

	if rec.Code != http.StatusOK {
		t.Skip("could not get successful response after 50 attempts")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	menuItems := result["menu_items"].([]interface{})
	for _, item := range menuItems {
		itemMap := item.(map[string]interface{})

		// Verify all expected fields exist
		expectedFields := []string{"id", "name", "price", "available", "description", "restaurant", "category", "prep_time_minutes"}
		for _, field := range expectedFields {
			if _, exists := itemMap[field]; !exists {
				t.Errorf("expected field %q in menu item", field)
			}
		}
	}
}

// =============================================================================
// menuItemByIDHandler Tests
// =============================================================================

func TestMenuItemByIDHandler_ValidID(t *testing.T) {
	server := NewServer()

	testCases := []struct {
		id           string
		expectedName string
	}{
		{"1", "Margherita Pizza"},
		{"2", "Chicken Pad Thai"},
		{"3", "Classic Burger"},
		{"4", "Caesar Salad"},
		{"5", "Sushi Platter"},
	}

	for _, tc := range testCases {
		t.Run("ID_"+tc.id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/menu/"+tc.id, nil)
			rec := httptest.NewRecorder()

			server.menuItemByIDHandler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			menuItem, ok := result["menu_item"].(map[string]interface{})
			if !ok {
				t.Fatal("expected 'menu_item' in response")
			}

			if menuItem["name"] != tc.expectedName {
				t.Errorf("expected name %q, got %v", tc.expectedName, menuItem["name"])
			}
		})
	}
}

func TestMenuItemByIDHandler_EmptyID(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/menu/", nil)
	rec := httptest.NewRecorder()

	server.menuItemByIDHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["message"] != "Menu item ID is required" {
		t.Errorf("expected message 'Menu item ID is required', got %q", result["message"])
	}
}

func TestMenuItemByIDHandler_WhitespaceID(t *testing.T) {
	server := NewServer()
	// Use URL encoding for whitespace to avoid httptest.NewRequest panic
	req := httptest.NewRequest(http.MethodGet, "/api/menu/%20%20%20", nil)
	rec := httptest.NewRecorder()

	server.menuItemByIDHandler(rec, req)

	// After URL decoding and trimming, the ID becomes empty, so we expect BadRequest
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for whitespace ID, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMenuItemByIDHandler_NonExistentID(t *testing.T) {
	server := NewServer()

	testCases := []string{"999", "abc", "0", "-1", "nonexistent"}

	for _, id := range testCases {
		t.Run("ID_"+id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/menu/"+id, nil)
			rec := httptest.NewRecorder()

			server.menuItemByIDHandler(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
			}

			var result map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if result["error"] != "Not Found" {
				t.Errorf("expected error 'Not Found', got %q", result["error"])
			}

			expectedMsg := "Menu item with ID '" + id + "' not found"
			if result["message"] != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, result["message"])
			}
		})
	}
}

func TestMenuItemByIDHandler_ResponseStructure(t *testing.T) {
	server := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/menu/1", nil)
	rec := httptest.NewRecorder()

	server.menuItemByIDHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	menuItem := result["menu_item"].(map[string]interface{})

	// Verify all fields with expected values
	expectations := map[string]interface{}{
		"id":                "1",
		"name":              "Margherita Pizza",
		"price":             12.99,
		"available":         true,
		"description":       "Fresh mozzarella, tomato sauce, basil",
		"restaurant":        "Tony's Pizza",
		"category":          "Pizza",
		"prep_time_minutes": float64(20), // JSON numbers are float64
	}

	for field, expected := range expectations {
		if menuItem[field] != expected {
			t.Errorf("field %q: expected %v, got %v", field, expected, menuItem[field])
		}
	}
}

// =============================================================================
// loggingMiddleware Tests
// =============================================================================

func TestLoggingMiddleware_PassesRequestThrough(t *testing.T) {
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := loggingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("middleware did not call the wrapped handler")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got %q", string(body))
	}
}

func TestLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	testCases := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, status := range testCases {
		t.Run(http.StatusText(status), func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})

			wrapped := loggingMiddleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != status {
				t.Errorf("expected status %d, got %d", status, rec.Code)
			}
		})
	}
}

func TestLoggingMiddleware_DefaultStatusOK(t *testing.T) {
	// Handler that writes body without explicit WriteHeader
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("implicit 200"))
	})

	wrapped := loggingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected implicit status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLoggingMiddleware_PreservesHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	wrapped := loggingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("X-Custom-Header") != "test-value" {
		t.Errorf("expected X-Custom-Header 'test-value', got %q", rec.Header().Get("X-Custom-Header"))
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}
}

// =============================================================================
// Integration Tests (Full HTTP flow)
// =============================================================================

func TestIntegration_FullServerRouting(t *testing.T) {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/api/menu", server.menuHandler)
	mux.HandleFunc("/api/menu/", server.menuItemByIDHandler)

	handler := loggingMiddleware(mux)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	t.Run("Health endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/health")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Menu item by ID endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/menu/1")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		if result["menu_item"] == nil {
			t.Error("expected menu_item in response")
		}
	})

	t.Run("Menu item not found", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/menu/999")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// =============================================================================
// MenuItem Struct Tests
// =============================================================================

func TestMenuItem_JSONSerialization(t *testing.T) {
	item := MenuItem{
		ID:          "test-1",
		Name:        "Test Dish",
		Price:       15.50,
		Available:   true,
		Description: "A delicious test",
		Restaurant:  "Test Kitchen",
		Category:    "Test Category",
		PrepTime:    30,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal MenuItem: %v", err)
	}

	var parsed MenuItem
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal MenuItem: %v", err)
	}

	if parsed.ID != item.ID {
		t.Errorf("ID mismatch: expected %q, got %q", item.ID, parsed.ID)
	}
	if parsed.Name != item.Name {
		t.Errorf("Name mismatch: expected %q, got %q", item.Name, parsed.Name)
	}
	if parsed.Price != item.Price {
		t.Errorf("Price mismatch: expected %f, got %f", item.Price, parsed.Price)
	}
	if parsed.Available != item.Available {
		t.Errorf("Available mismatch: expected %v, got %v", item.Available, parsed.Available)
	}
	if parsed.PrepTime != item.PrepTime {
		t.Errorf("PrepTime mismatch: expected %d, got %d", item.PrepTime, parsed.PrepTime)
	}
}

func TestMenuItem_JSONFieldNames(t *testing.T) {
	item := MenuItem{
		ID:       "1",
		PrepTime: 15,
	}

	data, _ := json.Marshal(item)
	jsonStr := string(data)

	// Verify JSON field names match the tags
	expectedFields := []string{
		`"id"`,
		`"name"`,
		`"price"`,
		`"available"`,
		`"description"`,
		`"restaurant"`,
		`"category"`,
		`"prep_time_minutes"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("expected JSON to contain field %s, got: %s", field, jsonStr)
		}
	}
}
