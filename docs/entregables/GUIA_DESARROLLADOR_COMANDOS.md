# Guía Práctica de Comandos (Runbook del Desarrollador)
**Proyecto:** Microservicio de Catálogos Hipotecarios (Go)

Este documento es una referencia rápida (*Cheat Sheet*) para cualquier desarrollador que clone este repositorio y necesite compilar, probar y monitorear el microservicio localmente en Windows (PowerShell) o Linux/macOS.

---

## 1. Pruebas y Certificación de Calidad (Testing)

El proyecto cuenta con una sólida pirámide de pruebas. Go facilita la ejecución sin necesitar librerías externas complejas.

### 1.1 Pruebas Unitarias e Integración (Go Test)
Valida la lógica interna del negocio y la comunicación simulada con el legado sin levantar contenedores.
```powershell
# Ejecutar todas las pruebas del proyecto de forma verbosa
go test -v ./...

# Ejecutar las pruebas calculando el porcentaje de cobertura de código
go test -coverprofile=coverage.out ./...

# Ver el reporte de cobertura gráficamente en el navegador
go tool cover -html=coverage.out
```

### 1.2 Pruebas E2E de Paridad Extrema (Script)
El repositorio incluye un script de PowerShell que realiza 13 pruebas automatizadas contra el servidor levantado, asegurando que cumple byte a byte con las reglas de negocio.
```powershell
# Primero, levanta el mock backend y el microservicio localmente (Ver sección 2)
# Luego, en otra consola, ejecuta la matriz de paridad saltando la restricción de Windows:
powershell -ExecutionPolicy Bypass -File .\test_paridad_extendido.ps1
```

---

## 2. Desarrollo y Compilación Local (Cross-Compilation)
Gracias a la directiva `go:embed`, el Swagger UI se inyecta directamente en el archivo compilado. No necesitas instalar Docker para generar artefactos listos para QA/Producción.

### 2.1 Ejecutar en Vivo
```powershell
go run ./cmd/server/main.go
```

### 2.2 Compilación Cruzada (Generar ejecutables para cualquier OS)
```powershell
# Compilar para Windows
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -o api-catalogos.exe ./cmd/server/main.go

# Compilar para Mac M1/M2/M3 (Apple Silicon)
$env:GOOS="darwin"; $env:GOARCH="arm64"
go build -o api-catalogos-mac-arm ./cmd/server/main.go

# Compilar para Servidores Linux
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -o api-catalogos-linux ./cmd/server/main.go
```

---

## 3. Gestión de Contenedores (Docker)

El empaquetado del proyecto utiliza Docker Multi-Stage.

### 3.1 Construir la Imagen (Build)
Empaqueta el binario estático y la documentación de Swagger en una imagen ultraligera de Alpine.
```powershell
docker build -t gcr.io/tu-proyecto/mortgage-loan-catalogs:latest .
```

### 3.2 Levantar el Ecosistema Localmente (Run)
Para emular el entorno de Producción, primero debemos levantar el *Mock Backend* (que simula a Apigee/Finnflow) y luego el microservicio conectándolo a dicho Mock.

```powershell
# 1. Levantar el Mock Backend (Ocupará el puerto 9090)
# (Ejecutar en una ventana de PowerShell separada y dejar corriendo)
go run ./cmd/mock_backend/main.go

# 2. Levantar el Contenedor Docker de Go (Ocupará el puerto 8083)
# Se le inyectan las variables de entorno para que sepa dónde está el Mock
docker run -d `
  --name catalogos-go `
  -p 8083:8080 `
  -e USE_REAL_BACKEND="true" `
  -e FINNFLOW_URL="http://host.docker.internal:9090" `
  -e FINNFLOW_KEY="user" `
  -e FINNFLOW_SECRET="pass" `
  gcr.io/tu-proyecto/mortgage-loan-catalogs:latest
```

### 3.3 Detener y Borrar el Contenedor
```powershell
# Forzar el borrado del contenedor activo
docker rm -f catalogos-go

# Listar todas las imágenes en tu máquina para verificar tamaños
docker images | Select-String "mortgage-loan-catalogs"
```

---

## 4. Observabilidad y Trazabilidad (Logs JSON)

El microservicio utiliza `log/slog` para escupir logs 100% compatibles con Splunk/Datadog. Cada petición atrapa el `X-Transaction-ID` del cliente.

### 4.1 Disparar una petición de prueba
Envía una petición HTTP POST usando PowerShell, inyectando cabeceras de trazabilidad corporativa:
```powershell
Invoke-WebRequest -Method Post `
  -Uri "http://localhost:8083/v1/bfcl/mortgage-loan/catalogs/Destino" `
  -Headers @{
      "Content-Type"="application/json"
      "X-Channel"="APP"
      "X-Commerce"="FALABELLA"
      "X-Transaction-ID"="TRACE-LOCAL-999"
  } | Select-Object StatusCode, Content
```

### 4.2 Visualizar la Observabilidad en Vivo
Para ver la reacción del contenedor y extraer el JSON estructurado (donde verás tu `"transactionId":"TRACE-LOCAL-999"` y el tiempo de respuesta):
```powershell
# Ver los últimos logs y quedarse escuchando (-f o follow)
docker logs -f catalogos-go
```

**Salida Esperada:**
Deberías ver una estructura JSON limpia como esta:
```json
{"time":"2026-09-10T...","level":"DEBUG","message":"Calling external API","url":"...","catalog":"Destino","transactionId":"TRACE-LOCAL-999"}
```
