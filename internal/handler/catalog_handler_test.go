package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bancofalabella/mortgage-catalogs-go/internal/endpoint"
)

func TestDecodeGetCatalogRequest_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/bfcl/mortgage-loan/catalogs/Destino", nil)
	req.Header.Set("X-Channel", "WEB")
	req.Header.Set("X-Commerce", "FALABELLA")
	req.Header.Set("X-Transaction-ID", "123")
	
	// Simular el PathValue que inyecta Go 1.22 ServeMux
	req.SetPathValue("catalog", "Destino")

	resp, err := DecodeGetCatalogRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("esperado nulo, obtenido %v", err)
	}

	dto, ok := resp.(endpoint.GetCatalogRequestDTO)
	if !ok {
		t.Fatalf("esperado GetCatalogRequestDTO")
	}

	if dto.CatalogName != "Destino" {
		t.Errorf("esperado Destino, obtenido %s", dto.CatalogName)
	}
}

func TestDecodeGetCatalogRequest_MissingHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/bfcl/mortgage-loan/catalogs/Destino", nil)
	// No headers
	req.SetPathValue("catalog", "Destino")

	_, err := DecodeGetCatalogRequest(context.Background(), req)
	if err != ErrMissingHeaders {
		t.Fatalf("esperado ErrMissingHeaders, obtenido %v", err)
	}
}

