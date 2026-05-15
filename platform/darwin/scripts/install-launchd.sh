#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# install-launchd.sh
# -----------------------------------------------------------------------------
# Author: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
# License: MIT
#
# Purpose:
#   Install and load a launchd service for socks2proxy on macOS.
#
# Usage:
#   sudo ./platform/darwin/scripts/install-launchd.sh
# -----------------------------------------------------------------------------
set -euo pipefail

PLIST_SRC="${PLIST_SRC:-./platform/darwin/launchd/org.socks2proxy.plist}"
PLIST_DST="${PLIST_DST:-/Library/LaunchDaemons/org.socks2proxy.plist}"

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  printf 'run as root (use sudo)\n' >&2
  exit 1
fi

install -m 0644 "$PLIST_SRC" "$PLIST_DST"
launchctl unload "$PLIST_DST" >/dev/null 2>&1 || true
launchctl load "$PLIST_DST"
printf 'installed and loaded %s\n' "$PLIST_DST"
