param(
    [String]$Version,
    [Boolean]$Proxy = $False
)


$Owner = "iyear"
$Repo = "tdl"
$Location = "$Env:SystemDrive\tdl"

$ErrorActionPreference = "Stop"

# check if run as admin
if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]"Administrator"))
{
    Write-Host "Please run this script as Administrator" -ForegroundColor Red
    exit 1
}

# use proxy if argument is passed
$PROXY_PREFIX = ""
if ($Proxy)
{
    $PROXY_PREFIX = "https://ghproxy.com/"
    Write-Host "Using GitHub proxy: $PROXY_PREFIX" -ForegroundColor Blue
}

# Set download ARCH based on system architecture
$Arch = ""
switch ($env:PROCESSOR_ARCHITECTURE)
{
    "AMD64" {
        $Arch = "64bit"
    }
    "x86" {
        $Arch = "32bit"
    }
    "ARM" {
        $Arch = "arm64"
    }
    default {
        Write-Host "Unsupported system architecture: $env:PROCESSOR_ARCHITECTURE" -ForegroundColor Red
        exit 1
    }
}

# set version
if (!$Version)
{
    $Version = (Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest").tag_name
}
Write-Host "Target version: $Version" -ForegroundColor Blue

# build download URL
$URL = "${PROXY_PREFIX}https://github.com/$Owner/$Repo/releases/download/$Version/${Repo}_Windows_$Arch.zip"
Write-Host "Downloading $Repo from $URL" -ForegroundColor Blue

# download and extract
Invoke-WebRequest -Uri $URL -OutFile "$Repo.zip"
# test zip path
if (-not(Test-Path "$Repo.zip"))
{
    Write-Host "Download $URL failed" -ForegroundColor Red
    exit 1
}
# only extract tdl.exe to $LOCATION , add to PATH and remove zip file
Expand-Archive -Path "$Repo.zip" -DestinationPath "$Location" -Force

# if $LOCATION has not been added to PATH yet, add it
if (-not($PathEnv -like "*$Location*"))
{
    Write-Host "Try to add $Location to Path Environment variable..." -ForegroundColor Blue

    $PathEnv = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine)
    $NewPath = $PathEnv + ";$Location"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, [EnvironmentVariableTarget]::Machine)
    # update current process' PATH
    [Environment]::SetEnvironmentVariable("Path", $NewPath, [EnvironmentVariableTarget]::Process)
}
# remove zip file
Remove-Item "$Repo.zip"

# test if installation is successful, and print instructions
if (-not(Get-Command $Repo -ErrorAction SilentlyContinue))
{
    Write-Host "Installation failed" -ForegroundColor Red
    exit 1
}

Write-Host "$Repo installed successfully! Location: $Location" -ForegroundColor Green
Write-Host "Run '$Repo' to get started" -ForegroundColor Green
Write-Host "To get started with tdl, please visit https://github.com/iyear/tdl#quick-start" -ForegroundColor Green
Write-Host "Note: Updates to PATH might not be visible until you restart your terminal application or reboot machine" -ForegroundColor Yellow
