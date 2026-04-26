$ErrorActionPreference = "Stop"
$REPO = "elC0mpa/aws-doctor"
$BINARY_NAME = "aws-doctor.exe"
$INSTALL_DIR = Join-Path $env:USERPROFILE ".aws-doctor\bin"

function Log-Info($msg) { 
    if ($msg) { Write-Host "[INFO] $msg" -ForegroundColor Green }
    else { Write-Host "" } 
}
function Log-Error($msg) { Write-Host "[ERROR] $msg" -ForegroundColor Red }

function Detect-Arch {
    $Arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
    if ($Arch -in @("amd64", "x86_64")) { return "amd64" }
    if ($Arch -in @("arm64", "aarch64")) { return "arm64" }
    Log-Error "Unsupported architecture: $Arch"
    exit 1
}

function Get-LatestVersion {
    $headers = @{"User-Agent" = "aws-doctor-installer"}
    $url = "https://api.github.com/repos/$REPO/releases/latest"
    
    try {
        $Response = Invoke-RestMethod -Uri $url -Headers $headers -UseBasicParsing
        return $Response.tag_name
    } catch {
        $status = $_.Exception.Response.StatusCode.value__
        if ($status -eq 403 -or $status -eq 429) {
            Log-Error "GitHub API rate limit exceeded (HTTP $status)."
            Log-Error "This often happens on shared IPs or when using VPNs/Proxies (like Cloudflare WARP)."
            Log-Error "Please wait a few minutes or provide the version manually: .\install.ps1 v2.11.0"
        } else {
            Log-Error "Failed to fetch latest version: $($_.Exception.Message)"
        }
        exit 1
    }
}

function Download-AndVerify($Version, $OS, $Arch, $TmpDir) {
    $CleanVersion = $Version -replace '^v', ''
    $FileName = "aws-doctor_${CleanVersion}_${OS}_${Arch}.zip"
    $DownloadUrl = "https://github.com/$REPO/releases/download/$Version/$FileName"
    $ChecksumUrl = "https://github.com/$REPO/releases/download/$Version/checksums.txt"

    Log-Info "Downloading $FileName..."
    Invoke-WebRequest -Uri $DownloadUrl -OutFile (Join-Path $TmpDir $FileName) -UseBasicParsing
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile (Join-Path $TmpDir "checksums.txt") -UseBasicParsing

    Log-Info "Verifying checksum..."
    $ExpectedLine = Get-Content (Join-Path $TmpDir "checksums.txt") | Where-Object { $_ -match [regex]::Escape($FileName) }
    if (-not $ExpectedLine) { Log-Error "Checksum for $FileName not found in checksums.txt"; exit 1 }
    
    $ExpectedChecksum = ($ExpectedLine -split '\s+')[0]
    $ActualChecksum = (Get-FileHash -Path (Join-Path $TmpDir $FileName) -Algorithm SHA256).Hash.ToLower()

    if ($ExpectedChecksum.ToLower() -ne $ActualChecksum) { Log-Error "Checksum failed!"; exit 1 }
    Log-Info "Checksum verified successfully."
    return (Join-Path $TmpDir $FileName)
}

function Update-Path {
    Log-Info "Checking Environment PATH..."
    $UserPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
    
    # Check if the install directory is already in the User PATH
    if ($UserPath -split ';' -notcontains $INSTALL_DIR) {
        Log-Info "Adding $INSTALL_DIR to User PATH..."
        $NewPath = "$UserPath;$INSTALL_DIR"
        # Remove any potential double semicolons
        $NewPath = $NewPath -replace ';;', ';'
        [System.Environment]::SetEnvironmentVariable("Path", $NewPath, [System.EnvironmentVariableTarget]::User)
        
        # Update the current session's path so the script could theoretically call it immediately
        $env:Path += ";$INSTALL_DIR"
        return $true
    }
    return $false
}

function Install-Binary($ZipPath) {
    Log-Info "Installing aws-doctor..."
    
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
    }

    $TargetExe = Join-Path $INSTALL_DIR $BINARY_NAME
    if (Test-Path $TargetExe) {
        Log-Info "Existing binary found. Renaming to allow replacement..."
        # This allows replacing the file even if aws-doctor is currently running
        Move-Item -Path $TargetExe -Destination "$TargetExe.old" -Force
    }

    Log-Info "Extracting to $INSTALL_DIR..."
    # Force overwrite existing files during extraction
    Expand-Archive -Path $ZipPath -DestinationPath $INSTALL_DIR -Force

    if (Test-Path "$TargetExe.old") {
        try { Remove-Item "$TargetExe.old" -ErrorAction SilentlyContinue } catch {}
    }

    Log-Info "Successfully installed to $TargetExe"
}

function Main {
    param([string]$VersionArg)
    Log-Info "AWS Doctor Installer (Windows)"
    Log-Info "=============================="
    
    $TmpDir = $env:TEMP
    if (-not $TmpDir) { $TmpDir = $PWD.Path }
    
    $Arch = Detect-Arch
    $Version = if ($VersionArg) { $VersionArg } else { Get-LatestVersion }
    
    Log-Info "Target Version: $Version"
    Log-Info "Architecture: $Arch"
    
    $ZipPath = Download-AndVerify -Version $Version -OS "windows" -Arch $Arch -TmpDir $TmpDir
    
    Install-Binary -ZipPath $ZipPath
    $PathUpdated = Update-Path
    
    Log-Info ""
    Log-Info "Installation complete!"
    if ($PathUpdated) {
        Log-Info "IMPORTANT: Please restart your terminal/shell to start using 'aws-doctor'."
    } else {
        Log-Info "You can now run 'aws-doctor' in your terminal."
    }
}

Main $args[0]