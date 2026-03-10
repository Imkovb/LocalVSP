param(
    [string]$ServerHost = "",
    [string]$User = "vsp",
    [string]$Password = "",
    [string]$HostKeyFingerprint = "",
    [string]$RemoteDir = "/opt/localvsp"
)

$ErrorActionPreference = "Stop"

function Step([string]$Message) {
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Ensure-Command([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH."
    }
}

function Invoke-External([scriptblock]$Command, [string]$FailureMessage) {
    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw $FailureMessage
    }
}

if ([string]::IsNullOrWhiteSpace($ServerHost)) {
    throw "Provide -ServerHost with the remote host or IP address."
}

if ([string]::IsNullOrWhiteSpace($Password)) {
    $Password = Read-Host "SSH password for $User@$ServerHost"
}

Step "Checking local tools"
Ensure-Command "tar"

$workspaceRoot = Split-Path -Parent $PSCommandPath
$archivePath = Join-Path $env:TEMP "localvsp-deploy.tar"
$remoteArchive = "/tmp/localvsp-deploy.tar"

Step "Creating project archive"
if (Test-Path $archivePath) {
    Remove-Item -Force $archivePath
}

Push-Location $workspaceRoot
try {
    & tar -cf $archivePath --exclude ".git" --exclude "data" --exclude ".vscode" .
} finally {
    Pop-Location
}

$remoteCmd = @'
set -euo pipefail;
printf '%s\n' '__PASSWORD__' | sudo -S -p '' -v;
sudo mkdir -p '__REMOTE_DIR__';
sudo find '__REMOTE_DIR__' -mindepth 1 -maxdepth 1 ! -name 'data' ! -name '.env' -exec rm -rf {} +;
sudo tar -xf /tmp/localvsp-deploy.tar -C '__REMOTE_DIR__';
sudo rm -f /tmp/localvsp-deploy.tar;
if sudo test -f '__REMOTE_DIR__/.env' && sudo grep -q '^CLOUDFLARE_TUNNEL_TOKEN=' '__REMOTE_DIR__/.env' && [ -n "$(sudo grep '^CLOUDFLARE_TUNNEL_TOKEN=' '__REMOTE_DIR__/.env' | cut -d= -f2-)" ]; then
    PROFILES='--profile cloudflare';
else
    PROFILES='';
fi;
sudo docker compose --project-directory '__REMOTE_DIR__' --env-file '__REMOTE_DIR__/.env' $PROFILES up --build -d --remove-orphans;
sudo docker compose --project-directory '__REMOTE_DIR__' --env-file '__REMOTE_DIR__/.env' ps
'@.Replace("__REMOTE_DIR__", $RemoteDir).Replace("__PASSWORD__", $Password)

$remoteCmd = $remoteCmd -replace "`r`n", "`n"

$pscp = Get-Command "pscp" -ErrorAction SilentlyContinue
$plink = Get-Command "plink" -ErrorAction SilentlyContinue

if ($pscp -and $plink) {
    Step "Uploading archive with pscp (password is used automatically)"
    $pscpArgs = @("-batch")
    if ($HostKeyFingerprint) {
        $pscpArgs += @("-hostkey", $HostKeyFingerprint)
    }
    $pscpArgs += @("-pw", $Password, $archivePath, "$User@$ServerHost`:$remoteArchive")
    Invoke-External { pscp @pscpArgs } "Upload failed with pscp."

    Step "Deploying and rebuilding on remote server with plink"
    $plinkArgs = @("-batch")
    if ($HostKeyFingerprint) {
        $plinkArgs += @("-hostkey", $HostKeyFingerprint)
    }
    $plinkArgs += @("-pw", $Password, "$User@$ServerHost", $remoteCmd)
    Invoke-External { plink @plinkArgs } "Remote deploy failed with plink."
} else {
    Ensure-Command "scp"
    Ensure-Command "ssh"

    Step "Uploading archive with scp"
    Write-Host "PuTTY tools not found; scp/ssh will prompt for password." -ForegroundColor Yellow
    Invoke-External { scp $archivePath "$User@$ServerHost`:$remoteArchive" } "Upload failed with scp."

    Step "Deploying and rebuilding on remote server with ssh"
    Invoke-External { ssh "$User@$ServerHost" $remoteCmd } "Remote deploy failed with ssh."
}

Step "Done"
Write-Host "Deployment and rebuild completed on $User@$ServerHost in $RemoteDir" -ForegroundColor Green
