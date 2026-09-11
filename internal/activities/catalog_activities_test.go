package activities

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
)

// TestFetchLegacyCatalogActivity_Interoperability_Success uses an httptest.Server
// to mock the Apigee/Legacy system locally and guarantee interoperability.
func TestFetchLegacyCatalogActivity_Interoperability_Success(t *testing.T) {
	// 1. Arrange: Create a local Mock Server for Apigee
	mockLegacyJSON := `[{"codigo_adm":"1","descripcion":"valor 1"}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate that the ACL forwards the required headers to the legacy system
		if r.Header.Get("X-Channel") != "APP" {
			t.Errorf("expected X-Channel APP, got %s", r.Header.Get("X-Channel"))
		}
		if r.Header.Get("X-Commerce") != "FALABELLA" {
			t.Errorf("expected X-Commerce FALABELLA, got %s", r.Header.Get("X-Commerce"))
		}

		// Ensure the path is correct
		if r.URL.Path != "/api/catalogo_detail/Comunas" {
			t.Errorf("unexpected path requested to legacy: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockLegacyJSON))
	}))
	defer server.Close() // Close local server after test

	// Inject the mock server URL as the Apigee Base URL
	activity := NewLegacyCatalogActivity(server.URL, "", "", server.Client())

	req := model.GetCatalogRequest{
		CatalogName:   "Comunas",
		Channel:       "APP",
		Commerce:      "FALABELLA",
		TransactionID: "trx-001",
	}

	// 2. Act: Call the activity
	result, err := activity.FetchLegacyCatalogActivity(context.Background(), req)

	// 3. Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatalf("expected valid result slice")
	}
}

// TestFetchLegacyCatalogActivity_Interoperability_Errors verifies that
// legacy HTTP errors are properly mapped to Domain errors for Temporal to handle.
func TestFetchLegacyCatalogActivity_Interoperability_Errors(t *testing.T) {
	tests := []struct {
		name           string
		legacyStatus   int
		expectedErrMsg string
	}{
		{"InvalidCatalog", http.StatusBadRequest, "InvalidCatalogError"},
		{"Unauthorized", http.StatusUnauthorized, "AuthenticationError"},
		{"NotFound", http.StatusNotFound, "CatalogNotFoundError"},
		{"InternalLegacyError", http.StatusInternalServerError, "LegacySystemError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Server returning the specific HTTP status
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.legacyStatus)
			}))

			activity := NewLegacyCatalogActivity(server.URL, "", "", server.Client())

			_, err := activity.FetchLegacyCatalogActivity(context.Background(), model.GetCatalogRequest{CatalogName: "Destino"})

			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			// Validate that the error type contains our expected Domain mapped error
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("expected error to contain %s, got %v", tt.expectedErrMsg, err)
			}

			server.Close()
		})
	}
}
