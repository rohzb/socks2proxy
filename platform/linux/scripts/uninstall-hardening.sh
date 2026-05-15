#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# uninstall-hardening.sh
# -----------------------------------------------------------------------------
# Author: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
# License: MIT
#
# Purpose:
#   Remove socks2proxy hardened system integration (unit and AppArmor profile).
#
# Usage:
#   sudo ./platform/linux/scripts/uninstall-hardening.sh
#
# Environment overrides:
#   BINARY_DST, CONFIG_DIR, UNIT_DST, APPARMOR_DST, KEEP_CONFIG, KEEP_USER
# -----------------------------------------------------------------------------
set -euo pipefail

SERVICE_NAME="socks2proxy.service"
SERVICE_USER="socks2proxy"
SERVICE_GROUP="socks2proxy"
BINARY_DST="${BINARY_DST:-/usr/local/bin/socks2proxy}"
CONFIG_DIR="${CONFIG_DIR:-/usr/local/etc/socks2proxy}"
UNIT_DST="${UNIT_DST:-/etc/systemd/system/socks2proxy.service}"
APPARMOR_DST="${APPARMOR_DST:-/etc/apparmor.d/usr.local.bin.socks2proxy}"
KEEP_CONFIG="${KEEP_CONFIG:-1}"
KEEP_USER="${KEEP_USER:-1}"

log() { printf '[uninstall] %s\n' "$*"; }
err() { printf '[uninstall][error] %s\n' "$*" >&2; }

require_root() {
  if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
    err "run as root (use sudo)"
    exit 1
  fi
}

stop_disable_service() {
  if systemctl list-unit-files | grep -q "^${SERVICE_NAME}"; then
    log "stopping and disabling ${SERVICE_NAME}"
    systemctl disable --now "$SERVICE_NAME" || true
  else
    log "service unit not registered: ${SERVICE_NAME}"
  fi
}

remove_systemd_unit() {
  if [[ -f "$UNIT_DST" ]]; then
    log "removing systemd unit: $UNIT_DST"
    rm -f "$UNIT_DST"
    systemctl daemon-reload
  fi
}

remove_apparmor() {
  if [[ -f "$APPARMOR_DST" ]]; then
    if command -v apparmor_parser >/dev/null 2>&1; then
      log "unloading AppArmor profile"
      apparmor_parser -R "$APPARMOR_DST" || true
    fi
    log "removing AppArmor profile: $APPARMOR_DST"
    rm -f "$APPARMOR_DST"
  fi
}

remove_binary() {
  if [[ -f "$BINARY_DST" ]]; then
    log "removing binary: $BINARY_DST"
    rm -f "$BINARY_DST"
  fi
}

remove_config() {
  if [[ "$KEEP_CONFIG" == "1" ]]; then
    log "keeping config directory (KEEP_CONFIG=1): $CONFIG_DIR"
    return
  fi
  if [[ -d "$CONFIG_DIR" ]]; then
    log "removing config directory: $CONFIG_DIR"
    rm -rf "$CONFIG_DIR"
  fi
}

remove_user() {
  if [[ "$KEEP_USER" == "1" ]]; then
    log "keeping service user (KEEP_USER=1): $SERVICE_USER"
    return
  fi
  if id -u "$SERVICE_USER" >/dev/null 2>&1; then
    log "removing service user: $SERVICE_USER"
    userdel "$SERVICE_USER" || true
  fi
  if getent group "$SERVICE_GROUP" >/dev/null 2>&1; then
    log "removing service group: $SERVICE_GROUP"
    groupdel "$SERVICE_GROUP" || true
  fi
}

main() {
  require_root
  stop_disable_service
  remove_systemd_unit
  remove_apparmor
  remove_binary
  remove_config
  remove_user
  log "completed"
}

main "$@"
