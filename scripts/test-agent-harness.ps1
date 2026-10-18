[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent
. (Join-Path $PSScriptRoot 'agent-docs-lib.ps1')
$artifactRoot = [IO.Path]::GetFullPath((Join-Path $repoRoot '.agent/artifacts'))
$fixtureRoot = [IO.Path]::GetFullPath((Join-Path $artifactRoot ('harness-tests/' + [guid]::NewGuid().ToString('N'))))
if (!$fixtureRoot.StartsWith($artifactRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'Fixture path escapes artifacts.' }
$timer = [Diagnostics.Stopwatch]::StartNew()
$checks = 0
function Assert-Check {
    param([bool]$Condition, [string]$Name)
    if (!$Condition) { throw "FAIL $Name" }
    $script:checks++
}
function Write-Fixture {
    param([string]$Relative, [string]$Text)
    $path = Join-Path $fixtureRoot $Relative
    New-Item -ItemType Directory -Path (Split-Path $path -Parent) -Force | Out-Null
    [IO.File]::WriteAllText($path, $Text)
}

Assert-Check ('backend/a/b.go' -cmatch (ConvertTo-DomainGlobRegex 'backend/**')) 'recursive glob'
Assert-Check ('backend/b.go' -cmatch (ConvertTo-DomainGlobRegex 'backend/**/b.go')) 'zero directory glob'
Assert-Check ('backend/a/b.go' -cnotmatch (ConvertTo-DomainGlobRegex 'backend/*.go')) 'single star cannot cross directories'
Assert-Check ('backend/aXgo' -cnotmatch (ConvertTo-DomainGlobRegex 'backend/a.go')) 'literal dot'

Write-Fixture 'backend/internal/service/example.go' 'package service'
Write-Fixture 'docs/domains/example.md' '# Example'
Write-Fixture 'docs/domains/other.md' '# Other'
$map = '| `backend/internal/service/*.go` | example | `docs/domains/example.md` |'
Assert-Check (@(Test-AgentDomainMap $fixtureRoot $map).Count -eq 0) 'mapped sources accepted'
$errors = @(Test-AgentDomainMap $fixtureRoot ($map -replace 'service/', 'missing/'))
Assert-Check (@($errors | Where-Object { $_ -like 'Unmapped source:*' }).Count -eq 1) 'unmapped source rejected'
Assert-Check (@($errors | Where-Object { $_ -like 'Stale domain glob:*' }).Count -eq 1) 'stale glob rejected'
$errors = @(Test-AgentDomainMap $fixtureRoot ($map + "`n" + ($map -replace 'example.md', 'other.md')))
Assert-Check (@($errors | Where-Object { $_ -like 'Ambiguous domain:*' }).Count -eq 1) 'ambiguous ownership rejected'
$errors = @(Test-AgentDomainMap $fixtureRoot ($map -replace 'example.md', 'missing.md'))
Assert-Check (@($errors | Where-Object { $_ -like 'Missing domain document:*' }).Count -eq 1) 'missing document rejected'
Assert-Check (@(Test-AgentDomainMap $fixtureRoot ($map -replace 'docs/domains/example.md', '../outside.md')).Count -gt 0) 'domain document cannot escape root'
Write-Fixture 'backend/vendor/third-party/ignored.go' 'package ignored'
Assert-Check (@(Test-AgentDomainMap $fixtureRoot $map).Count -eq 0) 'vendor excluded'

$skill = "---`nname: example`ndescription: Verify explicit example requests.`n---`nRun the appropriate checks."
Assert-Check (@(Test-AgentSkillManifest $skill 'example').Count -eq 0) 'valid skill accepted'
Assert-Check (@(Test-AgentSkillManifest $skill 'other').Count -gt 0) 'skill directory/name mismatch rejected'
Assert-Check (@(Test-AgentSkillManifest ($skill -replace 'description: ', ("name: duplicate-example`ndescription: ")) 'example').Count -gt 0) 'duplicate fields rejected'
Assert-Check (@(Test-AgentSkillManifest ($skill -replace 'description: [^\n]+', 'description: ') 'example').Count -gt 0) 'empty description rejected'
Assert-Check (@(Test-AgentSkillManifest ($skill -replace 'description: [^\n]+', 'description:    ') 'example').Count -gt 0) 'whitespace description rejected'
Assert-Check (@(Test-AgentSkillManifest ($skill + ' [TODO: fill]') 'example').Count -gt 0) 'unfinished scaffold rejected'
$state = "## Decisioni recenti`n" + ((1..10 | ForEach-Object { "- decision $_" }) -join "`n") + "`n## Log sessioni`n" + ((1..5 | ForEach-Object { "- session $_" }) -join "`n")
Assert-Check (@(Test-AgentStateLimits $state).Count -eq 0) 'bounded state accepted'
Assert-Check (@(Test-AgentStateLimits ($state + "`n- sixth session")).Count -eq 1) 'session growth rejected'
Assert-Check (@(Test-AgentStateLimits ($state -replace '## Log sessioni', "- eleventh decision`n## Log sessioni")).Count -eq 1) 'decision growth rejected'

# Exercise the complete validator in an isolated copy; never mutate live documents.
$canonicalRoot = Join-Path $fixtureRoot 'canonical'
$files = @('AGENTS.md', 'README.md', 'CHANGELOG.md') + @(Get-AgentMarkdownFiles $repoRoot | ForEach-Object { $_.FullName.Substring($repoRoot.Length).TrimStart('\', '/') })
foreach ($relative in $files) {
    $destination = Join-Path $canonicalRoot $relative
    New-Item -ItemType Directory -Path (Split-Path $destination -Parent) -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $repoRoot $relative) -Destination $destination
}
# Preserve only source names: mapping is structural and does not read application bodies.
foreach ($relative in @(Get-AgentSourcePaths $repoRoot) + @('frontend/nginx.conf', 'frontend/vite.config.ts', 'docker-compose.yml', '.env.example', '.github/workflows/ci.yml')) {
    $destination = Join-Path $canonicalRoot $relative
    New-Item -ItemType Directory -Path (Split-Path $destination -Parent) -Force | Out-Null
    [IO.File]::WriteAllText($destination, '')
}
# Linked executable paths must exist, but need not be executed inside the copy.
foreach ($file in Get-ChildItem -LiteralPath $PSScriptRoot -File) {
    $destination = Join-Path $canonicalRoot ('scripts/' + $file.Name)
    New-Item -ItemType Directory -Path (Split-Path $destination -Parent) -Force | Out-Null
    [IO.File]::WriteAllText($destination, '')
}
$ignored = Join-Path $canonicalRoot '.agent/artifacts/downloaded/broken.md'
New-Item -ItemType Directory -Path (Split-Path $ignored -Parent) -Force | Out-Null
[IO.File]::WriteAllText($ignored, '[third party](missing-file.md)')
$shell = (Get-Process -Id $PID).Path
$output = & $shell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot 'verify-agent-docs.ps1') -RepositoryRoot $canonicalRoot 2>&1
Assert-Check ($LASTEXITCODE -eq 0) ('canonical copy accepts ignored artifacts: ' + ($output -join ' '))
$invalid = Join-Path $canonicalRoot '.agents/skills/repo-verification/SKILL.md'
[IO.File]::WriteAllText($invalid, "---`nname: incorrect`ndescription: Invalid fixture.`n---`nCheck.")
$output = & $shell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot 'verify-agent-docs.ps1') -RepositoryRoot $canonicalRoot 2>&1
Assert-Check ($LASTEXITCODE -eq 1 -and ($output -join ' ') -like '*Invalid skill name*') 'complete validator rejects invalid skill'
$timer.Stop()
Write-Output "PASS: $checks harness checks; seconds=$([math]::Round($timer.Elapsed.TotalSeconds, 3)); fixtures=$fixtureRoot"
