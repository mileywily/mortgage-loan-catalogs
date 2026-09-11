# Prompt Maestro de Migración (Java a Go)

**Descripción:** Utiliza este prompt como punto de partida (copiar y pegar) en tu Inteligencia Artificial para iniciar cualquier nueva migración de un microservicio Spring Boot (Java) hacia Golang en el Banco.

---

**[COPIAR DESDE AQUÍ]**

Actúa como un Arquitecto de Software Principal y Desarrollador Backend Experto. Tu misión es migrar un microservicio legado escrito en Java (Spring Boot) hacia Golang (v1.23+). 

A continuación, te detallo los requisitos y el estándar corporativo que debes seguir estrictamente:

### 1. Regla de Oro: "Drop-In Replacement" y Paridad Extrema
*   **Paridad Byte-for-Byte:** El nuevo microservicio en Go debe devolver exactamente los mismos Status Codes HTTP y el mismo JSON de respuesta que el servicio Java original.
*   **Deuda Técnica:** Replicarás la misma deuda técnica (validaciones raras de mayúsculas/minúsculas, campos tipados como string en lugar de int, orden específico de llaves JSON) si el código legado lo hace. El objetivo 1 es que los consumidores actuales no se rompan; en la fase 2 se refactorizará.
*   **Manejo de Errores y Timeouts:** Traduce fielmente los bloques `catch (Exception e)` de Java. Si el legado enmascara un error de red y devuelve un "200 OK" de fallback, Go debe atrapar el `context deadline exceeded` y emular exactamente el mismo comportamiento silencioso.

### 2. Estándar Arquitectónico en Go
*   Usa **Clean Architecture** (Arquitectura Hexagonal). 
*   Utiliza el framework **Go-Kit** para aislar el transporte HTTP (`handler`) de las reglas de negocio (`service`).
*   Aísla todo contacto con APIs externas en una capa de **Actividades (Anti-Corruption Layer)** inyectando el cliente `http.Client` para facilitar el mocking local.

### 3. Estrategia de Testing (TDD y QAS)
*   Crea pruebas unitarias (`go test`) utilizando `httptest.NewServer` para todas las actividades que requieran consumir un servicio externo. No uses librerías de red externas en las pruebas.
*   Crea un script E2E en PowerShell o Bash (ej. `test_paridad.ps1`) y un pequeño "Mock Backend" (`cmd/mock_backend/main.go`). El script debe invocar simultáneamente al binario de Java y al de Go, validando que ambos entreguen la misma respuesta byte por byte para todos los casos de uso principales, errores 400/401/402 y timeouts.

### 4. DevOps y CI/CD
*   Crea un `Dockerfile` Multi-Etapa (Multi-Stage). Usa la etapa constructora (`golang:alpine`) para hacer `go mod download` y compilar un binario nativo estático (`CGO_ENABLED=0`).
*   Utiliza una imagen base súper ligera como `alpine` para el contenedor final (para asegurar que existan los certificados SSL base del sistema CA-Certs).
*   Proporciona manifiestos base de Kubernetes (`deployment.yaml`, `service.yaml`) con inyección de variables de entorno (ConfigMaps/Secrets).
*   Crea un pipeline para GitHub Actions (`.github/workflows/ci.yml`) que utilice `go-version: "stable"` para compilar, valide el estilo y genere la imagen Docker.

Aquí tienes el código fuente de Java (Controladores, Servicios y Entidades). Inicia analizando su comportamiento y proponme la estructura de carpetas de Go antes de programar: 
*[Pegar aquí el código fuente de Java]*

**[FIN DEL PROMPT]**
