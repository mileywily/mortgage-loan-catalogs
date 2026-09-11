package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
)

// LegacyCatalogClient defines the Anti-Corruption Layer (ACL) contract.
type LegacyCatalogClient interface {
	FetchLegacyCatalogActivity(ctx context.Context, req model.GetCatalogRequest) (interface{}, error)
}

// LegacyCatalogActivity handles API interaction with legacy Apigee/Finnflow systems
type LegacyCatalogActivity struct {
	apigeeBaseURL string
	username      string
	password      string
	httpClient    *http.Client
}

// NewLegacyCatalogActivity creates a new instance of LegacyCatalogActivity
func NewLegacyCatalogActivity(baseURL, user, pass string, client *http.Client) *LegacyCatalogActivity {
	return &LegacyCatalogActivity{
		apigeeBaseURL: baseURL,
		username:      user,
		password:      pass,
		httpClient:    client,
	}
}

// FetchLegacyCatalogActivity calls the legacy API via Apigee, handling HTTP errors
// and mapping them to domain errors expected by the Temporal Workflow.
func (a *LegacyCatalogActivity) FetchLegacyCatalogActivity(ctx context.Context, req model.GetCatalogRequest) (interface{}, error) {
	url := fmt.Sprintf("%s/api/catalogo_detail/%s", a.apigeeBaseURL, req.CatalogName)

	slog.Debug("Calling external API", "url", url, "catalog", req.CatalogName, "transactionId", req.TransactionID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-Channel", req.Channel)
	httpReq.Header.Set("X-Commerce", req.Commerce)
	httpReq.Header.Set("X-Transaction-ID", req.TransactionID)

	if a.username != "" && a.password != "" {
		httpReq.SetBasicAuth(a.username, a.password)
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") || strings.Contains(errStr, "deadline") {
			slog.Error("Network error calling external API", "catalog", req.CatalogName, "error", errStr)
			return []model.NoResultsCatalogItem{{CodRespuesta: 3, Mensaje: "Sin resultados.", Excepcion: "Ninguna"}}, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("HTTP error calling external API", "catalog", req.CatalogName, "status", resp.StatusCode)
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, temporalError("InvalidCatalogError", fmt.Sprintf("invalid catalog request: %d", resp.StatusCode))
		case http.StatusUnauthorized:
			return nil, temporalError("AuthenticationError", fmt.Sprintf("unauthorized: %d", resp.StatusCode))
		case http.StatusPaymentRequired, http.StatusNotFound:
			return nil, temporalError("CatalogNotFoundError", fmt.Sprintf("catalog not found: %d", resp.StatusCode))
		default:
			// Translates to a generic retryable error in Temporal (if status >= 500)
			return nil, fmt.Errorf("LegacySystemError: status %d", resp.StatusCode)
		}
	}

	// Parse response from upstream BFCL based on catalog type
	catalogLower := strings.ToLower(req.CatalogName)
	if strings.HasPrefix(catalogLower, "seguros") {
		// The legacy backend returns fields matching the legacy Insurance DTO
		// We define a temporary struct matching the upstream JSON to decode it, then map to Domain
		var upstreamItems []struct {
			IdentificadorSeguro       string  `json:"IdentificadorSeguro"`
			Descripcion               string  `json:"Descripcion"`
			Poliza                    string  `json:"Poliza"`
			NombreCompania            string  `json:"NombreCompania"`
			CodigoCompania            int     `json:"CodigoCompania"`
			CodigoTipoSeguro          int     `json:"CodigoTipoSeguro"`
			CorrelativoPoliza         int     `json:"CorrelativoPoliza"`
			Tasa                      float64 `json:"Tasa"`
			Factor                    float64 `json:"Factor"`
			IndicadorPolizaIndividual int     `json:"IndicadorPolizaIndividual"`
			PorValorCuota             int     `json:"PorValorCuota"`
			IndicadorPolizaExterna    int     `json:"IndicadorPolizaExterna"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&upstreamItems); err != nil {
			return nil, temporalError("UpstreamDecodeError", err.Error())
		}
		domainItems := make([]model.InsuranceCatalogItem, len(upstreamItems))
		for i, item := range upstreamItems {
			domainItems[i] = model.InsuranceCatalogItem{
				InsuranceIdentifier:  item.IdentificadorSeguro,
				Description:          item.Descripcion,
				Policy:               item.Poliza,
				CompanyName:          item.NombreCompania,
				CompanyCode:          item.CodigoCompania,
				InsuranceTypeCode:    item.CodigoTipoSeguro,
				PolicyCorrelative:    item.CorrelativoPoliza,
				Rate:                 item.Tasa,
				Factor:               item.Factor,
				IndividualPolicyFlag: item.IndicadorPolizaIndividual,
				PerQuotaValueFlag:    item.PorValorCuota,
				ExternalPolicyFlag:   item.IndicadorPolizaExterna,
			}
		}
		return domainItems, nil
	}

	if req.CatalogName == "TiposDocumentos" {
		var upstreamItems []struct {
			CodigoAdm   int    `json:"codigo_adm"`
			Descripcion string `json:"descripcion"`
			GrupoID     int    `json:"grupo_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&upstreamItems); err != nil {
			return nil, temporalError("UpstreamDecodeError", err.Error())
		}
		domainItems := make([]model.TipoDocumentoCatalogItem, len(upstreamItems))
		for i, item := range upstreamItems {
			domainItems[i] = model.TipoDocumentoCatalogItem{
				Code:        item.CodigoAdm,
				Description: item.Descripcion,
				GroupID:     item.GrupoID,
			}
		}
		return domainItems, nil
	}

	if req.CatalogName == "Comunas" {
		var upstreamItems []struct {
			CodigoAdm   string `json:"codigo_adm"`
			Descripcion string `json:"descripcion"`
			RegionID    int    `json:"region_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&upstreamItems); err != nil {
			return nil, temporalError("UpstreamDecodeError", err.Error())
		}
		domainItems := make([]model.ComunaCatalogItem, len(upstreamItems))
		for i, item := range upstreamItems {
			domainItems[i] = model.ComunaCatalogItem{
				Code:        item.CodigoAdm,
				Description: item.Descripcion,
				RegionID:    item.RegionID,
			}
		}
		return domainItems, nil
	}

	// Default Standard Catalogs
	var upstreamItems []struct {
		CodigoAdm    string `json:"codigo_adm"`
		Descripcion  string `json:"descripcion"`
		CodRespuesta *int   `json:"codRespuesta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&upstreamItems); err != nil {
		return nil, temporalError("UpstreamDecodeError", err.Error())
	}

	// Handle "Sin resultados" format transparently
	if len(upstreamItems) == 0 || (len(upstreamItems) > 0 && upstreamItems[0].CodRespuesta != nil) {
		return []model.NoResultsCatalogItem{{CodRespuesta: 3, Mensaje: "Sin resultados.", Excepcion: "Ninguna"}}, nil
	}

	domainItems := make([]model.CatalogItem, len(upstreamItems))
	for i, item := range upstreamItems {
		domainItems[i] = model.CatalogItem{
			Code:        item.CodigoAdm,
			Description: item.Descripcion,
		}
	}

	return domainItems, nil
}

// temporalError helper func to emulate temporal.NewApplicationError for brevity.
// In actual code, use temporal.NewApplicationError(msg, errType, nil)
func temporalError(errType, message string) error {
	return fmt.Errorf("%s: %s", errType, message)
}
