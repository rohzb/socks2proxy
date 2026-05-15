#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# install-hardening.sh
# -----------------------------------------------------------------------------
# Author: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
# License: MIT
#
# Purpose:
#   Install socks2proxy with the hardened systemd unit and AppArmor profile.
#
# Usage:
#   sudo ./platform/linux/scripts/install-hardening.sh
#
# Environment overrides:
#   BINARY_SRC, BINARY_DST, CONFIG_SRC, CONFIG_DIR, CONFIG_DST,
#   UNIT_SRC, UNIT_DST, APPARMOR_SRC, APPARMOR_DST
# -----------------------------------------------------------------------------
set -euo pipefail

SERVICE_USER="socks2proxy"
SERVICE_GROUP="socks2proxy"
BINARY_SRC="${BINARY_SRC:-./socks2proxy}"
BINARY_DST="${BINARY_DST:-/usr/local/bin/socks2proxy}"
CONFIG_SRC="${CONFIG_SRC:-./config.yaml}"
CONFIG_DIR="${CONFIG_DIR:-/usr/local/etc/socks2proxy}"
CONFIG_DST="${CONFIG_DST:-/usr/local/etc/socks2proxy/config.yaml}"
UNIT_SRC="${UNIT_SRC:-./platform/linux/systemd/socks2proxy.hardened.service}"
UNIT_DST="${UNIT_DST:-/etc/systemd/system/socks2proxy.service}"
APPARMOR_SRC="${APPARMOR_SRC:-./platform/linux/apparmor/usr.local.bin.socks2proxy}"
APPARMOR_DST="${APPARMOR_DST:-/etc/apparmor.d/usr.local.bin.socks2proxy}"

log() { printf '[install] %s\n' "$*"; }
err() { printf '[install][error] %s\n' "$*" >&2; }

require_root() {
  if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
    err "run as root (use sudo)"
    exit 1
  fi
}

require_file() {
  local f="$1"
  if [[ ! -f "$f" ]]; then
    err "required file not found: $f"
    exit 1
  fi
}

ensure_service_user() {
  if id -u "$SERVICE_USER" >/dev/null 2>&1; then
    log "service user exists: $SERVICE_USER"
    return
  fi
  log "creating service user: $SERVICE_USER"
  useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin "$SERVICE_USER"
}

install_binary() {
  log "installing binary: $BINARY_SRC -> $BINARY_DST"
  install -m 0755 "$BINARY_SRC" "$BINARY_DST"
}

install_config() {
  log "installing config directory: $CONFIG_DIR"
  install -d -m 0750 "$CONFIG_DIR"
  log "installing config: $CONFIG_SRC -> $CONFIG_DST"
  install -m 0640 "$CONFIG_SRC" "$CONFIG_DST"
  chown root:"$SERVICE_GROUP" "$CONFIG_DST"
}

install_apparmor() {
  require_file "$APPARMOR_SRC"
  log "installing AppArmor profile: $APPARMOR_DST"
  install -m 0644 "$APPARMOR_SRC" "$APPARMOR_DST"
  if command -v apparmor_parser >/dev/null 2>&1; then
    log "loading AppArmor profile"
    apparmor_parser -r "$APPARMOR_DST"
  else
    log "apparmor_parser not found; skipping profile load"
  fi
}

install_systemd_unit() {
  require_file "$UNIT_SRC"
  log "installing systemd unit: $UNIT_DST"
  install -m 0644 "$UNIT_SRC" "$UNIT_DST"
  log "reloading systemd daemon"
  systemctl daemon-reload
  log "enabling and starting socks2proxy.service"
  systemctl enable --now socks2proxy.service
}

print_status() {
  log "service status"
  systemctl status socks2proxy.service --no-pager || true
}

main() {
  require_root
  require_file "$BINARY_SRC"
  require_file "$CONFIG_SRC"
  ensure_service_user
  install_binary
  install_config
  install_apparmor
  install_systemd_unit
  print_status
  log "completed"
}

main "$@"
