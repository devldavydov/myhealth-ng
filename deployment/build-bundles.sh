#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
PROJECT_DIR=$(cd -- "$SCRIPT_DIR/.." && pwd)
CALLER_DIR=$PWD
VERSION=${1:-$(date -u +%Y%m%d%H%M%S)}
OUTPUT_ARCHIVE=${2:-$PROJECT_DIR/artifacts/myhealth-$VERSION.tar.gz}

if [[ ! $VERSION =~ ^[A-Za-z0-9._-]+$ ]]; then
  echo "Некорректная версия: допустимы буквы, цифры, точка, дефис и подчёркивание" >&2
  exit 1
fi
if [[ $OUTPUT_ARCHIVE != /* ]]; then
  OUTPUT_ARCHIVE=$CALLER_DIR/$OUTPUT_ARCHIVE
fi
if [[ $OUTPUT_ARCHIVE != *.tar.gz ]]; then
  echo "Выходной файл должен иметь расширение .tar.gz" >&2
  exit 1
fi
ARCHIVE_DIR_NAME=$(basename -- "$OUTPUT_ARCHIVE" .tar.gz)
if [[ -z $ARCHIVE_DIR_NAME ]]; then
  echo "Имя архива без расширения не должно быть пустым" >&2
  exit 1
fi

BUILD_DIR=$(mktemp -d)
trap 'rm -rf -- "$BUILD_DIR"' EXIT
PACKAGE_DIR=$BUILD_DIR/$ARCHIVE_DIR_NAME

cd "$PROJECT_DIR"
go test ./...
go vet ./...
npm --prefix client ci
npm --prefix client run test
npm --prefix client run build

mkdir -p "$BUILD_DIR/frontend" "$BUILD_DIR/backend" "$PACKAGE_DIR/deployment/templates"
cp -a "$PROJECT_DIR/client/dist/." "$BUILD_DIR/frontend/"

TARGET_GOOS=${MYHEALTH_GOOS:-linux}
TARGET_GOARCH=${MYHEALTH_GOARCH:-$(go env GOARCH)}
if [[ $TARGET_GOOS != linux || ! $TARGET_GOARCH =~ ^[a-z0-9]+$ ]]; then
  echo "Деплой поддерживает только Linux; некорректная платформа: $TARGET_GOOS/$TARGET_GOARCH" >&2
  exit 1
fi
CGO_ENABLED=0 GOOS=$TARGET_GOOS GOARCH=$TARGET_GOARCH go -C "$PROJECT_DIR" build \
  -trimpath -ldflags='-s -w' -o "$BUILD_DIR/backend/myhealth-server" ./cmd/server
printf '%s/%s\n' "$TARGET_GOOS" "$TARGET_GOARCH" > "$BUILD_DIR/backend/PLATFORM"

cp "$SCRIPT_DIR/install-server.sh" "$SCRIPT_DIR/upgrade-server.sh" \
  "$SCRIPT_DIR/generate-client-cert.sh" "$SCRIPT_DIR/backup-database.sh" \
  "$SCRIPT_DIR/restore-database.sh" "$SCRIPT_DIR/uninstall-server.sh" \
  "$SCRIPT_DIR/lib.sh" "$PACKAGE_DIR/deployment/"
cp "$SCRIPT_DIR/templates/nginx.conf" "$PACKAGE_DIR/deployment/templates/"
cp "$SCRIPT_DIR/templates/myhealth.service" "$PACKAGE_DIR/deployment/templates/"
cp "$SCRIPT_DIR/README.md" "$PACKAGE_DIR/deployment/"

printf '%s\n' "$VERSION" > "$BUILD_DIR/frontend/VERSION"
printf '%s\n' "$VERSION" > "$BUILD_DIR/backend/VERSION"
printf '%s\n' "$VERSION" > "$PACKAGE_DIR/release-version"

tar --sort=name --mtime='UTC 1970-01-01' --owner=0 --group=0 --numeric-owner \
  -czf "$PACKAGE_DIR/frontend-$VERSION.tar.gz" -C "$BUILD_DIR/frontend" .
tar --sort=name --mtime='UTC 1970-01-01' --owner=0 --group=0 --numeric-owner \
  -czf "$PACKAGE_DIR/backend-$VERSION.tar.gz" -C "$BUILD_DIR/backend" .

(
  cd "$PACKAGE_DIR"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
)

mkdir -p "$(dirname -- "$OUTPUT_ARCHIVE")"
tar --sort=name --mtime='UTC 1970-01-01' --owner=0 --group=0 --numeric-owner \
  -czf "$OUTPUT_ARCHIVE" -C "$BUILD_DIR" -- "$ARCHIVE_DIR_NAME"

echo "Релиз версии $VERSION создан: $OUTPUT_ARCHIVE"
