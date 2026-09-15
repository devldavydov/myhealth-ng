#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

require_root
BUNDLE_INPUT=${1:-}
[[ -n $BUNDLE_INPUT ]] && shift
DATABASE_URL=
SERVER_HOST=127.0.0.1
SERVER_PORT=3000
while (( $# > 0 )); do
  case "$1" in
    --database-url) DATABASE_URL=${2:-}; shift 2 ;;
    --server-host) SERVER_HOST=${2:-}; shift 2 ;;
    --server-port) SERVER_PORT=${2:-}; shift 2 ;;
    --help|-h)
      echo "Использование: sudo $0 <архив-релиза-или-каталог> [--database-url <url>] [--server-host 127.0.0.1] [--server-port 3000]"
      exit 0
      ;;
    *) echo "Неизвестный аргумент: $1" >&2; exit 1 ;;
  esac
done
[[ -n $BUNDLE_INPUT ]] || { echo "Не указан архив релиза" >&2; exit 1; }
if [[ ! $SERVER_HOST =~ ^[A-Za-z0-9.:-]+$ || ! $SERVER_PORT =~ ^[0-9]+$ ||
      $SERVER_PORT -lt 1 || $SERVER_PORT -gt 65535 ||
      $DATABASE_URL == *$'\n'* || $DATABASE_URL == *$'\r'* ]]; then
  echo "Некорректный адрес или порт backend" >&2
  exit 1
fi
if [[ ! -f $MYHEALTH_SERVICE_OVERRIDE && -z $DATABASE_URL ]]; then
  echo "Для первого обновления на PostgreSQL обязателен --database-url" >&2
  exit 1
fi
prepare_bundle_input "$BUNDLE_INPUT"
SERVICE_BACKUP=
SERVICE_EXISTED=false
OVERRIDE_BACKUP=
OVERRIDE_EXISTED=false
trap 'cleanup_bundle_input; [[ -z ${SERVICE_BACKUP:-} ]] || rm -f -- "$SERVICE_BACKUP"; [[ -z ${OVERRIDE_BACKUP:-} ]] || rm -f -- "$OVERRIDE_BACKUP"' EXIT

verify_bundles "$BUNDLE_DIR"
VERSION=$(read_release_version "$BUNDLE_DIR")
PREVIOUS_RELEASE=$(readlink -f "$MYHEALTH_CURRENT" 2>/dev/null || true)
SERVICE_FILE=/etc/systemd/system/myhealth.service
SERVICE_TEMPLATE=$BUNDLE_DIR/deployment/templates/myhealth.service
[[ -f $SERVICE_TEMPLATE ]] || SERVICE_TEMPLATE=$SCRIPT_DIR/templates/myhealth.service
[[ -f $SERVICE_TEMPLATE ]] || { echo "В релизе отсутствует шаблон myhealth.service" >&2; exit 1; }
SERVICE_BACKUP=$(mktemp /tmp/myhealth-service.XXXXXX)
if [[ -f $SERVICE_FILE ]]; then
  cp -- "$SERVICE_FILE" "$SERVICE_BACKUP"
  SERVICE_EXISTED=true
fi
if [[ -n $DATABASE_URL ]]; then
  OVERRIDE_BACKUP=$(mktemp /tmp/myhealth-service-override.XXXXXX)
  if [[ -f $MYHEALTH_SERVICE_OVERRIDE ]]; then
    cp -- "$MYHEALTH_SERVICE_OVERRIDE" "$OVERRIDE_BACKUP"
    OVERRIDE_EXISTED=true
  fi
fi
install_release "$BUNDLE_DIR" "$VERSION"
install -m 0644 "$SERVICE_TEMPLATE" "$SERVICE_FILE"
if [[ -n $DATABASE_URL ]]; then
  write_service_override "$DATABASE_URL" "$SERVER_HOST" "$SERVER_PORT"
fi
systemctl daemon-reload

if ! systemctl restart myhealth.service; then
  echo "Новая версия не запустилась, возвращаю предыдущую" >&2
  if [[ -n $PREVIOUS_RELEASE && -d $PREVIOUS_RELEASE ]]; then
    ln -sfn "$PREVIOUS_RELEASE" "$MYHEALTH_ROOT/current.rollback"
    mv -Tf "$MYHEALTH_ROOT/current.rollback" "$MYHEALTH_CURRENT"
  fi
  if [[ $SERVICE_EXISTED == true ]]; then
    install -m 0644 "$SERVICE_BACKUP" "$SERVICE_FILE"
  else
    rm -f -- "$SERVICE_FILE"
  fi
  if [[ -n $OVERRIDE_BACKUP && $OVERRIDE_EXISTED == true ]]; then
    install -m 0600 "$OVERRIDE_BACKUP" "$MYHEALTH_SERVICE_OVERRIDE"
  elif [[ -n $OVERRIDE_BACKUP ]]; then
    rm -f -- "$MYHEALTH_SERVICE_OVERRIDE"
  fi
  systemctl daemon-reload
  systemctl restart myhealth.service || true
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
