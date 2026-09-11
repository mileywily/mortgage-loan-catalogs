# Guía de Despliegue y DevOps - Migración Go
**Proyecto:** Microservicio de Catálogos Hipotecarios (BFCL)

## 1. Contenerización y Optimización Docker
El contenedor Docker de Java heredado (~400 MB) ha sido reemplazado por una imagen Docker Multi-Etapa de Go optimizada, con un tamaño final de **~28.9 MB**.

**Estrategia Multi-Stage:**
1.  **Etapa `builder`:** Usa `golang:alpine` para descargar dependencias (`go mod download`) y compilar el binario estáticamente para Linux (`CGO_ENABLED=0`).
2.  **Etapa Final:** Usa `alpine:3.19` puro. Se incluye Alpine en lugar de `scratch` vacío para asegurar la inyección de los certificados SSL/TLS base del sistema operativo, requeridos para conexiones seguras hacia Apigee. Se copian la carpeta `/docs` (Swagger) y el binario compilado.

## 2. Variables de Entorno Requeridas
El contenedor de Go espera exactamente las mismas variables que el contenedor de Java:
*   `PORT`: Puerto HTTP donde se expondrá internamente el microservicio (Ej. 8080).
*   `USE_REAL_BACKEND`: Flag `"true"` o `"false"` para determinar si debe pegarle a la red externa.
*   `FINNFLOW_URL`: Endpoint interno hacia el sistema legado en Apigee.
*   `FINNFLOW_KEY`: Usuario (Basic Auth) para Apigee.
*   `FINNFLOW_SECRET`: Contraseña (Basic Auth) para Apigee.
*   `FINNFLOW_TIMEOUT`: (Opcional, Default 10s) Timeout de red hacia Apigee. Go lo maneja nativamente propagando contextos (`context.WithTimeout`).

## 3. Manifiestos de Kubernetes (`/k8s`)
El repositorio incluye archivos K8s estándar listos para inyección en el Pipeline CI/CD:
*   `deployment.yaml`: Configurado con 2 réplicas, Request/Limits eficientes (CPU 100m, RAM 128Mi) y lectura segura de *ConfigMaps* y *Secrets*.
*   `service.yaml`: ClusterIP para el enrutamiento interno.

## 4. Pipeline de CI (GitHub Actions)
La ruta `.github/workflows/ci.yml` automatiza la integración en cada Push o PR a `main`:
1.  Descarga y setup dinámico de Go `stable`.
2.  Resolución de dependencias y caché.
3.  Pruebas Unitarias.
4.  Construcción del binario nativo.
5.  Construcción y validación estricta del contenedor Docker en la nube.
