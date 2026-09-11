package endpoint

import (
	"context"
	"errors"
	"testing"

	"github.com/bancofalabella/mortgage-catalogs-go/internal/model"
)

// MockCatalogService acts as a dummy implementation of the Service Use Case 
// for local testing of the Endpoint layer without Temporal or external dependencies.
type MockCatalogService struct {
	MockGetCatalog func(ctx context.Context, req model.GetCatalogRequest) (interface{}, error)
}

func (m *MockCatalogService) GetCatalog(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
	if m.MockGetCatalog != nil {
		return m.MockGetCatalog(ctx, req)
	}
	return nil, nil
}

func TestMakeGetCatalogEndpoint_Success(t *testing.T) {
	// 1. Arrange: Setup the mock to return a dummy response
	mockService := &MockCatalogService{
		MockGetCatalog: func(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
			// Assert that the mapped values arrived correctly to the service layer
			if req.CatalogName != "Destino" {
				t.Errorf("expected catalog name 'Destino', got %s", req.CatalogName)
			}
			
			// Dummy return data (the exact format will be handled by the handler)
			return []model.CatalogItem{{Code: "1", Description: "Dummy"}}, nil
		},
	}

	endpoint := MakeGetCatalogEndpoint(mockService)
	requestDTO := GetCatalogRequestDTO{
		CatalogName:   "Destino",
		Channel:       "WEB",
		Commerce:      "FALABELLA",
		TransactionID: "123",
	}

	// 2. Act: Execute the endpoint
	response, err := endpoint(context.Background(), requestDTO)
	
	// 3. Assert: Interoperability check
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	respDTO := response.(GetCatalogResponseDTO)
	if respDTO.Error != "" {
		t.Fatalf("expected no business error, got %v", respDTO.Error)
	}
	
	if respDTO.Data == nil {
		t.Fatalf("expected data in response")
	}
}

func TestMakeGetCatalogEndpoint_Error(t *testing.T) {
	// 1. Arrange: Setup mock to simulate a business error (e.g. Catalog Not Found)
	mockService := &MockCatalogService{
		MockGetCatalog: func(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
			return nil, errors.New("CatalogNotFoundError: invalid catalog")
		},
	}

	endpoint := MakeGetCatalogEndpoint(mockService)

	// 2. Act
	_, err := endpoint(context.Background(), GetCatalogRequestDTO{})
	// Endpoint actually returns (nil, err) now so the ServerErrorEncoder intercepts it
	if err == nil {
		t.Fatalf("expected error from endpoint, got nil")
	}

	if err.Error() != "CatalogNotFoundError: invalid catalog" {
		t.Errorf("expected error %v, got %v", "CatalogNotFoundError: invalid catalog", err)
	}
}
