package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bancofalabella/mortgage-loan-catalogs/internal/activities"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/endpoint"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/handler"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/service"
	httptransport "github.com/go-kit/kit/transport/http"
)

// dummyService implements the Service interface just for local Postman testing
type dummyService struct{}

func (s *dummyService) GetCatalog(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
	// Dummy response matching EXACTLY the Java legacy documentation
	if req.CatalogName == "SegurosIncendio" || req.CatalogName == "SegurosDesgravamen" || req.CatalogName == "SegurosCesantia" {
		return []model.InsuranceCatalogItem{
			{
				InsuranceIdentifier:  "1100438-01-2023-000",
				Description:          "INCENDIO - Everest compaÃ±ia de seguros generales Chile(0.2255300)",
				Policy:               "100438-01-2023-000",
				CompanyName:          "Everest compaÃ±ia de seguros generales Chile",
				CompanyCode:          39,
				InsuranceTypeCode:    0,
				PolicyCorrelative:    28,
				Rate:                 0.22553,
				Factor:               0.0002255,
				IndividualPolicyFlag: 0,
				PerQuotaValueFlag:    0,
				ExternalPolicyFlag:   0,
			},
		}, nil
	}

	if req.CatalogName == "TiposDocumentos" {
		return []model.TipoDocumentoCatalogItem{
			{
				Code:        1,
				Description: "Escritura de Propiedad",
				GroupID:     2,
			},
		}, nil
	}

	if req.CatalogName == "Vacio" || req.CatalogName == "ErrorRed" {
		// Simulates a network error or empty results from backend
		return []model.NoResultsCatalogItem{
			{
				CodRespuesta: 3,
				Mensaje:      "Sin resultados.",
				Excepcion:    "Ninguna",
			},
		}, nil
	}

	// Valid catalogs as per Java Swagger documentation
	validCatalogs := map[string]bool{
		"Destino": true, "Objetivo": true, "TiposConstrucciones": true, "TiposInmuebles": true,
		"Antiguedades": true, "Comunas": true, "Regiones": true, "Producto": true,
		"GruposDocumentos": true, "TiposDocumentos": true, "TipoParticipante": true,
		"CampanasHipotecarias": true, "SegurosIncendio": true, "SegurosDesgravamen": true, "SegurosCesantia": true,
	}

	if !validCatalogs[req.CatalogName] {
		// Emulate upstream BFCL 404 which gets mapped to CatalogNotFoundError
		return nil, errors.New("CatalogNotFoundError")
	}

	if req.CatalogName == "Comunas" {
		return []model.ComunaCatalogItem{
			{
				Code:        "01101",
				Description: "Iquique",
				RegionID:    1,
			},
		}, nil
	}

	// Standard Catalog exact match
	return []model.CatalogItem{
		{
			Code:        "1",
			Description: "Vivienda Principal",
		},
	}, nil
}

type localTestService struct {
	activity activities.LegacyCatalogClient
}

func (s *localTestService) GetCatalog(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
	// Directly calls the Activity (bypassing Temporal) for local parity testing
	return s.activity.FetchLegacyCatalogActivity(ctx, req)
}

func main() {
	var svc service.CatalogService

	if os.Getenv("USE_REAL_BACKEND") == "true" {
		finnflowURL := os.Getenv("FINNFLOW_URL")
		if finnflowURL == "" {
			finnflowURL = "http://localhost:9090"
		}
		finnflowKey := os.Getenv("FINNFLOW_KEY")
		finnflowSecret := os.Getenv("FINNFLOW_SECRET")
		finnflowTimeout := os.Getenv("FINNFLOW_TIMEOUT")

		timeoutDuration := 30 * time.Second
		if finnflowTimeout != "" {
			if d, err := time.ParseDuration(finnflowTimeout); err == nil {
				timeoutDuration = d
			}
		}

		httpClient := &http.Client{
			Timeout: timeoutDuration,
		}

		activity := activities.NewLegacyCatalogActivity(finnflowURL, finnflowKey, finnflowSecret, httpClient)
		svc = &localTestService{activity: activity}
		fmt.Printf("Starting in PARITY TEST MODE (Connecting to %s)\n", finnflowURL)
	} else {
		// 1. Inicializar el Servicio Dummy (sin Temporal para pruebas locales rÃ¡pidas)
		svc = &dummyService{}
		fmt.Println("Starting in DUMMY MODE")
	}

	// 2. Crear el Endpoint
	ep := endpoint.MakeGetCatalogEndpoint(svc)

	// 3. Crear el HTTP Handler (Go-Kit) y registrar el codificador de errores estricto
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(handler.EncodeError),
	}
	httpHandler := httptransport.NewServer(
		ep,
		handler.DecodeGetCatalogRequest,
		handler.EncodeResponse,
		options...,
	)

	// 4. Configurar el Mux (Enrutador) para coincidir exactamente con el legado
	mux := http.NewServeMux()
	
	// 4. Configurar el Mux (Enrutador)
	mux.Handle("POST /v1/bfcl/mortgage-loan/catalogs/{catalog}", httpHandler)

	// Servir Swagger UI y OpenAPI spec
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVER_PORT") // Spring Boot usually reads this
		if port == "" {
			port = "8080" // Fallback to application.yml default
		}
	}

	fmt.Printf("Servidor Mock iniciado en puerto %s. Listo para pruebas en Postman...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

