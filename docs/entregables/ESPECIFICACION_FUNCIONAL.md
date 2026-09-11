# Especificación Funcional - Migración Go
**Proyecto:** Microservicio de Catálogos Hipotecarios (BFCL)

## 1. Propósito del Microservicio
Este componente de backend provee los diccionarios, catálogos paramétricos y listas de valores estáticas requeridas por los frontales y la lógica de negocio del crédito hipotecario (Ej: Listado de comunas, seguros, tipos de documentos). Actúa como intermediario o caché pasivo entre las aplicaciones cliente y el sistema core legado (Finnflow).

## 2. Documentación Activa (Swagger)
Todo el contrato del API se expone activamente mediante una interfaz web interactiva generada por Swagger. Se accede levantando el servicio y navegando a:
`/docs/index.html`

## 3. Catálogos Soportados y Comportamientos Especiales
La migración respeta el comportamiento funcional asimétrico del sistema original:

*   **Catálogo "Seguros" (Sensible):** Devuelve una estructura plana única y compleja (`Tasa`, `Factor`, `IdentificadorSeguro`).
    *   *Regla Funcional:* Se permite buscarlo indistintamente de las mayúsculas (Ej. `SegurosIncendio`, `segurosincendio`, `SEGUROS...`).
*   **Catálogos Nativos Estrictos ("TiposDocumentos", "Comunas"):** Devuelven atributos únicos extendidos como `grupo_id` o `region_id` respectivamente, y formatean el campo `codigo_adm` como número entero (`1`).
    *   *Regla Funcional:* Deben escribirse **exactamente** con ese casing (PascalCase). Si se escriben en minúsculas, el sistema descarta las llaves adicionales y formatea el `codigo_adm` como texto (`"1"`).
*   **Catálogos Genéricos:** Cualquier otro catálogo ("Destino", "Regiones", etc.) o variaciones ortográficas se resuelven bajo el modelo `CatalogItem` estándar (`codigo_adm` (string) y `descripcion` (string)).

## 4. Fallbacks Legados
Si el sistema central legado experimenta problemas (timeout de red), o si el catálogo de origen no tiene información pero el legacy devuelve el error `codRespuesta: 3`, este microservicio "traga" la excepción de red y entrega exitosamente al cliente frontal un JSON amigable indicando `"Sin Resultados"`. Esto previene caídas masivas en cascada en la capa de UI.
