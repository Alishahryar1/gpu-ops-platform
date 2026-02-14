# GPU Ops Platform Build Script (PowerShell)

param(
    [switch]$Help = $false,
    [switch]$Clean = $false,
    [switch]$Daemon = $false,
    [switch]$CLI = $false,
    [switch]$Test = $false,
    [switch]$Install = $false,
    [switch]$RunDaemon = $false,
    [switch]$PolicyTest = $false,
    [string]$CLIArgs = ""
)

$ErrorActionPreference = "Stop"

$BINARY_DIR = "bin"
$DAEMON_BINARY = Join-Path $BINARY_DIR "gputld.exe"
$CLI_BINARY = Join-Path $BINARY_DIR "gputl.exe"

function Show-Help {
    Write-Host "GPU Ops Platform Build Script" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [options]" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Yellow
    Write-Host "  -Help           Show this help message"
    Write-Host "  -Clean          Clean build artifacts"
    Write-Host "  -Daemon         Build only the daemon"
    Write-Host "  -CLI            Build only the CLI"
    Write-Host "  -Test           Run tests"
    Write-Host "  -Install        Install binaries to PATH"
    Write-Host "  -RunDaemon      Run the daemon"
    Write-Host "  -PolicyTest     Test the policy engine"
    Write-Host "  -CLIArgs 'args' Arguments to pass to CLI"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1                  # Build all"
    Write-Host "  .\build.ps1 -Daemon -RunDaemon  # Build and run daemon"
    Write-Host "  .\build.ps1 -CLI -CLIArgs 'status'  # Build and run CLI status command"
    Write-Host "  .\build.ps1 -PolicyTest      # Test policy engine"
}

function Remove-BuildArtifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
    if (Test-Path $BINARY_DIR) {
        Remove-Item -Recurse -Force $BINARY_DIR
    }
    go clean
    Write-Host "Clean complete." -ForegroundColor Green
}

function Build-Daemon {
    Write-Host "Building gputld daemon..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path $BINARY_DIR | Out-Null
    go build -o $DAEMON_BINARY ./cmd/gputld
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Daemon built: $DAEMON_BINARY" -ForegroundColor Green
    } else {
        throw "Failed to build daemon"
    }
}

function Build-CLI {
    Write-Host "Building gputl CLI..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path $BINARY_DIR | Out-Null
    go build -o $CLI_BINARY ./cmd/gputl
    if ($LASTEXITCODE -eq 0) {
        Write-Host "CLI built: $CLI_BINARY" -ForegroundColor Green
    } else {
        throw "Failed to build CLI"
    }
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Yellow
    go test -v ./...
}

function Install-Binaries {
    Write-Host "Installing to local PATH..." -ForegroundColor Yellow
    # On Windows, add the binary directory to PATH or copy to a location in PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $fullBinPath = Join-Path $PSScriptRoot $BINARY_DIR

    if ($userPath -notlike "*$fullBinPath*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$fullBinPath", "User")
        Write-Host "Added $fullBinPath to user PATH" -ForegroundColor Green
        Write-Host "Please restart your terminal to use the new PATH setting" -ForegroundColor Yellow
    } else {
        Write-Host "Binary directory already in PATH" -ForegroundColor Green
    }
}

function Run-Daemon {
    Write-Host "Starting daemon..." -ForegroundColor Yellow
    if (Test-Path $DAEMON_BINARY) {
        & $DAEMON_BINARY
    } else {
        Write-Host "Daemon binary not found. Building..." -ForegroundColor Yellow
        Build-Daemon
        & $DAEMON_BINARY
    }
}

function Run-PolicyTest {
    Write-Host "Testing policy engine..." -ForegroundColor Yellow

    # Check for Python
    $pythonCmd = Get-Command python -ErrorAction SilentlyContinue
    $python3Cmd = Get-Command python3 -ErrorAction SilentlyContinue

    if (!$pythonCmd -and !$python3Cmd) {
        Write-Host "Python not found. Please install Python 3.10+ from https://www.python.org/downloads/" -ForegroundColor Red
        return
    }

    # Check for UV
    $uvCmd = Get-Command uv -ErrorAction SilentlyContinue
    if (!$uvCmd) {
        Write-Host "UV not found. Please install UV from https://github.com/astral-sh/uv" -ForegroundColor Red
        Write-Host "Run: pip install uv" -ForegroundColor Yellow
        return
    }

    # Set up Python command
    $py = if ($pythonCmd) { "python" } else { "python3" }

    # Install dependencies with UV
    Write-Host "Installing dependencies with UV..." -ForegroundColor Cyan
    Push-Location python
    uv sync
    Pop-Location

    # Run the engine
    Write-Host "Running policy engine..." -ForegroundColor Cyan
    Push-Location python
    & $py -m starlark_engine.engine --load-all
    Pop-Location
}

# Main script
if ($Help) {
    Show-Help
    exit 0
}

$buildFlag = !$Daemon -and !$CLI -and !$Test -and !$RunDaemon -and !$PolicyTest

if ($Clean) {
    Remove-BuildArtifacts
}

if ($buildFlag) {
    Build-Daemon
    Build-CLI
} else {
    if ($Daemon) { Build-Daemon }
    if ($CLI) { Build-CLI }
    if ($Test) { Run-Tests }
}

if ($Install) {
    Install-Binaries
}

if ($RunDaemon) {
    Run-Daemon
}

if ($PolicyTest) {
    Run-PolicyTest
}
