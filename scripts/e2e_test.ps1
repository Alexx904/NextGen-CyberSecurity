# ==============================================================================
# NextGen CyberSecurity Platform - E2E Integration Test Suite
# SDLC Milestone 1 - TASK-32 Validation Script
# ==============================================================================

[CmdletBinding()]
param (
    [string]$BackendUrl  = "http://localhost:8080",
    [string]$FrontendUrl = "http://localhost:3000",
    [string]$Email       = "admin@nextgen.local",
    [string]$Password    = "AdminPass123!",
    [int]$ScanTimeoutSec = 120
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

Write-Host "==========================================================" -ForegroundColor Magenta
Write-Host "  NextGen CyberSecurity - Collaudo E2E Full Stack (Docker)  " -ForegroundColor Magenta
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

try {
    # --------------------------------------------------------------------------
    # 1. Healthcheck Backend Diretto (Port 8080)
    # --------------------------------------------------------------------------
    Write-Step "1. Verifica Healthcheck Backend Diretto ($BackendUrl/api/v1/health)"
    $healthResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/health" -Method Get
    Assert-Condition ($healthResp.status -eq "ok" -or $healthResp.status -eq "healthy" -or $healthResp.app -ne $null) `
        "Backend healthcheck OK (App: $($healthResp.app), Versione: $($healthResp.version))" `
        "Risposta healthcheck non valida"

    # --------------------------------------------------------------------------
    # 2. Healthcheck Frontend Nginx Reverse Proxy (Port 3000 -> 8080)
    # --------------------------------------------------------------------------
    Write-Step "2. Verifica Nginx Reverse Proxy ($FrontendUrl/api/v1/health)"
    $proxyHealthResp = Invoke-RestMethod -Uri "$FrontendUrl/api/v1/health" -Method Get
    Assert-Condition ($proxyHealthResp.app -ne $null) `
        "Nginx Reverse Proxy operativo e trasparente verso /api/v1/health" `
        "Nginx Reverse Proxy non instrada correttamente le chiamate API"

    # --------------------------------------------------------------------------
    # 3. Erogazione SPA Frontend (Port 3000)
    # --------------------------------------------------------------------------
    Write-Step "3. Verifica Erogazione Frontend Static SPA ($FrontendUrl/)"
    $spaResp = Invoke-WebRequest -Uri "$FrontendUrl/" -Method Get -UseBasicParsing
    Assert-Condition ($spaResp.StatusCode -eq 200 -and $spaResp.Content -match "<div id=`"root`"") `
        "Frontend React SPA erogato con successo da Nginx (HTTP 200, index.html)" `
        "Frontend SPA non raggiungibile o contenuto HTML inatteso"

    # --------------------------------------------------------------------------
    # 4. Autenticazione SuperAdmin & Recupero Token JWT
    # --------------------------------------------------------------------------
    Write-Step "4. Autenticazione Amministratore (POST $BackendUrl/api/v1/auth/login)"
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

    $authHeaders = @{
        "Authorization" = "Bearer $accessToken"
        "Content-Type"  = "application/json"
    }

    # --------------------------------------------------------------------------
    # 5. Verifica Tenant Settings & Profilo Autenticato
    # --------------------------------------------------------------------------
    Write-Step "5. Verifica Tenant Multi-Tenant (GET $BackendUrl/api/v1/tenants/current)"
    $tenantResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/tenants/current" `
        -Method Get `
        -Headers $authHeaders

    $tenantData = if ($tenantResp.data) { $tenantResp.data } else { $tenantResp }
    Assert-Condition ($tenantData.id -ne $null -and $tenantData.name -ne $null) `
        "Tenant primario identificato: $($tenantData.name) (ID: $($tenantData.id))" `
        "Recupero dati tenant fallito"

    # --------------------------------------------------------------------------
    # 6. Registrazione Target di Collaudo
    # --------------------------------------------------------------------------
    Write-Step "6. Registrazione Target Bersaglio (POST $BackendUrl/api/v1/targets)"
    $targetBody = @{
        value          = "127.0.0.2"
        description    = "E2E Automated Localhost Target"
        is_blacklisted = $false
    } | ConvertTo-Json

    $targetResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets" `
        -Method Post `
        -Headers $authHeaders `
        -Body $targetBody

    $targetData = if ($targetResp.data) { $targetResp.data } else { $targetResp }
    Assert-Condition ($targetData.id -ne $null -and $targetData.value -eq "127.0.0.2") `
        "Target creato con successo: $($targetData.value) [Tipo: $($targetData.type)] (ID: $($targetData.id))" `
        "Creazione target fallita"

    $targetId = $targetData.id

    # --------------------------------------------------------------------------
    # 7. Selezione Profilo di Scansione
    # --------------------------------------------------------------------------
    Write-Step "7. Recupero Profili di Scansione Disponibili (GET $BackendUrl/api/v1/scan-profiles)"
    $profilesResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scan-profiles" `
        -Method Get `
        -Headers $authHeaders

    $profilesData = if ($profilesResp.data) { $profilesResp.data } else { $profilesResp }
    $profilesList = if ($profilesData.items) { $profilesData.items } else { $profilesData }
    Assert-Condition ($profilesList.Count -gt 0) `
        "Trovati $($profilesList.Count) profili di scansione configurati" `
        "Nessun profilo di scansione trovato"

    # Cerchiamo un profilo rapido (Fast Discovery) altrimenti prendiamo il primo
    $selectedProfile = $profilesList | Where-Object { $_.name -match "Fast" } | Select-Object -First 1
    if (-not $selectedProfile) {
        $selectedProfile = $profilesList[0]
    }
    Write-Pass "Profilo selezionato per il test: '$($selectedProfile.name)' (Nmap args: $($selectedProfile.nmap_args))"

    # --------------------------------------------------------------------------
    # 8. Lancio Scansione On-Demand
    # --------------------------------------------------------------------------
    Write-Step "8. Avvio Scansione On-Demand (POST $BackendUrl/api/v1/scans)"
    $scanBody = @{
        target_id  = $targetId
        profile_id = $selectedProfile.id
    } | ConvertTo-Json

    $scanResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans" `
        -Method Post `
        -Headers $authHeaders `
        -Body $scanBody

    $scanData = if ($scanResp.data) { $scanResp.data } else { $scanResp }
    Assert-Condition ($scanData.id -ne $null) `
        "Scansione orchestrata con successo (ID: $($scanData.id), Stato iniziale: $($scanData.status))" `
        "Avvio scansione fallito"

    $scanId = $scanData.id

    # --------------------------------------------------------------------------
    # 9. Polling Avanzamento Scansione fino a COMPLETED
    # --------------------------------------------------------------------------
    Write-Step "9. Polling Avanzamento Scansione fino a completamento (Timeout: ${ScanTimeoutSec}s)"
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $scanCompleted = $false
    $lastStatus = ""

    while ($stopwatch.Elapsed.TotalSeconds -lt $ScanTimeoutSec) {
        $pollResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans/$scanId" `
            -Method Get `
            -Headers $authHeaders

        $currentScan = if ($pollResp.data) { $pollResp.data } else { $pollResp }
        $status = $currentScan.status
        if ($status -ne $lastStatus) {
            Write-Host "  -> Stato scansione: $status (trascorsi $([math]::Round($stopwatch.Elapsed.TotalSeconds, 1))s)" -ForegroundColor Yellow
            $lastStatus = $status
        }

        if ($status -eq "COMPLETED") {
            $scanCompleted = $true
            break
        } elseif ($status -eq "FAILED" -or $status -eq "CANCELLED") {
            $errDetail = if ($currentScan.error_message) { $currentScan.error_message } else { "Nessun messaggio di dettaglio" }
            throw "La scansione è terminata in stato non atteso: $status. Dettaglio: $errDetail"
        }

        Start-Sleep -Seconds 2
    }
    $stopwatch.Stop()

    Assert-Condition ($scanCompleted -eq $true) `
        "Scansione completata con successo in $($stopwatch.Elapsed.TotalSeconds.ToString('F1'))s" `
        "Scansione scaduta o non completata entro il timeout di $ScanTimeoutSec secondi"

    # --------------------------------------------------------------------------
    # 10. Verifica Risultati & Findings
    # --------------------------------------------------------------------------
    Write-Step "10. Ispezione Risultati e Vulnerabilità Findings (GET $BackendUrl/api/v1/scans/$scanId/findings)"
    $findingsResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans/$scanId/findings" `
        -Method Get `
        -Headers $authHeaders

    $findingsData = if ($findingsResp.data) { $findingsResp.data } else { $findingsResp }
    $findingsCount = 0
    if ($findingsData.items) {
        $findingsCount = $findingsData.items.Count
    } elseif ($findingsData -is [System.Array]) {
        $findingsCount = $findingsData.Count
    }

    Write-Pass "Recuperati $findingsCount findings/servizi per la scansione eseguita"

    # --------------------------------------------------------------------------
    # 11. Esportazione & Validazione Report PDF White-Label
    # --------------------------------------------------------------------------
    Write-Step "11. Validazione Generazione Report PDF (GET $BackendUrl/api/v1/scans/$scanId/reports/pdf)"
    $pdfTmpPath = [System.IO.Path]::GetTempFileName() + ".pdf"
    
    Invoke-WebRequest -Uri "$BackendUrl/api/v1/scans/$scanId/reports/pdf" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $accessToken" } `
        -OutFile $pdfTmpPath `
        -UseBasicParsing

    $pdfBytes = [System.IO.File]::ReadAllBytes($pdfTmpPath)
    $isPdfValid = ($pdfBytes.Length -gt 100) -and ($pdfBytes[0] -eq 0x25 -and $pdfBytes[1] -eq 0x50 -and $pdfBytes[2] -eq 0x44 -and $pdfBytes[3] -eq 0x46)
    
    Assert-Condition ($isPdfValid -eq $true) `
        "Report PDF White-Label valido generato con successo (Dimensione: $($pdfBytes.Length) bytes, Magic bytes: %PDF)" `
        "Report PDF non conforme o non generato correttamente"

    if (Test-Path $pdfTmpPath) { Remove-Item $pdfTmpPath -Force }

    # --------------------------------------------------------------------------
    # 12. Esportazione & Validazione Report CSV (RFC 4180)
    # --------------------------------------------------------------------------
    Write-Step "12. Validazione Esportazione Report CSV (GET $BackendUrl/api/v1/scans/$scanId/reports/csv)"
    $csvTmpPath = [System.IO.Path]::GetTempFileName() + ".csv"

    Invoke-WebRequest -Uri "$BackendUrl/api/v1/scans/$scanId/reports/csv" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $accessToken" } `
        -OutFile $csvTmpPath `
        -UseBasicParsing

    $csvContent = [System.IO.File]::ReadAllText($csvTmpPath)
    $hasCsvHeaders = ($csvContent -match "Host IP" -or $csvContent -match "Port" -or $csvContent -match "Severity" -or $csvContent -match "host_ip")

    Assert-Condition ($csvContent.Length -gt 10 -and $hasCsvHeaders) `
        "Report CSV conforme RFC 4180 generato con successo (Dimensione: $($csvContent.Length) bytes)" `
        "Report CSV non conforme o privo di intestazioni tabellari"

    if (Test-Path $csvTmpPath) { Remove-Item $csvTmpPath -Force }

    # --------------------------------------------------------------------------
    # 13. Pulizia Risorse di Test
    # --------------------------------------------------------------------------
    Write-Step "13. Pulizia Target di Collaudo (DELETE $BackendUrl/api/v1/targets/$targetId)"
    $deleteResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets/$targetId" `
        -Method Delete `
        -Headers $authHeaders
    $delData = if ($deleteResp.data) { $deleteResp.data } else { $deleteResp }
    $delMsg = if ($delData.message) { $delData.message } else { "Target eliminato" }
    Write-Pass "Target di test eliminato con successo: $delMsg"

    # --------------------------------------------------------------------------
    # RIEPILOGO FINALE
    # --------------------------------------------------------------------------
    Write-Host "`n==========================================================" -ForegroundColor Green
    Write-Host "  COLLAUDO INTEGRATO E2E FULL STACK COMPLETATO CON SUCCESSO! " -ForegroundColor Green
    Write-Host "  Controlli Superati: $passedChecks / $totalChecks                         " -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
    exit 0

} catch {
    Write-Host "`n==========================================================" -ForegroundColor Red
    Write-Host "  ERRORE DURANTE IL COLLAUDO INTEGRATO E2E:                " -ForegroundColor Red
    Write-Host "  $_" -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Red
    exit 1
}

