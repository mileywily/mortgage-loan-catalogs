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
Bajo el principio de **Drop-In Replacement**, se violó deliberadamente el principio *Open/Closed (OCP)* en la capa de actividades para calcar el comportamiento "hardcodeado" de Java frente a catálogos especiales:
1.  **Seguros:** Se mantuvo la validación `HasPrefix(ToLower("seguros"))` para mapear el DTO especializado (`Factor`, `Tasa`).
2.  **Case Sensitivity:** Se replicó el matching estricto `== "TiposDocumentos"`. Si un cliente envía el path en minúsculas, cae sistemáticamente al ruteo genérico que recorta los `ids` nativos, emulando la omisión de `String.equals()` de Java.
3.  **JSON Key Ordering:** Se forzó el orden de llaves del serializador Jackson (Java) mediante el uso de *anonymous structs* en Go para el manejo del error `401 Unauthorized` (Messages Array).

## 4. Oportunidades de Mejora Futura (Siguiente Iteración)
Implementar el **Patrón Strategy** (o Registro de Decodificadores) en memoria durante el inicio del servidor, para reemplazar la cadena de sentencias `if/else` por cada nuevo catálogo, logrando finalmente la resolución del Open/Closed Principle.
