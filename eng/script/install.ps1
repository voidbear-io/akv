param (
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

$Repo = "voidbear-io/akv"
$ApiUrl = if ($Version -eq "latest") { "https://api.github.com/repos/$Repo/releases/latest" } else { "https://api.github.com/repos/$Repo/releases/tags/$Version" }

if ($IsWindows -or [System.Environment]::OSVersion.Platform -eq "Win32NT") {
    $OsName = "windows"
} elseif ($IsLinux) {
    $OsName = "linux"
} elseif ($IsMacOS) {
    $OsName = "darwin"
} else {
    throw "Unsupported Operating System"
}

$Arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
if ($Arch -eq [System.Runtime.InteropServices.Architecture]::X64) {
    $ArchName = "amd64"
} elseif ($Arch -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
    $ArchName = "arm64"
} elseif ($Arch -eq [System.Runtime.InteropServices.Architecture]::X86) {
    $ArchName = "386"
} else {
    throw "Unsupported Architecture: $Arch"
}

$InstallDir = $env:AKV_INSTALL_DIR
if ([string]::IsNullOrEmpty($InstallDir)) {
    if ($OsName -eq "windows") {
        $InstallDir = Join-Path $env:USERPROFILE "AppData\Local\Programs\bin"
    } else {
        $InstallDir = Join-Path $env:HOME ".local\bin"
    }
}

if (-not (Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$Ext = if ($OsName -eq "windows") { "zip" } else { "tar.gz" }

Write-Host "Fetching release information for $Repo ($Version)..."
$Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{"Accept"="application/vnd.github.v3+json"}
$Asset = $Release.assets | Where-Object { $_.name -like "akv-${OsName}-${ArchName}-v*.$Ext" }

if (-not $Asset) {
    throw "Could not find a release asset for ${OsName} ${ArchName}"
}

$DownloadUrl = $Asset.browser_download_url
$TmpDir = Join-Path [System.IO.Path]::GetTempPath() ([guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null
$TmpFile = Join-Path $TmpDir $Asset.name

Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile

if ($Ext -eq "zip") {
    Expand-Archive -Path $TmpFile -DestinationPath $TmpDir -Force
} else {
    & tar -xzf $TmpFile -C $TmpDir
    if ($LASTEXITCODE -ne 0) { throw "Failed to extract tar.gz archive" }
}

$ExecutableName = if ($OsName -eq "windows") { "akv.exe" } else { "akv" }
$ExtractedFile = Join-Path $TmpDir $ExecutableName
$DestFile = Join-Path $InstallDir $ExecutableName
Move-Item -Path $ExtractedFile -Destination $DestFile -Force

if ($OsName -ne "windows") {
    & chmod +x $DestFile
}

Remove-Item -Path $TmpDir -Recurse -Force

Write-Host "akv installed to $DestFile"
