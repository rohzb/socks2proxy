<#
.SYNOPSIS
  Install and start socks2proxy Windows service.

.NOTES
  Author: Ruslan Ovsyannikov
  License: MIT
#>

param(
  [string]$ServiceName = "socks2proxy"
)

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Resolve-Path (Join-Path $scriptDir "..")
& (Join-Path $root "service\\socks2proxy-service.ps1") -ServiceName $ServiceName
Start-Service -Name $ServiceName
Write-Host "Installed and started $ServiceName"
