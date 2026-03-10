$ErrorActionPreference = "Stop"
$archivePath = Join-Path $env:TEMP "localvsp-deploy.tar"
$ServerHost = Read-Host "Remote host or IP"
$User = "vsp"
$Password = Read-Host "SSH password for $User@$ServerHost"
$HostKeyFingerprint = ""
$RemoteDir = "/opt/localvsp"
$remoteArchive = "/tmp/localvsp-deploy.tar"
$workspaceRoot = Split-Path -Parent $PSCommandPath

Write-Host "==> Creating archive" -ForegroundColor Cyan
if (Test-Path $archivePath) { Remove-Item -Force $archivePath }
Push-Location $workspaceRoot
try {
    & "C:\Windows\System32\tar.exe" -cf $archivePath --exclude ".git" --exclude "data" --exclude ".vscode" .
    if ($LASTEXITCODE -ne 0) { throw "tar failed" }
} finally { Pop-Location }
Write-Host "Archive: $archivePath ($(((Get-Item $archivePath).Length / 1MB).ToString('0.0')) MB)"

Write-Host "`n==> Uploading with pscp" -ForegroundColor Cyan
$pscpArgs = @('-batch')
if ($HostKeyFingerprint) { $pscpArgs += '-hostkey'; $pscpArgs += $HostKeyFingerprint }
$pscpArgs += '-pw'; $pscpArgs += $Password
$pscpArgs += $archivePath
$pscpArgs += "${User}@${ServerHost}:${remoteArchive}"
& pscp @pscpArgs
if ($LASTEXITCODE -ne 0) { throw "Upload failed" }

Write-Host "`n==> Deploying on server" -ForegroundColor Cyan
$remoteCmd = @"
set -euo pipefail
printf '%s\n' '$Password' | sudo -S -p '' -v
sudo mkdir -p $RemoteDir
sudo find $RemoteDir -mindepth 1 -maxdepth 1 ! -name 'data' ! -name '.env' -exec rm -rf {} +
sudo tar -xf $remoteArchive -C $RemoteDir
sudo rm -f $remoteArchive
if sudo test -f $RemoteDir/.env && sudo grep -q '^CLOUDFLARE_TUNNEL_TOKEN=' $RemoteDir/.env && [ -n "`$(sudo grep '^CLOUDFLARE_TUNNEL_TOKEN=' $RemoteDir/.env | cut -d= -f2-)" ]; then
    PROFILES='--profile cloudflare'
else
    PROFILES=''
fi
sudo docker compose --project-directory $RemoteDir --env-file $RemoteDir/.env `$PROFILES up --build -d --remove-orphans
sudo docker compose --project-directory $RemoteDir --env-file $RemoteDir/.env ps
"@
$remoteCmd = $remoteCmd -replace "`r`n", "`n"
$plinkArgs = @('-batch')
if ($HostKeyFingerprint) { $plinkArgs += '-hostkey'; $plinkArgs += $HostKeyFingerprint }
$plinkArgs += '-pw'; $plinkArgs += $Password
$plinkArgs += "${User}@${ServerHost}"
$plinkArgs += $remoteCmd
& plink @plinkArgs
if ($LASTEXITCODE -ne 0) { throw "Remote deploy failed" }

Write-Host "`n==> Done!" -ForegroundColor Green
