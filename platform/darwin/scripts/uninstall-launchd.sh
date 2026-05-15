#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# uninstall-launchd.sh
# -----------------------------------------------------------------------------
# Author: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
# License: MIT
#
# Purpose:
#   Unload and remove a launchd service for socks2proxy on macOS.
#
# Usage:
#   sudo ./platform/darwin/scripts/uninstall-launchd.sh
# -----------------------------------------------------------------------------
set -euo pipefail

PLIST_DST="${PLIST_DST:-/Library/LaunchDaemons/org.socks2proxy.plist}"

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  printf 'run as root (use sudo)\n' >&2
  exit 1
fi

if [[ -f "$PLIST_DST" ]]; then
  launchctl unload "$PLIST_DST" >/dev/null 2>&1 || true
  rm -f "$PLIST_DST"
  printf 'removed %s\n' "$PLIST_DST"
else
  printf 'nothing to remove: %s\n' "$PLIST_DST"
fi
