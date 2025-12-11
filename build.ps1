# MailMerge-Go Build Script
# This script builds the MailMerge application for Windows

param(
    [switch]$Clean,
    [switch]$Dev,
    [switch]$NoPackage,
    [switch]$Debug,
    [switch]$Help
)

# Script configuration
$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$AppName = "MailMergeApp"
$BuildDir = Join-Path $ScriptDir "build\bin"

# Colors for output
function Write-Success { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Info { param($Message) Write-Host $Message -ForegroundColor Cyan }
function Write-Warning { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Error { param($Message) Write-Host $Message -ForegroundColor Red }

function Show-Help {
    Write-Host ""
    Write-Host "MailMerge-Go Build Script" -ForegroundColor Cyan
    Write-Host "=========================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Clean      Clean build directory before building"
    Write-Host "  -Dev        Build in development mode (includes devtools)"
    Write-Host "  -NoPackage  Skip packaging (don't create installer)"
    Write-Host "  -Debug      Build with debug symbols"
    Write-Host "  -Help       Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\build.ps1                 # Production build"
    Write-Host "  .\build.ps1 -Clean          # Clean production build"
    Write-Host "  .\build.ps1 -Dev            # Development build with devtools"
    Write-Host "  .\build.ps1 -Clean -Debug   # Clean debug build"
    Write-Host ""
}

function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    # Check Go
    try {
        $goVersion = go version 2>&1
        Write-Host "  [OK] Go: $goVersion" -ForegroundColor Green
    }
    catch {
        Write-Error "  [FAIL] Go is not installed or not in PATH"
        Write-Host "  Please install Go from https://go.dev/dl/"
        return $false
    }
    
    # Check Wails CLI
    try {
        $wailsVersion = wails version 2>&1
        Write-Host "  [OK] Wails CLI installed" -ForegroundColor Green
    }
    catch {
        Write-Error "  [FAIL] Wails CLI is not installed"
        Write-Host "  Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        return $false
    }
    
    # Check Node.js
    try {
        $nodeVersion = node --version 2>&1
        Write-Host "  [OK] Node.js: $nodeVersion" -ForegroundColor Green
    }
    catch {
        Write-Error "  [FAIL] Node.js is not installed or not in PATH"
        Write-Host "  Please install Node.js from https://nodejs.org/"
        return $false
    }
    
    # Check npm
    try {
        $npmVersion = npm --version 2>&1
        Write-Host "  [OK] npm: $npmVersion" -ForegroundColor Green
    }
    catch {
        Write-Error "  [FAIL] npm is not installed or not in PATH"
        return $false
    }
    
    return $true
}

function Clean-Build {
    Write-Info "Cleaning build directory..."
    
    if (Test-Path $BuildDir) {
        Remove-Item -Path $BuildDir -Recurse -Force
        Write-Host "  Removed: $BuildDir" -ForegroundColor Yellow
    }
    
    # Clean frontend node_modules if requested
    $frontendNodeModules = Join-Path $ScriptDir "frontend\node_modules"
    # Uncomment below to also clean node_modules
    # if (Test-Path $frontendNodeModules) {
    #     Remove-Item -Path $frontendNodeModules -Recurse -Force
    #     Write-Host "  Removed: frontend\node_modules" -ForegroundColor Yellow
    # }
    
    Write-Success "  Clean complete"
}

function Install-Dependencies {
    Write-Info "Installing frontend dependencies..."
    
    Push-Location (Join-Path $ScriptDir "frontend")
    try {
        npm install 2>&1 | Out-Null
        Write-Success "  Dependencies installed"
    }
    catch {
        Write-Error "  Failed to install dependencies: $_"
        throw
    }
    finally {
        Pop-Location
    }
}

function Build-Application {
    Write-Info "Building application..."
    
    Push-Location $ScriptDir
    try {
        # Build command arguments
        $buildArgs = @()
        
        if ($Dev) {
            $buildArgs += "-devtools"
            Write-Host "  Mode: Development (devtools enabled)" -ForegroundColor Yellow
        }
        else {
            Write-Host "  Mode: Production" -ForegroundColor Green
        }
        
        if ($NoPackage) {
            $buildArgs += "-skippackage"
            Write-Host "  Packaging: Skipped" -ForegroundColor Yellow
        }
        
        if ($Debug) {
            $buildArgs += "-debug"
            Write-Host "  Debug symbols: Enabled" -ForegroundColor Yellow
        }
        
        # Platform
        $buildArgs += "-platform"
        $buildArgs += "windows/amd64"
        
        Write-Host ""
        Write-Host "  Running: wails build $($buildArgs -join ' ')" -ForegroundColor Cyan
        Write-Host ""
        
        # Execute build
        $buildOutput = wails build @buildArgs 2>&1
        $buildOutput | ForEach-Object { Write-Host "  $_" }
        
        if ($LASTEXITCODE -ne 0) {
            throw "Wails build failed with exit code $LASTEXITCODE"
        }
        
        Write-Success "`n  Build complete!"
    }
    catch {
        Write-Error "  Build failed: $_"
        throw
    }
    finally {
        Pop-Location
    }
}

function Show-BuildInfo {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host " Build Summary" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    
    $exePath = Join-Path $BuildDir "$AppName.exe"
    
    if (Test-Path $exePath) {
        $fileInfo = Get-Item $exePath
        $sizeKB = [math]::Round($fileInfo.Length / 1KB, 2)
        $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
        
        Write-Host ""
        Write-Host "  Output: $exePath" -ForegroundColor Green
        Write-Host "  Size: $sizeMB MB ($sizeKB KB)"
        Write-Host "  Created: $($fileInfo.LastWriteTime)"
        Write-Host ""
        
        # Show build directory contents
        Write-Host "  Build directory contents:" -ForegroundColor Cyan
        Get-ChildItem $BuildDir | ForEach-Object {
            $size = if ($_.PSIsContainer) { "<DIR>" } else { "$([math]::Round($_.Length / 1KB, 1)) KB" }
            Write-Host "    $($_.Name) - $size"
        }
    }
    else {
        Write-Warning "  Executable not found at expected location"
    }
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
}

# Main execution
function Main {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host " MailMerge-Go Build Script" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    
    if ($Help) {
        Show-Help
        return
    }
    
    $startTime = Get-Date
    
    try {
        # Check prerequisites
        if (-not (Test-Prerequisites)) {
            Write-Error "`nPrerequisite check failed. Please install missing dependencies."
            exit 1
        }
        
        Write-Host ""
        
        # Clean if requested
        if ($Clean) {
            Clean-Build
            Write-Host ""
        }
        
        # Install dependencies
        Install-Dependencies
        Write-Host ""
        
        # Build
        Build-Application
        
        # Show build info
        Show-BuildInfo
        
        $endTime = Get-Date
        $duration = $endTime - $startTime
        
        Write-Success "Build completed successfully in $([math]::Round($duration.TotalSeconds, 2)) seconds!"
        Write-Host ""
    }
    catch {
        Write-Error "`nBuild failed: $_"
        exit 1
    }
}

# Run main
Main
