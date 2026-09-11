package service

import (
	"context"
	"github.com/bancofalabella/mortgage-catalogs-go/internal/model"
)

// CatalogService defines the pure Use Case interface for Mortgage Catalogs.
// The concrete implementation of this service will be responsible for initiating
// the Temporal Workflow.
type CatalogService interface {
	GetCatalog(ctx context.Context, req model.GetCatalogRequest) (interface{}, error)
}
