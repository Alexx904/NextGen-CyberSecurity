# ==============================================================================
# NextGen CyberSecurity Platform - Automated Container E2E Test Runner (PowerShell)
# Cross-Platform Windows/PowerShell Docker Orchestrator & E2E Test Suite
# ==============================================================================

[CmdletBinding()]
param (
    [string]$BackendUrl  = "http://localhost:8080",
    [string]$FrontendUrl = "http://localhost:3000",
    [string]$Email       = "admin@nextgen.local",
    [string]$Password    = "AdminPass123!",
    [int]$WaitTimeoutSec = 90,
    [switch]$KeepAlive,
    [switch]$NoCleanVolumes
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

function Write-Step {
    param([string]$Message)
    Write-Host "`n[STEP] $Message" -ForegroundColor Cyan
}

function Write-Pass {
    param([string]$Message)
    Write-Host "  [+] PASS: $Message" -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host "  [-] FAIL: $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "  [*] INFO: $Message" -ForegroundColor Yellow
}

Write-Host "==========================================================" -ForegroundColor Magenta
Write-Host "  NextGen CyberSecurity - Automated Container E2E Runner   " -ForegroundColor Magenta
Write-Host "  Stack: Docker Compose (PostgreSQL 16, Go, React/Nginx)  " -ForegroundColor Magenta
Write-Host "==========================================================" -ForegroundColor Magenta

$passedChecks = 0
$totalChecks = 0

function Assert-Condition {
    param([bool]$Condition, [string]$SuccessMsg, [string]$FailMsg)
    $script:totalChecks++
    if ($Condition) {
        Write-Pass $SuccessMsg
        $script:passedChecks++
    } else {
        Write-Fail $FailMsg
        throw "Assertion failed: $FailMsg"
    }
}

function Wait-For-Health {
    param([string]$Url, [string]$Name, [int]$TimeoutSec)
    Write-Info "In attesa che $Name sia pronto su $Url (Timeout: ${TimeoutSec}s)..."
    $elapsed = 0
    while ($elapsed -lt $TimeoutSec) {
        try {
            $resp = Invoke-RestMethod -Uri $Url -Method Get -TimeoutSec 3 -ErrorAction SilentlyContinue
            if ($resp) {
                Write-Pass "$Name e operativo e risponde correttamente!"
                return $true
            }
        } catch {
            # Attendere disponibilità
        }
        Start-Sleep -Seconds 2
        $elapsed += 2
    }
    Write-Fail "$Name non ha risposto entro $TimeoutSec secondi."
    return $false
}

try {
    # 1. Verifica prerequisiti
    Write-Step "1. Verifica prerequisiti host (Docker & Docker Compose)"
    $dockerCmd = Get-Command "docker" -ErrorAction SilentlyContinue
    Assert-Condition ($dockerCmd -ne $null) "Docker CLI presente nel PATH" "Docker non trovato nel PATH"

    $composeVersion = docker compose version 2>&1
    Assert-Condition ($LASTEXITCODE -eq 0) "Docker Compose (v2) operativo ($composeVersion)" "Docker Compose non disponibile"

    # 2. Avvio stack Docker Compose
    Write-Step "2. Avvio dello stack di container in background (postgres, backend, frontend)"
    docker compose up -d --build postgres backend frontend
    Assert-Condition ($LASTEXITCODE -eq 0) "Container avviati con successo via docker compose" "Errore durante docker compose up"

    # 3. Healthcheck e disponibilità servizi
    Write-Step "3. Healthcheck e attesa disponibilita endpoint"
    $backendReady = Wait-For-Health -Url "$BackendUrl/api/v1/health" -Name "Backend Go (Direct)" -TimeoutSec $WaitTimeoutSec
    Assert-Condition $backendReady "Backend Go operativo su $BackendUrl" "Backend Go non raggiungibile"

    $proxyReady = Wait-For-Health -Url "$FrontendUrl/api/v1/health" -Name "Frontend Nginx (Reverse Proxy)" -TimeoutSec $WaitTimeoutSec
    Assert-Condition $proxyReady "Nginx Reverse Proxy operativo su $FrontendUrl" "Nginx Reverse Proxy non raggiungibile"

    # 4. Verifica SPA HTML erogata
    Write-Step "4. Verifica erogazione HTML SPA React"
    $spaResp = Invoke-WebRequest -Uri "$FrontendUrl/" -Method Get -UseBasicParsing
    Assert-Condition ($spaResp.StatusCode -eq 200 -and $spaResp.Content -match "<div id=`"root`"") `
        "Frontend React SPA erogato con successo (HTTP 200, index.html con React root)" `
        "Contenuto SPA non conforme o index.html non accessibile"

    # 5. Autenticazione JWT
    Write-Step "5. Autenticazione SuperAdmin e acquisizione JWT Bearer Token"
    $loginBody = @{
        email    = $Email
        password = $Password
    } | ConvertTo-Json

    $loginResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody

    $loginData = if ($loginResp.data) { $loginResp.data } else { $loginResp }
    $accessToken = if ($loginData.tokens) { $loginData.tokens.access_token } else { $loginData.access_token }

    Assert-Condition ($accessToken -ne $null -and $loginData.user -ne $null) `
        "Login eseguito con successo. Utente: $($loginData.user.full_name) ($($loginData.user.role))" `
        "Login fallito: token JWT non restituito"

    $authHeader = @{ "Authorization" = "Bearer $accessToken" }

    # 6. Target Groups & Targets API
    Write-Step "6. Verifica Target Groups & Targets API"
    $groupsResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups" -Method Get -Headers $authHeader
    Assert-Condition ($groupsResp.success -eq $true) `
        "GET /api/v1/target-groups operativo con successo" `
        "Errore durante la chiamata a /target-groups"

    $targetsResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets" -Method Get -Headers $authHeader
    Assert-Condition ($targetsResp.success -eq $true) `
        "GET /api/v1/targets operativo con successo" `
        "Errore durante la chiamata a /targets"

    # 7. Scan Profiles API
    Write-Step "7. Verifica Profili di Scansione Nmap"
    $profilesResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scan-profiles" -Method Get -Headers $authHeader
    $hasPredefined = ($profilesResp.data | Where-Object { $_.name -like "*360*" -or $_.name -like "*Fast*" }) -ne $null
    Assert-Condition ($profilesResp.success -eq $true -and $hasPredefined) `
        "GET /api/v1/scan-profiles restituisce i profili certificati" `
        "Profili di scansione non conformi"

    # 8. Attack Path Graph & Chokepoints (Milestone 2 API)
    Write-Step "8. Verifica Attack Path Mapping & Chokepoints (Milestone 2 API)"
    $graphResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/attack-paths/graph" -Method Get -Headers $authHeader
    Assert-Condition ($graphResp.success -eq $true -and $graphResp.data.summary -ne $null) `
        "GET /api/v1/attack-paths/graph risponde con topologia e summary di rischio" `
        "Endpoint /attack-paths/graph non valido"

    $chokeResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/attack-paths/chokepoints" -Method Get -Headers $authHeader
    Assert-Condition ($chokeResp.success -eq $true) `
        "GET /api/v1/attack-paths/chokepoints risponde con lista di chokepoints" `
        "Endpoint /attack-paths/chokepoints non valido"

    # 9. Reports API
    Write-Step "9. Verifica Reports API"
    $reportsResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/reports" -Method Get -Headers $authHeader
    Assert-Condition ($reportsResp.success -eq $true) `
        "GET /api/v1/reports operativo con successo" `
        "Errore durante la chiamata a /reports"

    # 10. Riepilogo Finale
    Write-Step "10. Riepilogo Collaudo E2E"
    Write-Pass "Tutti i $passedChecks controlli E2E su stack containerizzato sono risultati conformi al 100%!"

} catch {
    Write-Fail "Eccezione durante l'esecuzione del runner E2E: $_"
    exit 1
} finally {
    if (-not $KeepAlive) {
        Write-Info "Arresto e pulizia dei container Docker Compose..."
        if (-not $NoCleanVolumes) {
            docker compose down -v --remove-orphans | Out-Null
        } else {
            docker compose down --remove-orphans | Out-Null
        }
        Write-Pass "Ambiente Docker Compose ripulito."
    } else {
        Write-Info "Switch -KeepAlive attivo: container lasciati in esecuzione."
    }
}

