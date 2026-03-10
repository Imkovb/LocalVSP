$ErrorActionPreference = "Stop"
# configuration - change as needed
$archivePath = Join-Path $env:TEMP "localvsp.tar"
$ServerHost = Read-Host "Remote host or IP"   # raspberry host or IP address
$User = Read-Host "SSH user"
$Password = Read-Host "SSH password for $User@$ServerHost"   # ssh password (if not using key auth)
$HostKeyFingerprint = ""      # leave blank to skip verifying host key
$projectDir = $PSScriptRoot          # assume script sits in project root
# use the ssh user for the destination home directory by default
$remoteBase = "/home/${User}/localvsp"  # destination directory on raspberry
$remoteOwner = '{0}:{0}' -f $User

Write-Host "==> Creating archive of project" -ForegroundColor Cyan
if (Test-Path $archivePath) { Remove-Item -Force $archivePath }
Push-Location $projectDir
try {
    # exclude git directory to keep the archive small
    & "C:\Windows\System32\tar.exe" -cf $archivePath --exclude=.git .
    if ($LASTEXITCODE -ne 0) { throw "tar failed" }
} finally { Pop-Location }
Write-Host "Archive: $archivePath"

Write-Host "`n==> Uploading" -ForegroundColor Cyan
# build pscp arguments; include hostkey only if provided
$pscpArgs = @('-batch', '-pw', $Password)
if ($HostKeyFingerprint) { $pscpArgs += '-hostkey'; $pscpArgs += $HostKeyFingerprint }
$pscpArgs += $archivePath
$pscpArgs += "${User}@${ServerHost}:/tmp/localvsp.tar"
& pscp @pscpArgs
if ($LASTEXITCODE -ne 0) { throw "Upload failed" }

Write-Host "`n==> Extracting on server" -ForegroundColor Cyan
$remoteCmd = @'
set -euo pipefail
printf '%s\n' '{0}' | sudo -S -p '' -v
sudo mkdir -p '{1}'
sudo tar -xf /tmp/localvsp.tar -C '{1}/'
# ensure ownership matches the ssh user (group assumed same as user)
sudo chown -R '{2}' '{1}'
sudo rm -f /tmp/localvsp.tar
ls -la '{1}'
'@ -f $Password, $remoteBase, $remoteOwner
$remoteCmd = $remoteCmd -replace "`r`n", "`n"
# prepare plink arguments similarly
$plinkArgs = @('-batch','-pw',$Password)
if ($HostKeyFingerprint) { $plinkArgs += '-hostkey'; $plinkArgs += $HostKeyFingerprint }
$plinkArgs += "${User}@${ServerHost}"
$plinkArgs += $remoteCmd
& plink @plinkArgs
if ($LASTEXITCODE -ne 0) { throw "Remote command failed" }

Write-Host "`n==> Done! Project copied to raspberry at $ServerHost." -ForegroundColor Green
