package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/bancofalabella/mortgage-loan-catalogs/internal/endpoint"
	"github.com/bancofalabella/mortgage-loan-catalogs/internal/model"
)

var (
	ErrMissingHeaders = errors.New("missing required headers")
	ErrInvalidCatalog = errors.New("invalid catalog name")
)

// catalogItemDTO represents the JSON payload for standard catalogs.
type catalogItemDTO struct {
	CodigoAdm   string `json:"codigo_adm"`
	Descripcion string `json:"descripcion"`
}

// comunaCatalogItemDTO represents the JSON payload for comunas.
type comunaCatalogItemDTO struct {
	CodigoAdm   string `json:"codigo_adm"`
	Descripcion string `json:"descripcion"`
	RegionID    int    `json:"region_id"`
}

// tipoDocumentoCatalogItemDTO represents the JSON payload for tipos documentos.
type tipoDocumentoCatalogItemDTO struct {
	CodigoAdm   int    `json:"codigo_adm"`
	Descripcion string `json:"descripcion"`
	GrupoID     int    `json:"grupo_id"`
}

// noResultsCatalogItemDTO represents the legacy empty/network error fallback
type noResultsCatalogItemDTO struct {
	CodRespuesta int    `json:"codRespuesta"`
	Mensaje      string `json:"Mensaje"`
	Excepcion    string `json:"Excepcion"`
}

// insuranceCatalogItemDTO represents the JSON payload for insurance catalogs.
type insuranceCatalogItemDTO struct {
	IdentificadorSeguro       string  `json:"IdentificadorSeguro"`
	Descripcion               string  `json:"Descripcion"`
	Poliza                    string  `json:"Poliza"`
	NombreCompania            string  `json:"NombreCompania"`
	CodigoCompania            int     `json:"CodigoCompania"`
	CodigoTipoSeguro          int     `json:"CodigoTipoSeguro"`
	CorrelativoPoliza         int     `json:"CorrelativoPoliza"`
	Tasa                      json.Number `json:"Tasa"`
	Factor                    json.Number `json:"Factor"`
	IndicadorPolizaIndividual int         `json:"IndicadorPolizaIndividual"`
	PorValorCuota             int         `json:"PorValorCuota"`
	IndicadorPolizaExterna    int         `json:"IndicadorPolizaExterna"`
}

// DecodeGetCatalogRequest extracts and validates the headers and path variables
// from the HTTP request, returning an endpoint.GetCatalogRequestDTO.
func DecodeGetCatalogRequest(_ context.Context, r *http.Request) (interface{}, error) {
	channel := r.Header.Get("X-Channel")
	commerce := r.Header.Get("X-Commerce")
	trxID := r.Header.Get("X-Transaction-ID")

	if channel == "" || commerce == "" || trxID == "" {
		return nil, ErrMissingHeaders
	}

	// Assuming Go 1.22+ for r.PathValue
	catalogName := r.PathValue("catalog")
	if catalogName == "" {
		return nil, ErrInvalidCatalog
	}

	return endpoint.GetCatalogRequestDTO{
		CatalogName:   catalogName,
		Channel:       channel,
		Commerce:      commerce,
		TransactionID: trxID,
	}, nil
}

