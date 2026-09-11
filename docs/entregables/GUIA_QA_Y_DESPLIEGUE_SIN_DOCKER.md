# Guía de Pruebas y Despliegue en Servidores (Sin Docker)
**Proyecto:** Microservicio de Catálogos Hipotecarios (Go)

Este documento detalla el paso a paso para el equipo de Infraestructura, DevOps o QA que necesite ejecutar, certificar y probar este microservicio en un servidor (Linux/Windows) sin utilizar contenedores Docker.

---

## 1. Despliegue para Pruebas de Integración Reales (Staging / QA)
Si el objetivo es colocar el código en un servidor de pruebas para que interactúe con el **Apigee/Finnflow Real** de su entorno de QA, sigan estos pasos:

### 1.1 Obtener el Binario Compilado
El desarrollador debe proveer el archivo ejecutable compilado nativamente para el sistema operativo del servidor (ej. `api-catalogos-linux` o `api-catalogos.exe`). Al estar hecho en Go, **no necesitan instalar Java (JRE) ni ninguna otra dependencia** en el servidor.

### 1.2 Configurar Variables de Entorno en el Servidor
Configuren las siguientes variables de entorno en su servidor apuntando a las credenciales y URLs reales del sistema de pruebas:

| Variable | Descripción | Ejemplo |
| :--- | :--- | :--- |
| `PORT` | Puerto donde escuchará este microservicio | `8080` |
| `USE_REAL_BACKEND` | Activa la salida hacia la red externa | `true` |
| `FINNFLOW_URL` | URL base del Finnflow de Pruebas | `https://apigee-qa.bancofalabella.cl` |
| `FINNFLOW_KEY` | Usuario de QA para Finnflow | `usuario_qa` |
| `FINNFLOW_SECRET` | Password de QA para Finnflow | `password_qa` |

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

### 2.3 Pruebas End-to-End (E2E) de Paridad Estricta
Para certificar que Go responde **exactamente** igual que el sistema legado (incluyendo manejo de minúsculas, mayúsculas, errores 401 y caídas de red), el equipo debe usar el simulador de red incluido (`mock_backend`).

**Paso A: Iniciar el Simulador Finnflow (Terminal 1)**
```bash
go run ./cmd/mock_backend/main.go
# El simulador quedará escuchando en el puerto 9090
```

**Paso B: Iniciar el Microservicio Go (Terminal 2)**
Apuntando al simulador:
```powershell
$env:PORT="8082"
$env:USE_REAL_BACKEND="true"
$env:FINNFLOW_URL="http://localhost:9090"
go run ./cmd/server/main.go
```

**Paso C: Ejecutar el Script de Certificación (Terminal 3)**
Con ambos servidores encendidos, ejecuten el script de pruebas de estrés y paridad:
```powershell
.\test_paridad_extendido.ps1
```
El script ejecutará 13 llamados HTTP distintos, validando que el servidor entregue los Status Codes correctos y la estructura JSON idéntica a la documentación corporativa heredada.
