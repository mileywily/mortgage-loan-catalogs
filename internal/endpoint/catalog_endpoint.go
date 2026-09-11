package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/service"
)

// GetCatalogRequestDTO is the input for the endpoint layer.
type GetCatalogRequestDTO struct {
	CatalogName   string
	Channel       string
	Commerce      string
	TransactionID string
}

// GetCatalogResponseDTO is the output of the endpoint layer.
type GetCatalogResponseDTO struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// MakeGetCatalogEndpoint builds the endpoint.Endpoint using the Service interface.
// This allows middlewares (tracing, logging) to wrap this endpoint easily.
func MakeGetCatalogEndpoint(s service.CatalogService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetCatalogRequestDTO)
		
		// Map DTO to pure Domain Model
		modelReq := model.GetCatalogRequest{
			CatalogName:   req.CatalogName,
			Channel:       req.Channel,
			Commerce:      req.Commerce,
			TransactionID: req.TransactionID,
		}

		result, err := s.GetCatalog(ctx, modelReq)
		if err != nil {
			return nil, err
		}

		return GetCatalogResponseDTO{Data: result}, nil
	}
}