// EncodeResponse writes the domain response mapped to DTO as JSON into the HTTP response.
// It explicitly unwraps the Endpoint response to ensure Drop-in replacement (returning a JSON array directly).
func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	if res, ok := response.(endpoint.GetCatalogResponseDTO); ok {
		// Map Pure Domain Entities to DTOs (which contain the correct legacy json tags)
		switch data := res.Data.(type) {
		case []model.CatalogItem:
			dtos := make([]catalogItemDTO, len(data))
			for i, item := range data {
				dtos[i] = catalogItemDTO{
					CodigoAdm:   item.Code,
					Descripcion: item.Description,
				}
			}
			return json.NewEncoder(w).Encode(dtos)
		case []model.ComunaCatalogItem:
			dtos := make([]comunaCatalogItemDTO, len(data))
			for i, item := range data {
				dtos[i] = comunaCatalogItemDTO{
					CodigoAdm:   item.Code,
					Descripcion: item.Description,
					RegionID:    item.RegionID,
				}
			}
			return json.NewEncoder(w).Encode(dtos)
		case []model.TipoDocumentoCatalogItem:
			dtos := make([]tipoDocumentoCatalogItemDTO, len(data))
			for i, item := range data {
				dtos[i] = tipoDocumentoCatalogItemDTO{
					CodigoAdm:   item.Code,
					Descripcion: item.Description,
					GrupoID:     item.GroupID,
				}
			}
			return json.NewEncoder(w).Encode(dtos)
		case []model.NoResultsCatalogItem:
			dtos := make([]noResultsCatalogItemDTO, len(data))
			for i, item := range data {
				dtos[i] = noResultsCatalogItemDTO{
					CodRespuesta: item.CodRespuesta,
					Mensaje:      item.Mensaje,
					Excepcion:    item.Excepcion,
				}
			}
			return json.NewEncoder(w).Encode(dtos)
		case []model.InsuranceCatalogItem:
			dtos := make([]insuranceCatalogItemDTO, len(data))
			for i, item := range data {
				dtos[i] = insuranceCatalogItemDTO{
					IdentificadorSeguro:       item.InsuranceIdentifier,
					Descripcion:               item.Description,
					Poliza:                    item.Policy,
					NombreCompania:            item.CompanyName,
					CodigoCompania:            item.CompanyCode,
					CodigoTipoSeguro:          item.InsuranceTypeCode,
					CorrelativoPoliza:         item.PolicyCorrelative,
					Tasa:                      json.Number(formatFloatLikeJackson(item.Rate)),
					Factor:                    json.Number(formatFloatLikeJackson(item.Factor)),
					IndicadorPolizaIndividual: item.IndividualPolicyFlag,
					PorValorCuota:             item.PerQuotaValueFlag,
					IndicadorPolizaExterna:    item.ExternalPolicyFlag,
				}
			}
			return json.NewEncoder(w).Encode(dtos)
		}

		// Fallback for direct map returning (like our current dummy)
		return json.NewEncoder(w).Encode(res.Data)
	}

	return json.NewEncoder(w).Encode(response)
}

// errorResponseDTO ensures exact JSON key ordering as Java
type errorResponseDTO struct {
	Detail       *string       `json:"detail"`
	Code         *string       `json:"code"`
	Messages     []interface{} `json:"messages"`
	ErrorsDetail *string       `json:"errors_detail"`
}

// EncodeError handles errors from both the Decoder and the Endpoint layer,
// translating them strictly to the legacy Java HTTP status codes and payloads.
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	errMessage := err.Error()

	// 1. Unauthorized / Token Error -> Match TokenNotValidException
	if errors.Is(err, ErrMissingHeaders) || strings.Contains(errMessage, "AuthenticationError") {
		w.WriteHeader(http.StatusUnauthorized)

		tokenMessage := struct {
			TokenClass string `json:"token_class"`
			TokenType  string `json:"token_type"`
			Message    string `json:"message"`
		}{
			TokenClass: "AccessToken",
			TokenType:  "access",
			Message:    "Token is invalid",
		}

		d := "Given token not valid for any token type"
		c := "token_not_valid"
		json.NewEncoder(w).Encode(errorResponseDTO{
			Detail:       &d,
			Code:         &c,
			Messages:     []interface{}{tokenMessage},
			ErrorsDetail: nil,
		})
		return
	}

	// 2. Catalog Not Found -> Match CatalogNotFoundException
	if strings.Contains(errMessage, "CatalogNotFoundError") || errors.Is(err, ErrInvalidCatalog) {
		w.WriteHeader(http.StatusPaymentRequired) // 402 Legacy match
		ed := "CatÃ¡logo no encontrado"
		json.NewEncoder(w).Encode(errorResponseDTO{
			Detail:       nil,
			Code:         nil,
			Messages:     nil,
			ErrorsDetail: &ed,
		})
		return
	}

	// 3. Default Bad Request -> Match CatalogException
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResponseDTO{
		Detail:       nil,
		Code:         nil,
		Messages:     nil,
		ErrorsDetail: &errMessage,
	})
}

// formatFloatLikeJackson formats a float64 precisely as Java's Jackson/Double.toString() does:
// Switching to scientific notation for |d| < 10^-3 or |d| >= 10^7, removing the leading 0 in the exponent.
func formatFloatLikeJackson(f float64) string {
	// Let Go do standard generic formatting first
	s := strconv.FormatFloat(f, 'g', -1, 64)

	// Java Double.toString() forces exponential if value < 1e-3
	if f != 0 && f > -1e-3 && f < 1e-3 {
		s = strconv.FormatFloat(f, 'E', -1, 64)
		// Go returns 2.255E-04, Java returns 2.255E-4
		s = strings.Replace(s, "E-0", "E-", 1)
		s = strings.Replace(s, "E+0", "E", 1)
	}
	return s
}
