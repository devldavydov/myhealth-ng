#!/usr/bin/env bash
set -Eeuo pipefail

if (( EUID != 0 )); then
  echo "Скрипт необходимо запускать через sudo" >&2
  exit 1
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
TARGET_DIR=/usr/local/lib/myhealth-deployment

install -d -m 0755 "$TARGET_DIR/templates"
install -m 0755 "$SCRIPT_DIR/install-server.sh" "$SCRIPT_DIR/upgrade-server.sh" \
  "$SCRIPT_DIR/generate-client-cert.sh" "$SCRIPT_DIR/uninstall-server.sh" "$SCRIPT_DIR/install-tools.sh" "$TARGET_DIR/"
install -m 0644 "$SCRIPT_DIR/lib.sh" "$TARGET_DIR/lib.sh"
install -m 0644 "$SCRIPT_DIR/templates/nginx.conf" "$TARGET_DIR/templates/nginx.conf"
install -m 0644 "$SCRIPT_DIR/templates/myhealth.service" "$TARGET_DIR/templates/myhealth.service"
rm -f /usr/local/sbin/myhealth-upgrade /usr/local/sbin/myhealth-generate-client-cert /usr/local/sbin/myhealth-uninstall
