# Prompt Maestro de Migración (Java a Go)

**Descripción:** Utiliza este prompt como punto de partida (copiar y pegar) en tu Inteligencia Artificial para iniciar cualquier nueva migración de un microservicio Spring Boot (Java) hacia Golang en el Banco.

---

**[COPIAR DESDE AQUÍ]**

Actúa como un Arquitecto de Software Principal y Desarrollador Backend Experto. Tu misión es migrar un microservicio legado escrito en Java (Spring Boot) hacia Golang (v1.23+). 

A continuación, te detallo los requisitos y el estándar corporativo que debes seguir estrictamente:

### 1. Regla de Oro: "Drop-In Replacement" y Paridad Extrema
*   **Paridad Byte-for-Byte:** El nuevo microservicio en Go debe devolver exactamente los mismos Status Codes HTTP y el mismo JSON de respuesta que el servicio Java original.
*   **Permisividad de Cabeceras:** Si el `@RestController` original de Java no exigía cabeceras de trazabilidad (como `X-Transaction-ID`), Go no debe quebrar la petición si faltan. Extrae las cabeceras solo si existen.
*   **Tipado Dinámico Tolerante:** Jackson en Java coerce silenciosamente los números (`1`) a strings (`"1"`). Para no quebrar al recibir respuestas no tipadas desde el backend (Apigee), usa `interface{}` en los DTOs de Go (Capa Anticorrupción) y fuérzalos a string antes de devolverlos.
*   **Notación Científica Exacta:** Java convierte `Double` muy pequeños (ej. `0.0002255`) a científica de forma estricta (ej. `2.255E-4`). Si el legado usaba esa notación, implementa `json.Number` en Go combinando `strconv.FormatFloat` para emular exactamente el recorte de ceros iniciales en el exponente.
*   **Deuda Técnica General:** Replicarás cualquier otra rareza (validaciones case-sensitive, orden específico de llaves JSON en errores 401, etc.).
*   **Manejo de Errores y Timeouts:** Traduce fielmente los bloques `catch (Exception e)`. Si el legado enmascara un error de red y devuelve un "200 OK" de fallback, Go debe atrapar el `context deadline exceeded` y emular el mismo comportamiento silencioso.

### 2. Estándar Arquitectónico en Go
*   Usa **Clean Architecture** (Arquitectura Hexagonal). 
*   Utiliza el framework **Go-Kit** para aislar el transporte HTTP (`handler`) de las reglas de negocio (`service`).
*   Aísla todo contacto con APIs externas en una capa de **Actividades (Anti-Corruption Layer)** inyectando el cliente `http.Client` para facilitar el mocking local.
*   **Observabilidad Nativa:** Utiliza la librería estándar `log/slog` configurada con `slog.NewJSONHandler` para que todos los logs se emitan en formato JSON estructurado, reemplazando así a `logback-json` de Java. Extrae siempre headers transaccionales (Ej. `X-Transaction-ID`) e inyéctalos como llaves nativas en los logs de error/debug.

### 3. Estrategia de Testing (TDD y QAS)
*   Crea pruebas unitarias (`go test`) utilizando `httptest.NewServer` para todas las actividades que requieran consumir un servicio externo. No uses librerías de red externas en las pruebas.
*   Crea un script E2E en PowerShell o Bash (ej. `test_paridad.ps1`) y un pequeño "Mock Backend" (`cmd/mock_backend/main.go`). El script debe invocar simultáneamente al binario de Java y al de Go, validando que ambos entreguen la misma respuesta byte por byte para todos los casos de uso principales, errores 400/401/402 y timeouts.

### 4. DevOps, Despliegue y CI/CD
*   **Single-File Deployment (Swagger):** Utiliza la directiva `//go:embed` de Go 1.16+ para inyectar todo el Swagger UI (HTML/YAML) dentro del ejecutable. QA no debe necesitar copiar carpetas externas.
*   **Entornos de QA (Certificados):** Agrega una bandera o variable de entorno (ej. `INSECURE_SKIP_VERIFY=true`) que desactive la validación SSL del `http.Client` para poder probar contra servidores corporativos que usan certificados autofirmados (reemplaza a `trust-all-certs` de Java).
*   Crea un `Dockerfile` Multi-Etapa (Multi-Stage) usando `golang:alpine` y compila un binario nativo estático (`CGO_ENABLED=0`).
*   Proporciona manifiestos base de Kubernetes (`deployment.yaml`, `service.yaml`) con inyección de variables de entorno.

Aquí tienes el código fuente de Java (Controladores, Servicios y Entidades). Inicia analizando su comportamiento y proponme la estructura de carpetas de Go antes de programar: 
*[Pegar aquí el código fuente de Java]*

**[FIN DEL PROMPT]**
