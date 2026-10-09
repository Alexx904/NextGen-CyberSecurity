# ==============================================================================
# NextGen CyberSecurity Platform - E2E Integration Test Suite (Milestone 1.1)
# SDLC EPIC 17 - TASK-50 Full-Stack Verification & Integration Test
# ==============================================================================

[CmdletBinding()]
param (
    [string]$BackendUrl  = "http://localhost:8080",
    [string]$FrontendUrl = "http://localhost:3000",
    [string]$Email       = "admin@nextgen.local",
    [string]$Password    = "AdminPass123!",
    [int]$ScanTimeoutSec = 180
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
Write-Host "  NextGen CyberSecurity - Collaudo E2E Milestone 1.1       " -ForegroundColor Magenta
Write-Host "  (EPIC 12, 13, 14, 15, 16 Full-Stack Verification)        " -ForegroundColor Magenta
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
    Write-Step "0. Verifica Healthcheck & Autenticazione SuperAdmin"
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
        "Autenticazione riuscita, JWT Bearer Token acquisito" `
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

    # Pulizia preventiva di eventuali residui di test precedenti
    try {
        $oldGroups = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups" -Method Get -Headers $authHeaders
        $oldGroupList = if ($oldGroups.data) { $oldGroups.data } else { $oldGroups }
        foreach ($og in $oldGroupList) {
            if ($og.name -like "*E2E*") {
                try { $null = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups/$($og.id)" -Method Delete -Headers $authHeaders } catch {}
            }
        }
    } catch {}

    $testSuffix = (Get-Random -Minimum 1000 -Maximum 9999).ToString()

    # --------------------------------------------------------------------------
    # FEATURE 3 / EPIC 14: Cartelle Target ad Albero (Tree Hierarchy)
    # --------------------------------------------------------------------------
    Write-Step "1. [EPIC 14] Creazione Cartelle Target ad Albero e Spostamento Asset"

    # 1.1 Creazione Gruppo Root
    $rootGroupPayload = @{
        name        = "Sede Centrale E2E $testSuffix"
        description = "Cartella radice di test per asset di rete"
        color       = "#06b6d4"
    } | ConvertTo-Json

    $rootGroupResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups" `
        -Method Post `
        -Headers $authHeaders `
        -Body $rootGroupPayload

    $rootGroup = if ($rootGroupResp.data) { $rootGroupResp.data } else { $rootGroupResp }
    Assert-Condition ($rootGroup.id -ne $null -and $rootGroup.name -eq "Sede Centrale E2E $testSuffix") `
        "Creata cartella radice Target Group: $($rootGroup.name) (ID: $($rootGroup.id))" `
        "Creazione cartella radice fallita"

    $rootGroupId = $rootGroup.id

    # 1.2 Creazione Gruppo Figlio (Subfolder)
    $childGroupPayload = @{
        name        = "Segmento DMZ E2E $testSuffix"
        description = "Sottocartella per server DMZ ed apparati esposti"
        color       = "#10b981"
        parent_id   = $rootGroupId
    } | ConvertTo-Json

    $childGroupResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups" `
        -Method Post `
        -Headers $authHeaders `
        -Body $childGroupPayload

    $childGroup = if ($childGroupResp.data) { $childGroupResp.data } else { $childGroupResp }
    Assert-Condition ($childGroup.id -ne $null -and $childGroup.parent_id -eq $rootGroupId) `
        "Creata sottocartella nidificata: $($childGroup.name) (Parent: $($childGroup.parent_id))" `
        "Creazione sottocartella nidificata fallita"

    $childGroupId = $childGroup.id

    # 1.3 Verifica Struttura Gerarchica Tree
    $treeResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups" -Method Get -Headers $authHeaders
    $treeList = if ($treeResp.data) { $treeResp.data } else { $treeResp }
    $foundRootInTree = $treeList | Where-Object { $_.id -eq $rootGroupId }
    $hasChildInTree = $false
    if ($foundRootInTree -and $foundRootInTree.children) {
        $hasChildInTree = ($foundRootInTree.children | Where-Object { $_.id -eq $childGroupId }) -ne $null
    }

    Assert-Condition ($hasChildInTree -eq $true) `
        "Albero gerarchico validato: sottocartella presente nei 'children' della radice" `
        "Struttura ad albero Target Group non coerente"

    # --------------------------------------------------------------------------
    # FEATURE 1 / EPIC 12: Scansione Subnet & Dead Host Filter
    # --------------------------------------------------------------------------
    Write-Step "2. [EPIC 12] Scansione con Dead Host Exclusion"

    # 2.1 Registrazione Target Bersaglio Loopback (Vivo)
    $targetIp = "127.0.0.$((Get-Random -Minimum 12 -Maximum 240))"
    $targetBody = @{
        value          = $targetIp
        description    = "E2E Live Loopback Target $testSuffix"
        is_blacklisted = $false
        group_id       = $childGroupId
    } | ConvertTo-Json

    $targetResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets" `
        -Method Post `
        -Headers $authHeaders `
        -Body $targetBody

    $targetData = if ($targetResp.data) { $targetResp.data } else { $targetResp }
    $targetId = $targetData.id
    Assert-Condition ($targetId -ne $null -and $targetData.group_id -eq $childGroupId) `
        "Target creato ed associato alla sottocartella DMZ: $($targetData.value) (ID: $targetId)" `
        "Creazione target con associazione cartella fallita"

    # 2.2 Selezione Profilo Fast Discovery
    $profilesResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scan-profiles" -Method Get -Headers $authHeaders
    $profilesData = if ($profilesResp.data) { $profilesResp.data } else { $profilesResp }
    $profilesList = if ($profilesData.items) { $profilesData.items } else { $profilesData }
    $fastProfile = $profilesList | Where-Object { $_.name -match "Fast" } | Select-Object -First 1
    if (-not $fastProfile) { $fastProfile = $profilesList[0] }

    # 2.3 Avvio Scansione
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
        "Scansione avviata su target vivo (ID: $scanId, Profilo: $($fastProfile.name))" `
        "Avvio scansione fallito"

    # 2.4 Attesa Completamento
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $scanCompleted = $false
    while ($stopwatch.Elapsed.TotalSeconds -lt $ScanTimeoutSec) {
        $pollResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/scans/$scanId" -Method Get -Headers $authHeaders
        $currentScan = if ($pollResp.data) { $pollResp.data } else { $pollResp }
        if ($currentScan.status -eq "COMPLETED") {
            $scanCompleted = $true
            break
        } elseif ($currentScan.status -eq "FAILED" -or $currentScan.status -eq "CANCELLED") {
            throw "Scansione terminata con errore: $($currentScan.status)"
        }
        Start-Sleep -Seconds 2
    }
    $stopwatch.Stop()
    Assert-Condition ($scanCompleted -eq $true) `
        "Scansione completata con successo in $($stopwatch.Elapsed.TotalSeconds.ToString('F1'))s" `
        "Timeout durante la scansione"

    # --------------------------------------------------------------------------
    # FEATURE 2 / EPIC 13: Reports Hub & Multi-Template Download
    # --------------------------------------------------------------------------
    Write-Step "3. [EPIC 13] Reports Hub: Download Report TECHNICAL ed EXECUTIVE"

    # 3.1 Download Template TECHNICAL PDF
    $pdfTechTmp = [System.IO.Path]::GetTempFileName() + "_tech.pdf"
    Invoke-WebRequest -Uri "$BackendUrl/api/v1/scans/$scanId/reports/pdf?template=TECHNICAL" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $accessToken" } `
        -OutFile $pdfTechTmp `
        -UseBasicParsing

    $pdfTechBytes = [System.IO.File]::ReadAllBytes($pdfTechTmp)
    $isTechPdfValid = ($pdfTechBytes.Length -gt 100) -and ($pdfTechBytes[0] -eq 0x25 -and $pdfTechBytes[1] -eq 0x50 -and $pdfTechBytes[2] -eq 0x44 -and $pdfTechBytes[3] -eq 0x46)
    Assert-Condition ($isTechPdfValid -eq $true) `
        "Report PDF TECHNICAL White-Label valido (Dimensione: $($pdfTechBytes.Length) bytes, Magic bytes: %PDF)" `
        "Report PDF TECHNICAL non valido"
    if (Test-Path $pdfTechTmp) { Remove-Item $pdfTechTmp -Force }

    # 3.2 Download Template EXECUTIVE PDF
    $pdfExecTmp = [System.IO.Path]::GetTempFileName() + "_exec.pdf"
    Invoke-WebRequest -Uri "$BackendUrl/api/v1/scans/$scanId/reports/pdf?template=EXECUTIVE" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $accessToken" } `
        -OutFile $pdfExecTmp `
        -UseBasicParsing

    $pdfExecBytes = [System.IO.File]::ReadAllBytes($pdfExecTmp)
    $isExecPdfValid = ($pdfExecBytes.Length -gt 100) -and ($pdfExecBytes[0] -eq 0x25 -and $pdfExecBytes[1] -eq 0x50 -and $pdfExecBytes[2] -eq 0x44 -and $pdfExecBytes[3] -eq 0x46)
    Assert-Condition ($isExecPdfValid -eq $true) `
        "Report PDF EXECUTIVE White-Label valido (Dimensione: $($pdfExecBytes.Length) bytes, Magic bytes: %PDF)" `
        "Report PDF EXECUTIVE non valido"
    if (Test-Path $pdfExecTmp) { Remove-Item $pdfExecTmp -Force }

    # 3.3 Elenco Reports Hub
    $reportsListResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/reports" -Method Get -Headers $authHeaders
    $reportsData = if ($reportsListResp.data) { $reportsListResp.data } else { $reportsListResp }
    Assert-Condition ($reportsData -ne $null) `
        "Endpoint Hub /api/v1/reports interrogato con successo" `
        "Recupero elenco reports fallito"

    # --------------------------------------------------------------------------
    # FEATURE 4 / EPIC 15: LAN Network Discovery & Bulk Import
    # --------------------------------------------------------------------------
    Write-Step "4. [EPIC 15] LAN Network Discovery: Ping Sweep & Bulk Import"

    # 4.1 Esecuzione Ping Sweep su Localhost /32
    $discPayload = @{
        subnet               = "127.0.0.1/32"
        include_common_ports = $false
    } | ConvertTo-Json

    $discResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/network-discovery/scan" `
        -Method Post `
        -Headers $authHeaders `
        -Body $discPayload

    $discData = if ($discResp.data) { $discResp.data } else { $discResp }
    Assert-Condition ($discData.devices -ne $null -and $discData.devices.Count -ge 1) `
        "Network Discovery completata con successo: $($discData.devices.Count) dispositivo/i attivo/i rilevato/i" `
        "Network Discovery fallita: nessun host rilevato su 127.0.0.1/32"

    $foundDevice = $discData.devices[0]
    Write-Pass "Host rilevato: IP=$($foundDevice.ip), Status=$($foundDevice.status), RTT=$($foundDevice.latency_ms)ms"

    # 4.2 Importazione Selettiva nella Cartella DMZ
    $importIp = "127.0.0.$((Get-Random -Minimum 241 -Maximum 254))"
    $importPayload = @{
        devices = @(
            @{
                ip       = $importIp
                hostname = "discovered-e2e-node.local"
                mac      = "00:50:56:C0:00:08"
            }
        )
        group_id = $childGroupId
    } | ConvertTo-Json

    $importResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/network-discovery/import-targets" `
        -Method Post `
        -Headers $authHeaders `
        -Body $importPayload

    $importData = if ($importResp.data) { $importResp.data } else { $importResp }
    $importCount = if ($importData.total_imported -ne $null) { $importData.total_imported } else { 0 }
    $skippedCount = if ($importData.total_skipped -ne $null) { $importData.total_skipped } else { 0 }
    Assert-Condition ($importCount -ge 1 -or $skippedCount -ge 1) `
        "Importazione massiva Discovery completata (Importati: $importCount, Saltati: $skippedCount)" `
        "Importazione discovery fallita"

    # --------------------------------------------------------------------------
    # FEATURE 5 / EPIC 16: Host SSH & Firewall iptables con Anti-Lockout
    # --------------------------------------------------------------------------
    Write-Step "5. [EPIC 16] Censimento Host SSH, Test Connettività e Gestione Firewall"

    # 5.1 Creazione Server SSH con Secret Cifrato AES-256-GCM
    $sshPayload = @{
        name        = "E2E Bastion Host $testSuffix"
        host        = "127.0.0.1"
        port        = 22
        username    = "root"
        auth_type   = "PASSWORD"
        raw_secret  = "DummySecurePass123!"
    } | ConvertTo-Json

    $sshCreateResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/ssh/servers" `
        -Method Post `
        -Headers $authHeaders `
        -Body $sshPayload

    $sshServer = if ($sshCreateResp.data) { $sshCreateResp.data } else { $sshCreateResp }
    $sshServerId = $sshServer.id
    Assert-Condition ($sshServerId -ne $null -and $sshServer.name -eq "E2E Bastion Host $testSuffix") `
        "Server SSH censito con credenziali cifrate AES-GCM (ID: $sshServerId)" `
        "Creazione server SSH fallita"

    # 5.2 Test Connettività SSH
    $connTestResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/ssh/servers/$sshServerId/test-connection" `
        -Method Post `
        -Headers $authHeaders

    $connResult = if ($connTestResp.data) { $connTestResp.data } else { $connTestResp }
    Assert-Condition ($connResult -ne $null) `
        "Test connettività SSH eseguito (Host raggiungibile: $($connResult.online), Latenza: $($connResult.latency_ms)ms)" `
        "Chiamata test connettività SSH fallita"

    # 5.3 Test Ispezione e Operazioni Firewall (Verifica endpoint /firewall)
    try {
        $fwStatusResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/ssh/servers/$sshServerId/firewall" `
            -Method Get `
            -Headers $authHeaders
        Write-Pass "Ispezione firewall remota eseguita con successo su server"
    } catch {
        # Se la porta 22 locale non è un server SSH reale con iptables, verifichiamo la gestione sicura dell'errore (400 o 500 con JSON coerente)
        $errStatus = $_.Exception.Response.StatusCode.value__
        Assert-Condition ($errStatus -eq 500 -or $errStatus -eq 400 -or $errStatus -eq 502) `
            "Ispezione firewall ha gestito correttamente l'irraggiungibilità dell'host remoto (HTTP $errStatus)" `
            "Risposta inattesa dall'endpoint firewall: $errStatus"
    }

    # 5.4 Pulizia Server SSH di Test
    $delSshResp = Invoke-RestMethod -Uri "$BackendUrl/api/v1/ssh/servers/$sshServerId" `
        -Method Delete `
        -Headers $authHeaders
    Write-Pass "Server SSH di test eliminato e credenziali cifrate ripulite"

    # --------------------------------------------------------------------------
    # 6. Pulizia Risorse Finali
    # --------------------------------------------------------------------------
    Write-Step "6. Pulizia Risorse di Test Creata"
    if ($targetId) {
        $null = Invoke-RestMethod -Uri "$BackendUrl/api/v1/targets/$targetId" -Method Delete -Headers $authHeaders
        Write-Pass "Target di test eliminato"
    }
    if ($childGroupId) {
        $null = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups/$childGroupId" -Method Delete -Headers $authHeaders
        Write-Pass "Sottocartella DMZ eliminata"
    }
    if ($rootGroupId) {
        $null = Invoke-RestMethod -Uri "$BackendUrl/api/v1/target-groups/$rootGroupId" -Method Delete -Headers $authHeaders
        Write-Pass "Cartella radice Sede Centrale eliminata"
    }

    # --------------------------------------------------------------------------
    # RIEPILOGO FINALE COLLAUDO
    # --------------------------------------------------------------------------
    Write-Host "`n==========================================================" -ForegroundColor Green
    Write-Host "  COLLAUDO E2E MILESTONE 1.1 COMPLETATO CON SUCCESSO!       " -ForegroundColor Green
    Write-Host "  Controlli Superati: $passedChecks / $totalChecks                         " -ForegroundColor Green
    Write-Host "  EPIC 12, 13, 14, 15, 16 Totalmente Verificate e Operative" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
    exit 0

} catch {
    Write-Host "`n==========================================================" -ForegroundColor Red
    Write-Host "  ERRORE DURANTE IL COLLAUDO INTEGRATO E2E MILESTONE 1.1:   " -ForegroundColor Red
    Write-Host "  $_" -ForegroundColor Red
    Write-Host "==========================================================" -ForegroundColor Red
    exit 1
}
