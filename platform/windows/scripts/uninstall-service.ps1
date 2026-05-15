<#
.SYNOPSIS
  Stop and remove socks2proxy Windows service.

.NOTES
  Author: Ruslan Ovsyannikov
  License: MIT
#>

param(
  [string]$ServiceName = "socks2proxy"
)

if (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) {
  Stop-Service -Name $ServiceName -ErrorAction SilentlyContinue
  sc.exe delete $ServiceName | Out-Null
  Write-Host "Removed service $ServiceName"
} else {
  Write-Host "Service $ServiceName not found"
}
