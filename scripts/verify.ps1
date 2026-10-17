[CmdletBinding()]
param(
    [ValidateSet('Quick', 'Final', 'Full', 'Docs')][string]$Profile = 'Final',
    [ValidateSet('Backend', 'Frontend', 'All')][string]$Scope = 'All',
    [string[]]$BackendPackages = @(),
    [string[]]$FrontendFiles = @(),
    [switch]$Integration
)
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent
$results = [System.Collections.Generic.List[object]]::new()
$runId = (Get-Date -Format 'yyyyMMdd-HHmmss') + '-' + [guid]::NewGuid().ToString('N').Substring(0, 8)
$artifactDirectory = Join-Path $repoRoot '.agent/artifacts/testing'
$reportPath = Join-Path $artifactDirectory "$runId.json"
$runExitCode = 0

function Invoke-Check {
    param([string]$Directory, [string]$Program, [string[]]$ArgumentList)
    $timer = [System.Diagnostics.Stopwatch]::StartNew()
    $commandExitCode = 1
    Push-Location (Join-Path $repoRoot $Directory)
    try {
        Write-Host "RUN [$Directory] $Program $($ArgumentList -join ' ')"
        & $Program @ArgumentList | Out-Host
        $commandExitCode = $LASTEXITCODE
        if ($null -eq $commandExitCode) { $commandExitCode = 1 }
    } finally {
        Pop-Location
        $timer.Stop()
        $results.Add([pscustomobject]@{
            directory = $Directory
            command = "$Program $($ArgumentList -join ' ')"
            seconds = [math]::Round($timer.Elapsed.TotalSeconds, 3)
            exitCode = $commandExitCode
        })
        Write-Host "RESULT exit=$commandExitCode seconds=$([math]::Round($timer.Elapsed.TotalSeconds, 3))"
    }
    if ($commandExitCode -ne 0) { throw "Verification failed (exit $commandExitCode)" }
}

try {
    if ($Profile -eq 'Full' -and $Scope -ne 'All') { throw 'Full requires Scope All.' }
    if ($Integration -and ($Profile -in @('Quick', 'Docs') -or $Scope -eq 'Frontend')) {
        throw 'Integration requires Final with Backend/All, or Full.'
    }
    if ($Profile -ne 'Quick' -and ($BackendPackages.Count -gt 0 -or $FrontendFiles.Count -gt 0)) {
        throw 'Target selections are only allowed in Quick; final suites must be complete.'
    }
    $backendSelected = $Scope -in @('Backend', 'All')
    $frontendSelected = $Scope -in @('Frontend', 'All')
    if ($Profile -eq 'Quick') {
        if ($backendSelected -and $BackendPackages.Count -eq 0) { throw 'Quick requires BackendPackages for the selected backend.' }
        if ($frontendSelected -and $FrontendFiles.Count -eq 0) { throw 'Quick requires FrontendFiles for the selected frontend.' }
        if ((!$backendSelected -and $BackendPackages.Count -gt 0) -or (!$frontendSelected -and $FrontendFiles.Count -gt 0)) {
            throw 'Target selections must match Scope.'
        }
        foreach ($package in $BackendPackages) {
            if ($package -notmatch '^\./[a-zA-Z0-9_./-]+$' -or $package -match '(^|/)\.\.(/|$)') {
                throw 'BackendPackages must be relative Go package paths inside backend.'
            }
        }
        foreach ($file in $FrontendFiles) {
            $frontendRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot 'frontend')) + [System.IO.Path]::DirectorySeparatorChar
            $resolved = [System.IO.Path]::GetFullPath((Join-Path $frontendRoot $file))
            if (!$resolved.StartsWith($frontendRoot, [System.StringComparison]::OrdinalIgnoreCase) -or
                $file -notmatch '^src/.*\.test\.tsx?$' -or !(Test-Path -LiteralPath $resolved -PathType Leaf)) {
                throw 'FrontendFiles must be existing src/*.test.ts or .tsx paths inside frontend.'
            }
        }
    }
    if ($Integration -or $Profile -eq 'Full') {
        foreach ($key in @('TEST_DB_HOST', 'TEST_DB_PORT', 'TEST_DB_NAME', 'TEST_DB_USER', 'TEST_DB_PASSWORD')) {
            if ([string]::IsNullOrEmpty([Environment]::GetEnvironmentVariable($key))) {
                throw "$key is required before running an integration profile."
            }
        }
    }
    $npmProgram = if ($env:OS -eq 'Windows_NT') { 'npm.cmd' } else { 'npm' }
    if ($Profile -ne 'Docs') {
        if ($backendSelected) {
            $packages = if ($Profile -eq 'Quick') { $BackendPackages } else { @('./...') }
            Invoke-Check 'backend' 'go' (@('test') + $packages)
            if ($Profile -ne 'Quick') {
                Invoke-Check 'backend' 'go' @('vet', './...')
                Invoke-Check 'backend' 'go' @('build', './...')
            }
        }
        if ($frontendSelected) {
            $testArguments = if ($Profile -eq 'Quick') { @('test', '--') + $FrontendFiles } else { @('test') }
            Invoke-Check 'frontend' $npmProgram $testArguments
            if ($Profile -ne 'Quick') { Invoke-Check 'frontend' $npmProgram @('run', 'build') }
        }
        if ($Integration -or $Profile -eq 'Full') {
            Invoke-Check 'backend' 'go' @('test', '-tags=integration', '-count=1', './internal/repository/postgres')
        }
    }
    if ($Profile -ne 'Quick') {
        $shellProgram = (Get-Process -Id $PID).Path
        Invoke-Check '.' $shellProgram @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', 'scripts/verify-agent-docs.ps1')
        Invoke-Check '.' 'git' @('diff', '--check')
    }
} catch {
    $runExitCode = 1
    Write-Host "FAIL: $($_.Exception.Message)"
} finally {
    New-Item -ItemType Directory -Path $artifactDirectory -Force | Out-Null
    [pscustomobject]@{
        runId = $runId
        timestampUtc = [DateTime]::UtcNow.ToString('o')
        profile = $Profile
        scope = $Scope
        integrationRequested = [bool]($Integration -or $Profile -eq 'Full')
        exitCode = $runExitCode
        checks = @($results.ToArray())
        note = 'Timing evidence only; validate unchanged inputs/environment before reusing results. Not a browser or migration rollback check.'
    } | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $reportPath -Encoding UTF8
    Write-Host "Evidence: $reportPath"
}
exit $runExitCode
