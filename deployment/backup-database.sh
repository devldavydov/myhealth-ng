#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

OUTPUT_FILE=
DATABASE_URL=
while (( $# > 0 )); do
  case "$1" in
    --database-url)
      (( $# >= 2 )) && [[ $2 != --* ]] || { echo "Для $1 требуется значение" >&2; exit 1; }
      DATABASE_URL=$2
      shift 2
      ;;
    --help|-h)
      echo "Использование: sudo $0 [файл.dump] --database-url <url>"
      echo "По умолчанию backup создаётся в /var/backups/myhealth."
      exit 0
      ;;
    --*) echo "Неизвестный аргумент: $1" >&2; exit 1 ;;
    *)
      [[ -z $OUTPUT_FILE ]] || { echo "Неожиданный позиционный аргумент: $1" >&2; exit 1; }
      OUTPUT_FILE=$1
      shift
      ;;
  esac
done

[[ -n $DATABASE_URL ]] || { echo "Обязателен --database-url" >&2; exit 1; }
if [[ $DATABASE_URL == *$'\n'* || $DATABASE_URL == *$'\r'* ]]; then
  echo "Некорректная строка подключения PostgreSQL" >&2
  exit 1
fi
require_root
for command in pg_dump pg_restore sha256sum; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Не найдена обязательная команда: $command. Установите совместимый postgresql-client" >&2
    exit 1
  }
done
if [[ -z $OUTPUT_FILE ]]; then
  OUTPUT_FILE=/var/backups/myhealth/myhealth-$(date -u +%Y%m%dT%H%M%SZ).dump
fi
if [[ $OUTPUT_FILE != /* ]]; then
  OUTPUT_FILE=$PWD/$OUTPUT_FILE
fi
[[ ! -e $OUTPUT_FILE && ! -e $OUTPUT_FILE.sha256 ]] || {
  echo "Файл backup уже существует: $OUTPUT_FILE" >&2
  exit 1
}
OUTPUT_DIR=$(dirname -- "$OUTPUT_FILE")
[[ -d $OUTPUT_DIR ]] || install -d -m 0700 "$OUTPUT_DIR"
TEMP_DUMP=$(mktemp "$OUTPUT_DIR/.myhealth-backup.XXXXXX")
TEMP_CHECKSUM=$(mktemp "$OUTPUT_DIR/.myhealth-backup-checksum.XXXXXX")
cleanup() {
  rm -f -- "$TEMP_DUMP" "$TEMP_CHECKSUM"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

pg_dump \
  --dbname="$DATABASE_URL" \
  --format=custom \
  --no-owner \
  --no-privileges \
  --file="$TEMP_DUMP"
pg_restore --list "$TEMP_DUMP" >/dev/null
chmod 0600 "$TEMP_DUMP"
DUMP_HASH=$(sha256sum "$TEMP_DUMP" | awk '{print $1}')
printf '%s  %s\n' "$DUMP_HASH" "$(basename -- "$OUTPUT_FILE")" > "$TEMP_CHECKSUM"
chmod 0600 "$TEMP_CHECKSUM"
mv -- "$TEMP_DUMP" "$OUTPUT_FILE"
mv -- "$TEMP_CHECKSUM" "$OUTPUT_FILE.sha256"
trap - EXIT INT TERM

echo "Backup PostgreSQL создан: $OUTPUT_FILE"
echo "SHA-256: $OUTPUT_FILE.sha256"
