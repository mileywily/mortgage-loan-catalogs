package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/catalogo_detail/", func(w http.ResponseWriter, r *http.Request) {
		catalog := r.URL.Path[len("/api/catalogo_detail/"):]
		w.Header().Set("Content-Type", "application/json")

		// Manejo especial según requerimientos extendidos
		if catalog == "Inexistente" || catalog == "Error" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"mensaje": "No se encontró el catálogo solicitado"}`))
			return
		}

		if catalog == "TokenInvalido" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"detail": "Given token not valid"}`))
			return
		}

		if catalog == "Timeout" {
			// Simular un delay mayor al timeout del cliente (30s)
			// Para propósitos prácticos de la prueba, podemos pausar 2s
			// pero tenemos que asegurar que el TIMEOUT del cliente se baje a 1s
			// O podemos simplemente cerrar la conexión abruptamente para simular caída
			time.Sleep(3 * time.Second) // Assuming we will pass TIMEOUT=1s to the test
			w.WriteHeader(http.StatusOK)
			return
		}

		catalogLower := strings.ToLower(catalog)
		if strings.HasPrefix(catalogLower, "seguros") {
			w.WriteHeader(http.StatusOK)
			// Tasa and Factor with scientific notation to match Rule 1!
			w.Write([]byte(`[{"IdentificadorSeguro":"PARITY-000","Descripcion":"SEGURO PARITY","Poliza":"P-000","NombreCompania":"PARITY CIA","CodigoCompania":99,"CodigoTipoSeguro":1,"CorrelativoPoliza":1,"Tasa":0.1,"Factor":2.255E-4,"IndicadorPolizaIndividual":0,"PorValorCuota":0,"IndicadorPolizaExterna":0}]`))
			return
		}

		if catalog == "TiposDocumentos" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"codigo_adm":1,"descripcion":"Escritura Parity Test","grupo_id":10}]`))
			return
		}

		if catalog == "Comunas" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"codigo_adm":"13101","descripcion":"Santiago Centro","region_id":13}]`))
			return
		}
		
		if catalog == "Regiones" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"codigo_adm":"13","descripcion":"Metropolitana de Santiago"}]`))
			return
		}

		if catalog == "Vacio" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}
		
		if catalog == "SinResultados" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"codRespuesta":3,"Mensaje":"Sin resultados.","Excepcion":"Ninguna"}]`))
			return
		}

		// Fallback para case-insensitivity tests (tiposdocumentos, comunas) u otros
		w.WriteHeader(http.StatusOK)
		// Forzar codigo_adm como string para validar descarte de tipos correctos
		w.Write([]byte(`[{"codigo_adm":"1","descripcion":"Destino Parity"}]`))
	})

	fmt.Println("Mock BFCL Backend (Upstream) listening on :9090")
	log.Fatal(http.ListenAndServe(":9090", mux))
}
