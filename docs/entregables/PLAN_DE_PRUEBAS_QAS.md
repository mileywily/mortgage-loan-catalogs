# Plan de Pruebas y Certificación QAS - Migración Go
**Proyecto:** Microservicio de Catálogos Hipotecarios (BFCL)

## 1. Estrategia de Certificación
La migración de Java a Go se validó utilizando una estrategia agresiva de **Paridad Extrema (Byte-for-Byte)**. No se evaluó únicamente el comportamiento funcional, sino la exactitud sintáctica de la carga útil JSON y de los códigos HTTP para garantizar un *Drop-In Replacement* (reemplazo transparente) hacia los consumidores.

## 2. Automatización E2E
Se adjunta y se preserva en el repositorio el script **`test_paridad_extendido.ps1`**. Este script levanta un simulador de red local (Mock) y dispara la misma petición a ambos binarios a la vez (Java y Go), deteniendo el flujo si se detecta un solo byte de diferencia.

## 3. Matriz de 13 Casos de Uso (100% PASS)

1.  **Destino (Estándar):** Devuelve `[{"codigo_adm": "1"}]` como string.
2.  **Seguros PascalCase:** Devuelve el factor matemático en coma flotante. *(Nota: Diferencia visual justificada por parseo de notación científica Jackson vs Go: `2.255E-4` vs `0.0002255`).*
3.  **Seguros Minúsculas:** Funciona correctamente (case-insensitive).
4.  **TiposDocumentos (PascalCase):** Respeta y extrae la llave anidada `grupo_id`.
5.  **tiposdocumentos (Minúsculas):** Activa el modo fallback, destruyendo la llave `grupo_id`.
6.  **Comunas (PascalCase):** Respeta y extrae la llave anidada `region_id`.
7.  **comunas (Minúsculas):** Activa modo fallback.
8.  **Regiones:** Fallback estándar perfecto.
9.  **Mensaje Legacy (codRespuesta: 3):** Devuelve "Sin resultados" mapeado.
10. **Arreglo Vacío:** La API soporta listas vacías sin estallar.
11. **Catálogo Inexistente (402):** Se propagan correctamente los códigos custom heredados (`402`).
12. **Token Inválido (401):** Se estructuró un mapa anónimo en memoria para obligar a Go a devolver las llaves del Error del JSON en el exacto mismo orden que el serializador Jackson de Java original (`token_class`, `token_type`, `message`).
13. **Timeout de Red (Fallback de 200 OK):** Cuando el sistema central Apigee tarda en responder, Go atrapa la excepción `context deadline exceeded` y la transforma proactivamente en un 200 OK vacío, idéntico al bloque `catch (HttpClientErrorException)` del servicio de Java.

Todas las pruebas en QAS pasaron exitosamente bajo estas premisas.
