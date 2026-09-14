# Documento de Arquitectura y Diseño - Migración Go
**Proyecto:** Microservicio de Catálogos Hipotecarios (BFCL)
**De:** Spring Boot (Java 17) **A:** Golang 1.23+

## 1. Topología y Tenencia Corporativa
El microservicio opera bajo la **Tenencia BFCL (Banco Falabella Chile)** como un dominio de backend puro. Respeta la separación del borde, exponiéndose a través del API Gateway Corporativo (Apigee) sin asumir responsabilidades del Backend For Frontend (Kong - FIF). 

## 2. Patrón Arquitectónico: Clean Architecture (Arquitectura Hexagonal)
Para asegurar alta cohesión y bajo acoplamiento, se reemplazó la estructura MVC tradicional de Spring Boot por Clean Architecture apoyada en el framework **Go-Kit**.

*   **Transport / Handler (`internal/handler`):** Se encarga netamente de la decodificación HTTP (JSON, Headers) y la serialización de respuestas/errores.
*   **Endpoint (`internal/endpoint`):** Capa de middleware donde se empaquetan los Request/Response puros de dominio.
*   **Service (`internal/service`):** Lógica de orquestación de negocio. Aislada de la red mediante interfaces estables (`CatalogService`).
*   **Activities / Anti-Corruption Layer (`internal/activities`):** Aquí reside `LegacyCatalogActivity`. Actúa como capa anticorrupción para encapsular todas las deudas técnicas del sistema heredado (Finnflow), parseando respuestas extrañas (ej. `[{"codRespuesta": 3}]`) en entidades puras de Go (`CatalogItem`).

## 3. Manejo de Deuda Técnica y Paridad Estricta
Bajo el principio de **Drop-In Replacement**, se mantuvo la paridad exacta con el comportamiento del motor Jackson (Java) mediante adaptaciones de bajo nivel:
1.  **Tipado Dinámico Tolerante (Capa Anticorrupción):** Se configuraron los DTOs intermedios con `interface{}` para el campo `codigo_adm`. Esto previene crashes por tipado estricto cuando Apigee (upstream) retorna números (`1`) en lugar de textos (`"1"`), emulando la auto-coerción silenciosa que hacía Java.
2.  **Notación Científica Emulada:** Se implementó un tipo especial (`json.Number`) con formateo manual para las llaves `Factor` y `Tasa` de los Seguros, forzando a Go a omitir ceros iniciales en el exponente (ej. `2.255E-4`) exactamente como lo hacía `Double.toString()` en Java.
3.  **Permisividad de Cabeceras:** Aunque Go captura proactivamente los headers corporativos (`X-Transaction-ID`, etc.) para trazabilidad, no rompe la petición (HTTP 200 OK) si el Front-End no los envía, respetando la validación laxa del antiguo `@RestController` de Java.
4.  **Seguros y Case Sensitivity:** Se mantuvo la validación `HasPrefix(ToLower("seguros"))` y el matching estricto `== "TiposDocumentos"`. Si un cliente envía el path en minúsculas, cae al ruteo genérico.

## 3.1 Single-File Deployment (Swagger Embebido)
Para optimizar la entrega del artefacto sin dependencias externas, se utilizó la directiva `//go:embed` de Go 1.16+. Toda la interfaz de **Swagger UI** y el archivo `openapi.yaml` residen dentro de la memoria del archivo `.exe`. El equipo de QA y Operaciones ya no necesita copiar carpetas accesorias; basta con ejecutar el binario y visitar `/docs/`.

## 4. Observabilidad y Trazabilidad (JSON Logging)
Se configuró la librería nativa `log/slog` de Go 1.21+ para emular el comportamiento del appender `logback-json-classic` del legado.
*   **Logs Estructurados:** Las salidas a STDOUT son diccionarios JSON estrictos listos para su ingesta en Splunk/Datadog.
*   **Trazabilidad Distribuida:** A diferencia del sistema legado (que ignoraba los headers), este microservicio extrae proactivamente `X-Transaction-ID`, `X-Channel` y `X-Commerce` del cliente, **los inyecta en cada JSON log de DEBUG/ERROR**, y adicionalmente los propaga hacia el sistema central (Apigee) en las cabeceras HTTP de salida.

## 5. Oportunidades de Mejora Futura (Siguiente Iteración)
Implementar el **Patrón Strategy** (o Registro de Decodificadores) en memoria durante el inicio del servidor, para reemplazar la cadena de sentencias `if/else` por cada nuevo catálogo, logrando finalmente la resolución del Open/Closed Principle.
