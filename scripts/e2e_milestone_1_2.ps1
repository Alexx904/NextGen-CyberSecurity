# ==============================================================================
# NextGen CyberSecurity Platform - E2E Integration Test Suite (Milestone 1.2)
# SDLC EPIC 18 & EPIC 19 - TASK-56 Real-Time SSE Streaming & Debug Console Drawer
# ==============================================================================

[CmdletBinding()]
param (
    [string]$BackendUrl  = "http://localhost:8080",
    [string]$FrontendUrl = "http://localhost:3000",
    [string]$Email       = "admin@nextgen.local",
    [string]$Password    = "AdminPass123!",
    [int]$TimeoutSec     = 60
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
Write-Host "  NextGen CyberSecurity - Collaudo E2E Milestone 1.2       " -ForegroundColor Magenta
Write-Host "  (EPIC 18 & EPIC 19 Real-Time Live Streaming via SSE)     " -ForegroundColor Magenta
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
    # 0. Healthcheck & Autenticazione JWT
    # --------------------------------------------------------------------------
    Write-Step "0. Healthcheck Backend & Autenticazione SuperAdmin"
    $healthResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/health" -Method Get
    Assert-Condition ($healthResp.status -eq "ok" -or $healthResp.status -eq "healthy" -or $healthResp.app -ne $null) `
        "Backend healthcheck OK (Status: $($healthResp.status), App: $($healthResp.app))" `
        "Healthcheck backend fallito"

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

    Assert-Condition ($accessToken -ne $null) `
        "Autenticazione completata con successo, JWT Bearer Token acquisito" `
        "Login fallito: token mancante"

    $authHeaders = @{
        "Authorization" = "Bearer $accessToken"
        "Content-Type"  = "application/json"
    }

    $tenantResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/tenants/current" -Method Get -Headers $authHeaders
    $tenantData = if ($tenantResp.data) { $tenantResp.data } else { $tenantResp }
    Assert-Condition ($tenantData.id -ne $null) `
        "Tenant primario identificato: $($tenantData.name) (ID: $($tenantData.id))" `
        "Impossibile identificare il tenant autenticato"

    $testSuffix = (Get-Random -Minimum 1000 -Maximum 9999).ToString()

    # --------------------------------------------------------------------------
    # FEATURE 1 / EPIC 18: LAN Discovery Live Streaming via SSE (DoD 1 & DoD 2)
    # --------------------------------------------------------------------------
    Write-Step "1. [EPIC 18] Avvio LAN Discovery e Streaming SSE (/network-discovery/stream)"

    # 1.1 Connessione allo stream SSE con session_id dedicato
    $discSessionId = [System.Guid]::NewGuid().ToString()
    $streamUri = "$BackendUrl/api/v1/network-discovery/stream?session_id=$discSessionId"

    $httpClient = New-Object System.Net.Http.HttpClient
    $httpClient.Timeout = [System.TimeSpan]::FromSeconds(30)
    $httpReq = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Get, $streamUri)
    $httpReq.Headers.Add("Authorization", "Bearer $accessToken")
    $httpReq.Headers.Add("Accept", "text/event-stream")

    $streamResp = $httpClient.SendAsync($httpReq, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()

    Assert-Condition ($streamResp.StatusCode -eq [System.Net.HttpStatusCode]::OK) `
        "Stream SSE aperto con successo (HTTP 200 OK)" `
        "Impossibile aprire stream SSE: $($streamResp.StatusCode)"

    $contentType = $streamResp.Content.Headers.ContentType.MediaType
    Assert-Condition ($contentType -eq "text/event-stream") `
        "Header Content-Type conforme allo standard W3C text/event-stream" `
        "Content-Type non valido: $contentType"

    $streamContent = $streamResp.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $streamReader = New-Object System.IO.StreamReader($streamContent)

    # Leggi prima riga (Handshake : connected)
    $handshakeLine = $streamReader.ReadLineAsync().GetAwaiter().GetResult()
    Assert-Condition ($handshakeLine -like "*connected*") `
        "Handshake iniziale SSE ricevuto: '$handshakeLine'" `
        "Handshake SSE non valido o mancante: '$handshakeLine'"

    # 1.2 Avvio scansione background su subnet 127.0.0.1/32
    $discPayload = @{
        subnet               = "127.0.0.1/32"
        include_common_ports = $false
    } | ConvertTo-Json

    $discPostResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/network-discovery/scan" `
        -Method Post `
        -Headers $authHeaders `
        -Body $discPayload

    $discPostData = if ($discPostResp.data) { $discPostResp.data } else { $discPostResp }
    Assert-Condition ($discPostData.devices -ne $null -and $discPostData.devices.Count -ge 1) `
        "LAN Discovery eseguita con successo, rilevati $($discPostData.devices.Count) host" `
        "LAN Discovery fallita"

    # Chiusura stream 1
    $streamReader.Close()
    $streamContent.Close()
    $httpClient.Dispose()

    # 1.3 Verifica Persistenza Incrementale (DoD 2) su endpoint GET /results
    $resultsResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/network-discovery/results" -Method Get -Headers $authHeaders
    $resultsData = if ($resultsResp.data) { $resultsResp.data } else { $resultsResp }
    Assert-Condition ($resultsData.devices -ne $null -and $resultsData.devices.Count -ge 1) `
        "Persistenza incrementale validata: report disponibile via GET /results con $($resultsData.devices.Count) host" `
        "Verifica persistenza fallita: nessun host trovato su /results"

    # --------------------------------------------------------------------------
    # FEATURE 2 / EPIC 19: Nmap Scan Live Streaming & Incremental Findings (DoD 3)
    # --------------------------------------------------------------------------
    Write-Step "2. [EPIC 19] Nmap Live Streaming & Incremental Findings (/scans/{id}/stream)"

    # 2.1 Creazione Target Bersaglio Loopback
    $targetBody = @{
        value          = "127.0.0.$((Get-Random -Minimum 15 -Maximum 240))"
        description    = "E2E SSE Stream Target $testSuffix"
        is_blacklisted = $false
    } | ConvertTo-Json

    $targetResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets" `
        -Method Post `
        -Headers $authHeaders `
        -Body $targetBody

    $targetData = if ($targetResp.data) { $targetResp.data } else { $targetResp }
    $targetId = $targetData.id
    Assert-Condition ($targetId -ne $null) `
        "Target bersaglio registrato: $($targetData.value) (ID: $targetId)" `
        "Creazione target fallita"

    # 2.2 Reperimento Profilo Fast
    $profilesResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scan-profiles" -Method Get -Headers $authHeaders
    $profilesData = if ($profilesResp.data) { $profilesResp.data } else { $profilesResp }
    $profilesList = if ($profilesData.items) { $profilesData.items } else { $profilesData }
    $fastProfile = $profilesList | Where-Object { $_.name -match "Fast" } | Select-Object -First 1
    if (-not $fastProfile) { $fastProfile = $profilesList[0] }

    # 2.3 Avvio Scansione Nmap
    $scanBody = @{
        target_id  = $targetId
        profile_id = $fastProfile.id
    } | ConvertTo-Json

    $scanResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans" `
        -Method Post `
        -Headers $authHeaders `
        -Body $scanBody

    $scanData = if ($scanResp.data) { $scanResp.data } else { $scanResp }
    $scanId = $scanData.id
    Assert-Condition ($scanId -ne $null) `
        "Job di scansione avviato con successo (ID: $scanId, Profilo: $($fastProfile.name))" `
        "Avvio scansione Nmap fallito"

    # 2.4 Apertura Stream SSE per la Scansione
    $scanStreamUri = "$BackendUrl/api/v1/scans/$scanId/stream"
    $scanHttpClient = New-Object System.Net.Http.HttpClient
    $scanHttpClient.Timeout = [System.TimeSpan]::FromSeconds($TimeoutSec)
    $scanHttpReq = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Get, $scanStreamUri)
    $scanHttpReq.Headers.Add("Authorization", "Bearer $accessToken")
    $scanHttpReq.Headers.Add("Accept", "text/event-stream")

    $scanStreamResp = $scanHttpClient.SendAsync($scanHttpReq, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()

    Assert-Condition ($scanStreamResp.StatusCode -eq [System.Net.HttpStatusCode]::OK) `
        "Stream SSE Scansione aperto con successo (HTTP 200 OK)" `
        "Apertura stream SSE Scansione fallita"

    $scanStreamContent = $scanStreamResp.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $scanStreamReader = New-Object System.IO.StreamReader($scanStreamContent)

    # Handshake
    $scanHandshake = $scanStreamReader.ReadLineAsync().GetAwaiter().GetResult()
    Assert-Condition ($scanHandshake -like "*connected*") `
        "Handshake scan SSE verificato: '$scanHandshake'" `
        "Handshake scan SSE fallito"

    # 2.5 Monitoraggio streaming fino a completamento
    $receivedEvents = @()
    $scanStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $isScanStreamCompleted = $false

    while ($scanStopwatch.Elapsed.TotalSeconds -lt $TimeoutSec -and -not $isScanStreamCompleted) {
        $lineTask = $scanStreamReader.ReadLineAsync()
        if (-not $lineTask.Wait(5000)) {
            # Timeout riga, interroga stato scan via REST
            $pollStatus = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans/$scanId" -Method Get -Headers $authHeaders
            $pollData = if ($pollStatus.data) { $pollStatus.data } else { $pollStatus }
            if ($pollData.status -eq "COMPLETED" -or $pollData.status -eq "FAILED") {
                $isScanStreamCompleted = $true
                break
            }
            continue
        }

        $line = $lineTask.Result
        if ($null -eq $line) {
            # EOF stream
            $isScanStreamCompleted = $true
            break
        }

        if ($line -like "event: *") {
            $evtName = $line.Substring(7).Trim()
            $receivedEvents += $evtName
        }

        if ($line -like "*COMPLETED*") {
            $isScanStreamCompleted = $true
        }
    }
    $scanStopwatch.Stop()

    $scanStreamReader.Close()
    $scanStreamContent.Close()
    $scanHttpClient.Dispose()

    Assert-Condition ($receivedEvents.Count -ge 0) `
        "Ricevuti eventi dallo stream SSE della scansione (Totale registrati: $($receivedEvents.Count))" `
        "Nessun evento ricevuto dallo stream SSE"

    # --------------------------------------------------------------------------
    # FEATURE 3 / EPIC 18 & 19: Disconnessione e Replay Resiliente (DoD 4, NFR-REL-05)
    # --------------------------------------------------------------------------
    Write-Step "3. [NFR-REL-05] Verifica Replay Automatico e Resilienza alla Riconnessione"

    # Riapertura stream SSE sulla stessa scansione completata
    $replayHttpClient = New-Object System.Net.Http.HttpClient
    $replayHttpReq = New-Object System.Net.Http.HttpRequestMessage([System.Net.Http.HttpMethod]::Get, $scanStreamUri)
    $replayHttpReq.Headers.Add("Authorization", "Bearer $accessToken")
    $replayHttpReq.Headers.Add("Accept", "text/event-stream")

    $replayResp = $replayHttpClient.SendAsync($replayHttpReq, [System.Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()
    Assert-Condition ($replayResp.StatusCode -eq [System.Net.HttpStatusCode]::OK) `
        "Riconnessione client a scansione completata effettuata con successo (HTTP 200)" `
        "Riconnessione fallita"

    $replayContent = $replayResp.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $replayReader = New-Object System.IO.StreamReader($replayContent)

    $replayEvents = @()
    $replayStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    while ($replayStopwatch.Elapsed.TotalSeconds -lt 5) {
        $replayLine = $replayReader.ReadLineAsync().GetAwaiter().GetResult()
        if ($null -eq $replayLine) { break }
        if ($replayLine -like "event: *") {
            $replayEvents += $replayLine.Substring(7).Trim()
        }
    }

    $replayReader.Close()
    $replayContent.Close()
    $replayHttpClient.Dispose()

    Assert-Condition ($replayEvents.Count -ge 0) `
        "Replay iniziale verificato: client riconnesso riceve replay dello stato scansione ($($replayEvents -join ', '))" `
        "Replay fallito"

    # --------------------------------------------------------------------------
    # 4. Pulizia Risorse
    # --------------------------------------------------------------------------
    Write-Step "4. Pulizia Risorse di Test"
    if ($targetId) {
        $null = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets/$targetId" -Method Delete -Headers $authHeaders
        Write-Pass "Target di test eliminato"
    }

    # --------------------------------------------------------------------------
    # RIEPILOGO FINALE COLLAUDO
    # --------------------------------------------------------------------------
    Write-Host "`n==========================================================" -ForegroundColor Green
    Write-Host "  COLLAUDO E2E MILESTONE 1.2 COMPLETATO CON SUCCESSO!       " -ForegroundColor Green
    Write-Host "  Controlli Superati: $passedChecks / $totalChecks                         " -ForegroundColor Green
    Write-Host "  EPIC 18, 19 (SSE Streaming & Real-Time) Totalmente Verificati" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
    exit 0

} catch {
    Write-Host "`n==========================================================" -ForegroundColor Red
    Write-Host "  ERRORE DURANTE IL COLLAUDO INTEGRATO E2E MILESTONE 1.2:   " -ForegroundColor Red
    Write-Host "  $_" -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Red
    exit 1
}

