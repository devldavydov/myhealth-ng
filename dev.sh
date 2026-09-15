#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
DATABASE_URL='postgresql://myhealth:myhealth@127.0.0.1:5432/myhealth?sslmode=disable'
CLEANED_UP=false
COMPOSE_TOUCHED=false
BACKEND_PID=
FRONTEND_PID=

cleanup() {
  local exit_code=$?
  if [[ $CLEANED_UP == true ]]; then
    return
  fi
  CLEANED_UP=true
  trap - EXIT INT TERM
  set +e

  echo
  echo "Останавливаю локальное окружение MyHealth..."

  for pid in "$BACKEND_PID" "$FRONTEND_PID"; do
    if [[ -n $pid ]]; then
      kill -- "-$pid" 2>/dev/null
    fi
  done
  for pid in "$BACKEND_PID" "$FRONTEND_PID"; do
    if [[ -n $pid ]]; then
      wait "$pid" 2>/dev/null
    fi
  done

  if [[ $COMPOSE_TOUCHED == true ]]; then
    docker compose --project-directory "$SCRIPT_DIR" down --volumes --remove-orphans
  fi
  echo "Локальное окружение MyHealth остановлено и очищено."
  exit "$exit_code"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

for command in docker go npm setsid; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Не найдена обязательная команда: $command" >&2
    exit 1
  fi
done
if ! docker compose version >/dev/null 2>&1; then
  echo "Не установлен Docker Compose plugin" >&2
  exit 1
fi
if [[ ! -d $SCRIPT_DIR/client/node_modules ]]; then
  echo "Frontend-зависимости не установлены. Выполните: npm --prefix client install" >&2
  exit 1
fi

cd "$SCRIPT_DIR"
echo "Запускаю PostgreSQL..."
COMPOSE_TOUCHED=true
docker compose up -d --wait postgres

# LEGACY IMPORT: delete this block together with cmd/legacy-import,
# internal/legacyimport and legacy_data after the transition is complete.
if compgen -G "$SCRIPT_DIR/legacy_data/*.csv" >/dev/null; then
  echo "Загружаю локальные legacy-данные..."
  go run ./cmd/legacy-import \
    --database-url "$DATABASE_URL" \
    --data-dir "$SCRIPT_DIR/legacy_data" \
    --user-id '00000000-0000-4000-8000-000000000000'
else
  echo "Локальные legacy-выгрузки не найдены, импорт пропущен."
fi
# END LEGACY IMPORT

echo "Запускаю backend локально..."
setsid go run ./cmd/server \
  --database-url "$DATABASE_URL" &
BACKEND_PID=$!

echo "Запускаю frontend локально..."
setsid npm --prefix client run dev &
FRONTEND_PID=$!

echo "После запуска интерфейс будет доступен на http://localhost:5173."
echo "Нажмите Ctrl-C для полной остановки и очистки Docker."

set +e
wait -n "$BACKEND_PID" "$FRONTEND_PID"
PROCESS_EXIT_CODE=$?
set -e

if [[ $PROCESS_EXIT_CODE -ne 0 ]]; then
  echo "Backend или frontend завершился с ошибкой." >&2
else
  echo "Backend или frontend завершился; останавливаю окружение."
fi
exit "$PROCESS_EXIT_CODE"
