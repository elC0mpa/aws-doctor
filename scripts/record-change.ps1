param(
    [Parameter(Mandatory = $true)]
    [string]$Summary,
    [ValidateSet("feat", "fix", "docs", "refactor", "test", "chore", "perf", "ci", "build")]
    [string]$Type = "chore"
)

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$changelog = Join-Path $root "CHANGELOG.md"

$entryDate = Get-Date -Format "yyyy-MM-dd"
$entry = "- $entryDate: $Type - $Summary"

if (-not (Test-Path $changelog)) {
    @"
# Changelog

All notable changes to this project are recorded in this file.
This file follows the Keep a Changelog format.

## [Unreleased]

"@ | Set-Content -Path $changelog
}

$lines = Get-Content $changelog
$unreleasedIndex = [array]::IndexOf($lines, "## [Unreleased]")

if ($unreleasedIndex -lt 0) {
    throw "CHANGELOG.md is missing the '## [Unreleased]' section."
}

$insertAt = $unreleasedIndex + 1
if ($insertAt -ge $lines.Count) {
    $lines += ""
} elseif ($lines[$insertAt].Trim() -ne "") {
    $lines = @($lines[0..$unreleasedIndex]) + "" + $lines[$insertAt..($lines.Count - 1)]
}

$insertAt = $unreleasedIndex + 2
if ($insertAt -ge $lines.Count) {
    $lines += $entry
} else {
    $lines = @($lines[0..($insertAt - 1)]) + $entry + $lines[$insertAt..($lines.Count - 1)]
}

Set-Content -Path $changelog -Value $lines
Write-Host "Added entry to CHANGELOG.md"
