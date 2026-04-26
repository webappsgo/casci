# CASCI Windows Service Installer
# Run this script with PowerShell as Administrator

$ServiceName = "CASCI"
$DisplayName = "CASCI - CI/CD Application Server"
$Description = "CI/CD Application Server for Continuous Integration"
$BinaryName = "casci.exe"
$InstallDir = "$env:ProgramFiles\CASCI"
$DataDir = "$env:ProgramData\CASCI"
$LogDir = "$env:ProgramData\CASCI\logs"

Write-Host "Installing CASCI Windows Service..." -ForegroundColor Cyan

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "Error: This script must be run as Administrator" -ForegroundColor Red
    exit 1
}

# Find binary
$BinaryPath = $null
$SearchPaths = @(".\$BinaryName", ".\bin\$BinaryName", ".\binaries\$BinaryName")
foreach ($path in $SearchPaths) {
    if (Test-Path $path) {
        $BinaryPath = Resolve-Path $path
        break
    }
}

if (-not $BinaryPath) {
    Write-Host "Error: CASCI binary ($BinaryName) not found" -ForegroundColor Red
    exit 1
}

Write-Host "Found binary: $BinaryPath" -ForegroundColor Green

# Create directories
Write-Host "Creating directories..."
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
New-Item -ItemType Directory -Force -Path "$DataDir\artifacts" | Out-Null
New-Item -ItemType Directory -Force -Path "$DataDir\cache" | Out-Null

# Copy binary
Write-Host "Installing binary to $InstallDir..."
Copy-Item $BinaryPath "$InstallDir\$BinaryName" -Force

# Set environment variables for the service
$env:CASCI_DB_TYPE = "sqlite"
$env:CASCI_DB_DSN = "$DataDir\casci.db"

# Check if service already exists
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Host "Service already exists. Stopping and removing..." -ForegroundColor Yellow
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Create Windows service using Virtual Service Account
Write-Host "Creating Windows service..."
$BinaryPathWithArgs = "`"$InstallDir\$BinaryName`""

# Create service with Virtual Service Account (NT SERVICE\CASCI)
$result = New-Service -Name $ServiceName `
    -BinaryPathName $BinaryPathWithArgs `
    -DisplayName $DisplayName `
    -Description $Description `
    -StartupType Automatic `
    -ErrorAction Stop

if ($result) {
    Write-Host "Service created successfully" -ForegroundColor Green
}

# Grant Virtual Service Account permissions to data directories
Write-Host "Setting directory permissions for NT SERVICE\$ServiceName..."
$acl = Get-Acl $DataDir
$permission = "NT SERVICE\$ServiceName","FullControl","ContainerInherit,ObjectInherit","None","Allow"
$accessRule = New-Object System.Security.AccessControl.FileSystemAccessRule $permission
$acl.SetAccessRule($accessRule)
Set-Acl $DataDir $acl

$acl = Get-Acl $LogDir
$acl.SetAccessRule($accessRule)
Set-Acl $LogDir $acl

# Configure service recovery options
Write-Host "Configuring service recovery options..."
sc.exe failure $ServiceName reset= 86400 actions= restart/10000/restart/10000/restart/30000

# Start service
Write-Host "Starting service..."
Start-Service -Name $ServiceName

# Wait a moment and check status
Start-Sleep -Seconds 2
$serviceStatus = Get-Service -Name $ServiceName

Write-Host ""
Write-Host "✓ Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Service commands:" -ForegroundColor Cyan
Write-Host "  Start:   Start-Service -Name $ServiceName"
Write-Host "  Stop:    Stop-Service -Name $ServiceName"
Write-Host "  Restart: Restart-Service -Name $ServiceName"
Write-Host "  Status:  Get-Service -Name $ServiceName"
Write-Host "  Logs:    Get-EventLog -LogName Application -Source $ServiceName -Newest 50"
Write-Host ""
Write-Host "Configuration:" -ForegroundColor Cyan
Write-Host "  Install directory: $InstallDir"
Write-Host "  Data directory:    $DataDir"
Write-Host "  Log directory:     $LogDir"
Write-Host "  Service account:   NT SERVICE\$ServiceName (Virtual Service Account)"
Write-Host ""
Write-Host "Current status: $($serviceStatus.Status)" -ForegroundColor $(if ($serviceStatus.Status -eq 'Running') { 'Green' } else { 'Yellow' })
Write-Host ""
Write-Host "The service is configured to start automatically on boot."
