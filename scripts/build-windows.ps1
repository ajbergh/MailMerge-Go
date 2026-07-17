<#
.SYNOPSIS
Builds the Windows release of MailMerge-Go.

.EXAMPLE
.\scripts\build-windows.ps1
.\scripts\build-windows.ps1 -Clean -Debug
#>
param(
    [switch]$Clean,
    [switch]$Debug,
    [switch]$SkipFrontendInstall
)

$ErrorActionPreference = "Stop"
$RepositoryRoot = Split-Path -Parent $PSScriptRoot
$BuildDirectory = Join-Path $RepositoryRoot "build\\bin"

function Require-Command {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "'$Name' is required but was not found on PATH."
    }
}

Require-Command "go"
Require-Command "node"
Require-Command "npm"
Require-Command "wails"

if ($Clean -and (Test-Path $BuildDirectory)) {
    Remove-Item -LiteralPath $BuildDirectory -Recurse -Force
}

Push-Location $RepositoryRoot
try {
    if (-not $SkipFrontendInstall) {
        Push-Location "frontend"
        try {
            npm ci
        }
        finally {
            Pop-Location
        }
    }

    $Version = (git describe --tags --always 2>$null)
    if (-not $Version) { $Version = "dev" }
    $Commit = (git rev-parse --short HEAD 2>$null)
    if (-not $Commit) { $Commit = "unknown" }
    $BuildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd")
    $Ldflags = "-X main.Version=$Version -X main.Commit=$Commit -X main.BuildDate=$BuildDate"

    $BuildArguments = @("-platform", "windows/amd64", "-ldflags", $Ldflags)
    if ($Debug) { $BuildArguments += "-debug" }

    & wails build @BuildArguments
    if ($LASTEXITCODE -ne 0) {
        throw "Wails build failed with exit code $LASTEXITCODE."
    }
}
finally {
    Pop-Location
}
