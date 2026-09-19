#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

BACKUP_FILE=
DATABASE_URL=
ASSUME_YES=false
while (( $# > 0 )); do
  case "$1" in
    --database-url)
      (( $# >= 2 )) && [[ $2 != --* ]] || { echo "Для $1 требуется значение" >&2; exit 1; }
      DATABASE_URL=$2
      shift 2
      ;;
    --yes|-y)
      ASSUME_YES=true
      shift
      ;;
    --help|-h)
      echo "Использование: sudo $0 <файл.dump> --database-url <url> [--yes]"
      echo "Восстановление заменяет объекты и данные целевой базы содержимым backup."
      exit 0
      ;;
    --*) echo "Неизвестный аргумент: $1" >&2; exit 1 ;;
    *)
      [[ -z $BACKUP_FILE ]] || { echo "Неожиданный позиционный аргумент: $1" >&2; exit 1; }
      BACKUP_FILE=$1
      shift
      ;;
  esac
done
[[ -n $BACKUP_FILE ]] || { echo "Не указан файл backup" >&2; exit 1; }
[[ -n $DATABASE_URL ]] || { echo "Обязателен --database-url" >&2; exit 1; }
if [[ $DATABASE_URL == *$'\n'* || $DATABASE_URL == *$'\r'* ]]; then
  echo "Некорректная строка подключения PostgreSQL" >&2
  exit 1
fi

require_root
for command in pg_restore sha256sum; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Не найдена обязательная команда: $command. Установите совместимый postgresql-client" >&2
    exit 1
  }
done
if [[ $BACKUP_FILE != /* ]]; then
  BACKUP_FILE=$PWD/$BACKUP_FILE
fi
[[ -f $BACKUP_FILE ]] || { echo "Файл backup не найден: $BACKUP_FILE" >&2; exit 1; }
BACKUP_FILE=$(realpath "$BACKUP_FILE")
CHECKSUM_FILE=$BACKUP_FILE.sha256
[[ -f $CHECKSUM_FILE ]] || { echo "Не найден checksum: $CHECKSUM_FILE" >&2; exit 1; }
(
  cd -- "$(dirname -- "$BACKUP_FILE")"
  sha256sum -c "$(basename -- "$CHECKSUM_FILE")"
)
pg_restore --list "$BACKUP_FILE" >/dev/null
if [[ $ASSUME_YES != true ]]; then
  echo "Содержимое целевой базы будет заменено данными из: $BACKUP_FILE"
  read -r -p "Введите RESTORE для продолжения: " confirmation
  [[ $confirmation == RESTORE ]] || { echo "Восстановление отменено"; exit 1; }
fi

SERVICE_WAS_ACTIVE=false
finish() {
  local exit_code=$?
  trap - EXIT INT TERM
  if [[ $SERVICE_WAS_ACTIVE == true ]] && ! systemctl start myhealth.service; then
    echo "Не удалось запустить myhealth.service после восстановления" >&2
    exit_code=1
  fi
  exit "$exit_code"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet myhealth.service; then
  SERVICE_WAS_ACTIVE=true
  systemctl stop myhealth.service
fi

pg_restore \
  --dbname="$DATABASE_URL" \
  --clean \
  --if-exists \
  --no-owner \
  --no-privileges \
  --exit-on-error \
  --single-transaction \
  "$BACKUP_FILE"

echo "База PostgreSQL восстановлена из: $BACKUP_FILE"
