# Pure checks shared by the document validator and its fixture tests.
function Get-AgentMarkdownFiles {
    param([string]$Root)
    $pending = [Collections.Generic.Stack[string]]::new()
    foreach ($folder in @('docs', '.agent', '.agents')) {
        $path = Join-Path $Root $folder
        if (Test-Path -LiteralPath $path) { $pending.Push($path) }
    }
    $ignored = [IO.Path]::GetFullPath((Join-Path $Root '.agent/artifacts'))
    while ($pending.Count -gt 0) {
        foreach ($entry in Get-ChildItem -LiteralPath $pending.Pop() -Force) {
            if ($entry.FullName -eq $ignored -or ($entry.Attributes -band [IO.FileAttributes]::ReparsePoint)) { continue }
            if ($entry.PSIsContainer) { $pending.Push($entry.FullName) }
            elseif ($entry.Extension -eq '.md') { $entry }
        }
    }
}

function ConvertTo-DomainGlobRegex {
    param([string]$Pattern)
    '^' + [regex]::Escape($Pattern).Replace('\*\*/', '(?:.*/)?').Replace('\*\*', '.*').Replace('\*', '[^/]*').Replace('\?', '[^/]') + '$'
}

function Get-AgentSourcePaths {
    param([string]$Root)
    $pending = [Collections.Generic.Stack[string]]::new()
    foreach ($folder in @('backend', 'frontend/src')) {
        $path = Join-Path $Root $folder
        if (Test-Path -LiteralPath $path) { $pending.Push($path) }
    }
    while ($pending.Count -gt 0) {
        foreach ($entry in Get-ChildItem -LiteralPath $pending.Pop() -Force) {
            if ($entry.Attributes -band [IO.FileAttributes]::ReparsePoint) { continue }
            if ($entry.PSIsContainer) {
                if ($entry.Name -notin @('.git', 'node_modules', 'vendor', 'dist', 'build', '.gocache')) { $pending.Push($entry.FullName) }
            } elseif ($entry.Extension -in @('.go', '.sql', '.ts', '.tsx', '.css')) {
                $entry.FullName.Substring($Root.Length).TrimStart('\', '/').Replace('\', '/')
            }
        }
    }
}

function Test-AgentDomainMap {
    param([string]$Root, [string]$MapText)
    $rules = @()
    foreach ($line in ($MapText -split '\r?\n')) {
        if ($line -notmatch '^\|\s*`') { continue }
        $cells = $line.Split('|')
        if ($cells.Count -ne 5) { 'Invalid domain row'; continue }
        $documentMatch = [regex]::Match($cells[3], '`([^`]+\.md)`')
        if (!$documentMatch.Success) { 'Missing domain document'; continue }
        $document = $documentMatch.Groups[1].Value
        if ($document -notmatch '^docs/[^:]+\.md$' -or $document -match '(\\|(^|/)\.\.(/|$))') { "Nonportable domain document: $document"; continue }
        if (!(Test-Path -LiteralPath (Join-Path $Root $document) -PathType Leaf)) { "Missing domain document: $document" }
        foreach ($match in [regex]::Matches($cells[1], '`([^`]+)`')) {
            $glob = $match.Groups[1].Value
            if ($glob -match '(^/|\\|(^|/)\.\.(/|$)|:)') { "Nonportable domain glob: $glob"; continue }
            $rules += [pscustomobject]@{ glob = $glob; regex = (ConvertTo-DomainGlobRegex $glob); document = $document; hits = 0 }
        }
    }
    if (!$rules.Count) { 'Empty domain map'; return }
    $paths = @(Get-AgentSourcePaths $Root)
    foreach ($extra in @('frontend/nginx.conf', 'frontend/vite.config.ts')) {
        if (Test-Path -LiteralPath (Join-Path $Root $extra)) { $paths += $extra }
    }
    foreach ($path in $paths) {
        $owners = @()
        foreach ($rule in $rules) {
            if ($path -cmatch $rule.regex) { $rule.hits++; $owners += $rule.document }
        }
        $owners = @($owners | Select-Object -Unique)
        if (!$owners.Count) { "Unmapped source: $path" }
        elseif ($owners.Count -gt 1) { "Ambiguous domain: $path -> $($owners -join ', ')" }
    }
    foreach ($rule in $rules) { if (!$rule.hits) { "Stale domain glob: $($rule.glob)" } }
}

function Test-AgentSkillManifest {
    param([string]$Text, [string]$FolderName)
    $front = [regex]::Match($Text, '(?s)\A---\r?\n(.*?)\r?\n---(?:\r?\n|$)')
    if (!$front.Success) { 'Missing skill frontmatter'; return }
    # Repository skills deliberately use only two plain scalar fields; no YAML dependency needed.
    $fields = @{}
    foreach ($line in ($front.Groups[1].Value -split '\r?\n')) {
        if ($line -notmatch '^(name|description):\s+(.+)$') { "Unsupported skill manifest line: $line"; continue }
        $key = $Matches[1]; $value = $Matches[2].Trim()
        if ($fields.ContainsKey($key)) { "Duplicate skill field: $key" }
        $fields[$key] = $value
    }
    if (!$fields.ContainsKey('name') -or $fields['name'] -cnotmatch '^[a-z0-9]+(-[a-z0-9]+)*$' -or $fields['name'].Length -gt 64 -or $fields['name'] -cne $FolderName) { "Invalid skill name: $FolderName" }
    if (!$fields.ContainsKey('description') -or [string]::IsNullOrWhiteSpace($fields['description']) -or $fields['description'].Length -gt 300 -or $fields['description'] -match '[<>]|:\s') { "Invalid skill description: $FolderName" }
    if ($Text.Substring($front.Length).Trim().Length -eq 0) { "Empty skill body: $FolderName" }
    if ($Text -match '\[TODO[^\]]*\]|\[INSERT[^\]]*\]') { "Unfinished skill scaffold: $FolderName" }
}

function Test-AgentStateLimits {
    param([string]$Text)
    foreach ($spec in @(@('Decisioni recenti', 10), @('Log sessioni', 5))) {
        $section = [regex]::Match($Text, '(?ms)^## ' + [regex]::Escape($spec[0]) + '[^\n]*\n(.*?)(?=^## |\z)')
        if (!$section.Success) { "Missing state section: $($spec[0])"; continue }
        $count = [regex]::Matches($section.Groups[1].Value, '(?m)^- ').Count
        if ($count -gt $spec[1]) { "State limit exceeded: $($spec[0]) ($count > $($spec[1]))" }
    }
}
