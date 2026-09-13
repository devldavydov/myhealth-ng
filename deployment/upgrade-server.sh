#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

require_root
BUNDLE_INPUT=${1:-}
[[ -n $BUNDLE_INPUT ]] || { echo "Использование: sudo $0 <архив-релиза-или-каталог>" >&2; exit 1; }
prepare_bundle_input "$BUNDLE_INPUT"
trap 'cleanup_bundle_input' EXIT

verify_bundles "$BUNDLE_DIR"
VERSION=$(read_release_version "$BUNDLE_DIR")
PREVIOUS_RELEASE=$(readlink -f "$MYHEALTH_CURRENT" 2>/dev/null || true)
install_release "$BUNDLE_DIR" "$VERSION"

if ! systemctl restart myhealth.service; then
  echo "Новая версия не запустилась, возвращаю предыдущую" >&2
  if [[ -n $PREVIOUS_RELEASE && -d $PREVIOUS_RELEASE ]]; then
    ln -sfn "$PREVIOUS_RELEASE" "$MYHEALTH_ROOT/current.rollback"
    mv -Tf "$MYHEALTH_ROOT/current.rollback" "$MYHEALTH_CURRENT"
    systemctl restart myhealth.service
  fi
  exit 1
fi

# Миграция конфигураций, созданных до добавления UTF-8 в nginx-ответ 403.
NGINX_SITE=/etc/nginx/sites-available/myhealth
if [[ -f $NGINX_SITE ]] && ! grep -q 'charset utf-8;' "$NGINX_SITE"; then
  sed -i '/default_type application\/json;/a\        charset utf-8;' "$NGINX_SITE"
fi
if [[ -f $NGINX_SITE ]] && ! grep -q 'charset_types application/json;' "$NGINX_SITE"; then
  sed -i '/charset utf-8;/a\        charset_types application/json;' "$NGINX_SITE"
fi

nginx -t
systemctl reload nginx.service
echo "MyHealth обновлён до версии $VERSION"
if [[ -x $BUNDLE_DIR/deployment/install-tools.sh ]]; then
  "$BUNDLE_DIR/deployment/install-tools.sh"
elif [[ -x $SCRIPT_DIR/install-tools.sh ]]; then
  "$SCRIPT_DIR/install-tools.sh"
fi
