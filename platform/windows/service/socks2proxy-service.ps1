<#
.SYNOPSIS
  Register socks2proxy as a Windows service.

.DESCRIPTION
  Creates a Windows service entry using sc.exe with a configurable binary path
  and config file.

.NOTES
  Author: Ruslan Ovsyannikov
  License: MIT
#>

param(
  [string]$ServiceName = "socks2proxy",
  [string]$BinaryPath = "C:\\Program Files\\socks2proxy\\socks2proxy.exe",
  [string]$ConfigPath = "C:\\ProgramData\\socks2proxy\\config.yaml"
)

$bin = '"{0}" --config "{1}"' -f $BinaryPath, $ConfigPath
sc.exe create $ServiceName binPath= $bin start= auto
sc.exe description $ServiceName "SOCKS5 egress router"
Write-Host "Registered service $ServiceName"
