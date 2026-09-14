# Guía de Pruebas y Despliegue en Servidores (Sin Docker)
**Proyecto:** Microservicio de Catálogos Hipotecarios (Go)

Este documento detalla el paso a paso para el equipo de Infraestructura, DevOps o QA que necesite ejecutar, certificar y probar este microservicio en un servidor (Linux/Windows) sin utilizar contenedores Docker.

---

## 1. Despliegue para Pruebas de Integración Reales (Staging / QA)
Si el objetivo es colocar el código en un servidor de pruebas para que interactúe con el **Apigee/Finnflow Real** de su entorno de QA, sigan estos pasos:

### 1.1 Obtener el Binario Compilado (Single-File Deployment)
El desarrollador debe proveer **únicamente** el archivo ejecutable compilado nativamente para el sistema operativo del servidor (ej. `api-catalogos-linux` o `api-catalogos.exe`). Al estar hecho en Go, **no necesitan instalar Java (JRE) ni ninguna otra dependencia** en el servidor. 
*Nota Mágica:* Gracias a la directiva `go:embed`, toda la interfaz gráfica de **Swagger UI** está inyectada dentro de la memoria del propio ejecutable. ¡No necesitan copiar carpetas `docs` ni archivos anexos!

### 1.2 Configurar Variables de Entorno en el Servidor
Configuren las siguientes variables de entorno en su servidor apuntando a las credenciales y URLs reales del sistema de pruebas:

| Variable | Descripción | Ejemplo |
| :--- | :--- | :--- |
| `PORT` | Puerto donde escuchará este microservicio | `8080` |
| `USE_REAL_BACKEND` | Activa la salida hacia la red externa | `true` |
| `FINNFLOW_URL` | URL base del Finnflow de Pruebas | `https://apigee-qa.bancofalabella.cl` |
| `FINNFLOW_KEY` | Usuario de QA para Finnflow | `usuario_qa` |
| `FINNFLOW_SECRET` | Password de QA para Finnflow | `password_qa` |
| `FINNFLOW_INSECURE_SKIP_VERIFY` | Ignorar errores de certificados SSL (Típico en QA) | `true` |

### 1.3 Ejecutar el Servicio
Simplemente enciendan el binario en la terminal del servidor:
*   **Linux:** `./api-catalogos-linux`
*   **Windows:** `.\api-catalogos.exe`

El servidor registrará en consola (en formato JSON) un mensaje confirmando la conexión hacia Apigee. A partir de este momento, pueden dispararle peticiones de QA usando herramientas como Postman.

---

## 2. Ejecución de Pruebas Automatizadas (Para Certificación Local/CI)
Si el equipo de pruebas desea validar que el código de Go no ha roto ninguna de las **13 reglas de negocio heredadas de Java** (Validación E2E Byte a Byte), deben hacerlo en un entorno controlado (localmente) apoyándose del Mock interno.

### 2.1 Requisito
Deben tener **Go 1.23+** instalado en su máquina (Solo para ejecutar los tests de código).

### 2.2 Pruebas Unitarias e Integración
Estas pruebas validan la lógica interna del código (sin salir a la red). Se ejecutan abriendo una terminal en la carpeta raíz del proyecto:
```bash
# Ejecutar todas las pruebas del proyecto
go test -v ./...

# Opcional: Revisar la cobertura de código
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2.3 Pruebas End-to-End (E2E) de Paridad Estricta (Usando Ejecutables .exe)
Para certificar que Go responde **exactamente** igual que el sistema legado (incluyendo manejo de minúsculas, mayúsculas, errores 401 y caídas de red), el equipo debe usar el simulador de red incluido (`mock_backend`). **No se requiere instalar Go ni Docker.**

**Paso A: Preparación (A cargo del desarrollador)**
El desarrollador debe compilar los dos ejecutables en su máquina y entregarlos al equipo de QA en una carpeta junto con el script de pruebas:
```powershell
go build -o api-catalogos.exe ./cmd/server/main.go
go build -o simulador-finnflow.exe ./cmd/mock_backend/main.go
```

**Paso B: Iniciar el Simulador Finnflow (QA - Terminal 1)**
Abran una consola PowerShell en la carpeta donde guardaron los archivos y ejecuten el simulador:
```powershell
.\simulador-finnflow.exe
# El simulador quedará escuchando en el puerto 9090
```

**Paso C: Iniciar el Microservicio Go (QA - Terminal 2)**
Abran una segunda consola PowerShell, inyecten las variables para que mire al simulador local, y enciendan el API:
```powershell
$env:PORT="8082"
$env:USE_REAL_BACKEND="true"
$env:FINNFLOW_URL="http://localhost:9090"
.\api-catalogos.exe
```

**Paso D: Ejecutar el Script de Certificación (QA - Terminal 3)**
Con ambos servidores encendidos, abran una tercera consola y ejecuten el script de pruebas de estrés y paridad. Si Windows bloquea la ejecución por políticas de seguridad, utilicen este comando explícito:
```powershell
powershell -ExecutionPolicy Bypass -File .\test_paridad_extendido.ps1
```
El script de PowerShell disparará 13 peticiones HTTP distintas hacia el archivo `.exe`, validando que el servidor de Go atrape los timeouts del simulador y devuelva la estructura JSON idéntica a la documentación corporativa heredada.
