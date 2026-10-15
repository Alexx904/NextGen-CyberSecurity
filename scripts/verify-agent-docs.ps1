[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent
$failures = [System.Collections.Generic.List[string]]::new()
$required = @(
    'AGENTS.md', 'README.md', 'CHANGELOG.md', 'docs/README.md', 'docs/PROJECT_STATE.md',
    'docs/01-product/PRD.md', 'docs/01-product/REQUIREMENTS.md',
    'docs/01-product/USE_CASES.md', 'docs/01-product/USER_STORIES.md',
    'docs/02-design/UX_FLOW.md', 'docs/02-design/UI_SPECIFICATIONS.md',
    'docs/03-architecture/HLD.md', 'docs/03-architecture/DOMAIN_MAP.md',
    'docs/03-architecture/LLD/API_SPEC.md', 'docs/03-architecture/LLD/DATABASE_SCHEMA.md',
    'docs/03-architecture/LLD/COMPONENT_DESIGN.md', 'docs/03-architecture/ADR/README.md',
    'docs/04-development/BACKEND_SETUP.md', 'docs/04-development/FRONTEND_SETUP.md',
    'docs/04-development/TASK_LIST.md', 'docs/04-development/plans/README.md',
    'docs/05-qa/TEST_PLAN.md', 'docs/06-devops/DEPLOYMENT.md', 'docs/06-devops/CI_CD_PIPELINE.md'
)
foreach ($relative in $required) {
    if (-not (Test-Path -LiteralPath (Join-Path $repoRoot $relative))) { $failures.Add("Missing: $relative") }
}
$docsRoot = Join-Path $repoRoot 'docs'
$allowedDirectories = @('01-product', '02-design', '03-architecture', '04-development', '05-qa', '06-devops', 'domains')
foreach ($entry in Get-ChildItem -LiteralPath $docsRoot -Force) {
    if ($entry.PSIsContainer) {
        if ($allowedDirectories -notcontains $entry.Name) { $failures.Add("Unexpected docs directory: $($entry.Name)") }
    } elseif (@('README.md', 'PROJECT_STATE.md') -notcontains $entry.Name) {
        $failures.Add("Unexpected docs root file: $($entry.Name)")
    }
}
foreach ($entry in Get-ChildItem -LiteralPath $docsRoot -Recurse -Force -File) {
    if ($entry.Extension -ne '.md') { $failures.Add("Non-Markdown artifact in docs: $($entry.FullName)") }
}
$files = @(
    Get-Item -LiteralPath (Join-Path $repoRoot 'AGENTS.md'), (Join-Path $repoRoot 'README.md'), (Join-Path $repoRoot 'CHANGELOG.md')
    Get-ChildItem -LiteralPath (Join-Path $repoRoot '.agent'), (Join-Path $repoRoot 'docs') -Recurse -File -Filter '*.md'
)
$links = 0
foreach ($file in $files) {
    $raw = [System.IO.File]::ReadAllText($file.FullName)
    # Ignore fenced examples; inspect Markdown links in prose and tables.
    $prose = [regex]::Replace($raw, '(?ms)^```[^\n]*\n.*?^```\s*$', '')
    foreach ($match in [regex]::Matches($prose, '\[[^\]]+\]\(([^\s)]+)\)')) {
        $target = $match.Groups[1].Value
        if ($target -match '^(https?://|mailto:)') { continue }
        $links++
        if ($target -match '^file:') { $failures.Add("Nonportable link: $($file.Name) -> $target"); continue }
        $parts = $target -split '#', 2
        $destination = if ($parts[0]) { Join-Path $file.DirectoryName ([uri]::UnescapeDataString($parts[0])) } else { $file.FullName }
        if (-not (Test-Path -LiteralPath $destination)) {
            $failures.Add("Broken link: $($file.FullName) -> $target")
            continue
        }
        if ($parts.Count -eq 2 -and $parts[1] -and [System.IO.Path]::GetExtension($destination) -eq '.md') {
            $doc = [System.IO.File]::ReadAllText($destination)
            $slugs = @()
            foreach ($heading in [regex]::Matches($doc, '(?m)^#{1,6}\s+(.+?)\s*\r?$')) {
                $slug = $heading.Groups[1].Value.ToLowerInvariant()
                $slug = [regex]::Replace($slug, '[^\p{L}\p{Nd}_\-\s]', '') -replace '\s', '-'
                $slugs += $slug
            }
            $explicit = '<a\s+(?:id|name)=["' + "'" + ']' + [regex]::Escape($parts[1]) + '["' + "'" + ']'
            if ($slugs -notcontains [uri]::UnescapeDataString($parts[1]) -and $doc -notmatch $explicit) {
                $failures.Add("Unknown anchor: $($file.Name) -> $target")
            }
        }
    }
    $fenceCount = ([regex]::Matches($raw, '(?m)^```')).Count
    if ($fenceCount % 2 -ne 0) { $failures.Add("Unbalanced code fences: $($file.FullName)") }
}
$registry = [System.IO.File]::ReadAllText((Join-Path $repoRoot '.agent/README.md'))
foreach ($folder in @('roles', 'workflows')) {
    foreach ($file in Get-ChildItem -LiteralPath (Join-Path $repoRoot ".agent/$folder") -File -Filter '*.md') {
        if ($registry -notmatch [regex]::Escape("$folder/$($file.Name)")) { $failures.Add("Not indexed: $folder/$($file.Name)") }
    }
}
if ($failures.Count -gt 0) {
    $failures | ForEach-Object { Write-Output "FAIL $_" }
    exit 1
}
Write-Output "PASS: agent-only docs structure; $($required.Count) required documents; $($files.Count) Markdown files; $links local links and anchors; indexed roles/workflows; balanced fences."
exit 0
