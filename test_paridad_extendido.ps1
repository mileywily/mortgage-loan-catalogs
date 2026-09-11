# Script Extendido de Comparación de Paridad Java (8080) vs Go (8082)

$javaUrl = "http://localhost:8080/v1/bfcl/mortgage-loan/catalogs"
$goUrl   = "http://localhost:8082/v1/bfcl/mortgage-loan/catalogs"

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host " INICIANDO PRUEBAS DE PARIDAD EXTENDIDAS" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

function Get-HttpResponse($url, $useHeaders=$true) {
    try {
        $headers = @{}
        if ($useHeaders) {
            $headers = @{
                "X-Channel"        = "WEB"
                "X-Commerce"       = "FALABELLA"
                "X-Transaction-ID" = "123"
            }
        }
        $res = Invoke-WebRequest -Uri $url -Method POST -Headers $headers -UseBasicParsing -ErrorAction Stop
        return @{
            Status = [int]$res.StatusCode
            Body   = $res.Content.Trim()
        }
    } catch {
        if ($_.Exception.Response) {
            $resp = $_.Exception.Response
            $status = [int]$resp.StatusCode
            $stream = $resp.GetResponseStream()
            if ($null -ne $stream) {
                $reader = New-Object System.IO.StreamReader($stream, [System.Text.Encoding]::UTF8)
                $body = $reader.ReadToEnd().Trim()
            } else {
                $body = ""
            }
            if ($body -eq "" -and $_.ErrorDetails -ne $null) {
                $body = $_.ErrorDetails.Message.Trim()
            }
            return @{
                Status = $status
                Body   = $body
            }
        } else {
            return @{
                Status = 0
                Body   = "Error de conexión: $($_.Exception.Message)"
            }
        }
    }
}

function Run-Test($rule, $name, $endpoint, $useHeaders=$true) {
    Write-Host "`n[+] Probando Regla $($rule): $name" -ForegroundColor Yellow

    $java = Get-HttpResponse "$javaUrl/$endpoint" $useHeaders
    $go   = Get-HttpResponse "$goUrl/$endpoint" $useHeaders

    $statusJava = $java.Status
    $statusGo   = $go.Status
    $bodyJava   = $java.Body
    $bodyGo     = $go.Body

    if ($statusJava -eq $statusGo) {
        Write-Host "  [OK] Status Code Coincide: $statusJava" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Status Code Desalineado! Java: $statusJava | Go: $statusGo" -ForegroundColor Red
    }

    if ($bodyJava -eq $bodyGo) {
        Write-Host "  [OK] Cuerpos JSON Idénticos (Paridad Byte a Byte)" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Cuerpos JSON Diferentes!" -ForegroundColor Red
        Write-Host "   -> Java ($statusJava): $bodyJava" -ForegroundColor Gray
        Write-Host "   -> Go   ($statusGo): $bodyGo"   -ForegroundColor Gray
    }
}

Run-Test 1 "Destino (Estándar)" "Destino"
Run-Test 2 "Seguros PascalCase" "SegurosIncendio"
Run-Test 3 "Seguros Minúsculas" "segurosincendio"
Run-Test 4 "TiposDocumentos (Exact case)" "TiposDocumentos"
Run-Test 5 "tiposdocumentos (Minúsculas - Fallback genérico)" "tiposdocumentos"
Run-Test 6 "Comunas (Exact case)" "Comunas"
Run-Test 7 "comunas (Minúsculas - Fallback genérico)" "comunas"
Run-Test 8 "Regiones (Estándar genérico)" "Regiones"
Run-Test 9 "Mensaje Legacy (SinResultados)" "SinResultados"
Run-Test 10 "Arreglo Vacío (Vacio)" "Vacio"
Run-Test 11 "Catálogo Inexistente (402)" "Inexistente"
Run-Test 12 "Token Inválido (401)" "TokenInvalido"
Run-Test 13 "Timeout Upstream (Fallback de red 200)" "Timeout"

Write-Host "`n==================================================" -ForegroundColor Cyan
Write-Host " PRUEBAS EXTENDIDAS FINALIZADAS" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan

