#!/usr/bin/env pwsh
# Install script for akv CLI tool
# Supports Windows with automatic platform detection
# Usage: ./install.ps1 [VERSION]
# Environment variables:
#   AKV_INSTALL_DIR - Installation directory (default: ~/AppData/Local/Programs/bin)

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "voidbear-io/akv"
$BinaryName = "akv.exe"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    $input
    $host.UI.RawUI.ForegroundColor = $fc
}

# Get default install directory
function Get-DefaultInstallDir {
    if ($env:AKV_INSTALL_DIR) {
        return $env:AKV_INSTALL_DIR
    }
    return Join-Path $env:USERPROFILE "AppData\Local\Programs\bin"
}

# Detect architecture
function Get-Architecture {
    $arch = [System.Environment]::Is64BitOperatingSystem
    if ($arch) {
        return "amd64"
    }
    return "unknown"
}

# Get latest version from GitHub
function Get-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    
    $headers = @{}
    if ($env:GITHUB_TOKEN) {
        $headers["Authorization"] = "token $($env:GITHUB_TOKEN)"
    }
    
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Headers $headers
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to detect latest version: $_"
        exit 1
    }
}

# Download file
function Download-File($Url, $Output) {
    Write-Host "Downloading from $Url..."
    
    $headers = @{}
    if ($env:GITHUB_TOKEN) {
        $headers["Authorization"] = "token $($env:GITHUB_TOKEN)"
    }
    
    try {
        $ProgressPreference = 'SilentlyContinue'
        Invoke-WebRequest -Uri $Url -OutFile $Output -Headers $headers
        $ProgressPreference = 'Continue'
    }
    catch {
        Write-ColorOutput Red "Error: Failed to download from $Url"
        return $false
    }
    
    return $true
}

# Main installation
function Main {
    $arch = Get-Architecture
    
    if ($arch -eq "unknown") {
        Write-ColorOutput Red "Error: Unsupported architecture"
        exit 1
    }
    
    Write-Host "Detected platform: windows/$arch"
    
    # Get version
    if (-not $Version) {
        Write-Host "Detecting latest version..."
        $Version = Get-LatestVersion
        if (-not $Version) {
            Write-ColorOutput Red "Error: Could not detect latest version"
            exit 1
        }
        Write-Host "Latest version: $Version"
    }
    
    # Remove 'v' prefix if present
    $versionForUrl = $Version -replace '^v', ''
    
    # Set install directory
    $installDir = Get-DefaultInstallDir
    Write-Host "Install directory: $installDir"
    
    # Create install directory
    if (-not (Test-Path $installDir)) {
        Write-Host "Creating install directory..."
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }
    
    # Construct download URL
    $archiveName = "akv-windows-$arch-v$versionForUrl.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/v$versionForUrl/$archiveName"
    
    Write-Host "Downloading ${archiveName}..."
    
    # Create temp directory
    $tempDir = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    try {
        # Download archive
        $archivePath = Join-Path $tempDir $archiveName
        if (-not (Download-File $downloadUrl $archivePath)) {
            # Try with 'v' prefix in tag
            $archiveName = "akv-windows-$arch-$Version.zip"
            $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"
            Write-Host "Retrying with alternate URL format..."
            if (-not (Download-File $downloadUrl $archivePath)) {
                Write-ColorOutput Red "Error: Failed to download release archive"
                exit 1
            }
        }
        
        Write-Host "Download complete"
        
        # Extract archive
        Write-Host "Extracting archive..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
        
        # Find binary
        $binaryPath = Get-ChildItem -Path $tempDir -Filter $BinaryName -Recurse | Select-Object -First 1
        if (-not $binaryPath) {
            Write-ColorOutput Red "Error: Could not find binary in archive"
            exit 1
        }
        
        # Install binary
        $installPath = Join-Path $installDir $BinaryName
        Write-Host "Installing to ${installPath}..."
        Copy-Item -Path $binaryPath.FullName -Destination $installPath -Force
        
        # Verify installation
        if (Test-Path $installPath) {
            Write-ColorOutput Green "✓ akv ${Version} installed successfully!"
            
            # Check if install directory is in PATH
            $userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
            if ($userPath -notlike "*$installDir*") {
                Write-Host ""
                Write-ColorOutput Yellow "Warning: ${installDir} is not in your PATH"
                Write-Host "Add the following directory to your PATH environment variable:"
                Write-Host "  $installDir"
                Write-Host ""
                Write-Host "You can do this by running:"
                Write-Host "  [System.Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$installDir`", 'User')"
            }
            
            # Show version
            Write-Host ""
            Write-Host "Installed version:"
            try {
                & $installPath --version 2>$null
            }
            catch {
                Write-Host "(version command not available)"
            }
        }
        else {
            Write-ColorOutput Red "Error: Installation failed"
            exit 1
        }
    }
    finally {
        # Cleanup temp directory
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force
        }
    }
}

# Run main
Main
