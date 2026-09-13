#!/usr/bin/env bash

MYHEALTH_ROOT=/opt/myhealth
MYHEALTH_RELEASES=$MYHEALTH_ROOT/releases
MYHEALTH_CURRENT=$MYHEALTH_ROOT/current
MYHEALTH_USER=myhealth
BUNDLE_TEMP_DIR=
BUNDLE_DIR=

require_root() {
  if (( EUID != 0 )); then
    echo "Скрипт необходимо запускать через sudo" >&2
    exit 1
  fi
}

prepare_bundle_input() {
  local input=$1
  local resolved_input

  [[ -n $input ]] || { echo "Не указан архив с релизом" >&2; exit 1; }
  resolved_input=$(realpath "$input")
  if [[ -d $resolved_input ]]; then
    BUNDLE_DIR=$resolved_input
    return
  fi
  if [[ ! -f $resolved_input || $resolved_input != *.tar.gz ]]; then
    echo "Ожидается каталог релиза или архив .tar.gz: $input" >&2
    exit 1
  fi

  BUNDLE_TEMP_DIR=$(mktemp -d /tmp/myhealth-bundles.XXXXXX)
  tar -xzf "$resolved_input" -C "$BUNDLE_TEMP_DIR" --no-same-owner
  BUNDLE_DIR=$BUNDLE_TEMP_DIR
}

cleanup_bundle_input() {
  if [[ -n $BUNDLE_TEMP_DIR && -d $BUNDLE_TEMP_DIR && $BUNDLE_TEMP_DIR == /tmp/myhealth-bundles.* ]]; then
    rm -rf -- "$BUNDLE_TEMP_DIR"
  fi
}

read_release_version() {
  local bundle_dir=$1
  local version_file=$bundle_dir/release-version
  [[ -f $version_file ]] || { echo "Нет файла $version_file" >&2; exit 1; }

  local version
  version=$(tr -d '\r\n' < "$version_file")
  [[ $version =~ ^[A-Za-z0-9._-]+$ ]] || { echo "Некорректная версия в release-version" >&2; exit 1; }
  printf '%s' "$version"
}

verify_bundles() {
  local bundle_dir=$1
  [[ -f $bundle_dir/SHA256SUMS ]] || { echo "Нет файла SHA256SUMS" >&2; exit 1; }
  (cd "$bundle_dir" && sha256sum --check --strict SHA256SUMS)
}

validate_staged_release() {
  local staging_dir=$1
  local expected_version=$2
  local frontend_version backend_version

  if [[ ! -f $staging_dir/client/index.html ]]; then
    echo "Frontend-бандл повреждён: отсутствует index.html" >&2
    return 1
  fi
  if [[ ! -f $staging_dir/server/server.cjs ]]; then
    if [[ -f $staging_dir/server/server.mjs ]]; then
      echo "Backend-бандл устарел: найден server.mjs вместо server.cjs. Пересоберите релиз текущим build-bundles.sh" >&2
    else
      echo "Backend-бандл повреждён: отсутствует server.cjs" >&2
    fi
    return 1
  fi
  if [[ ! -f $staging_dir/client/VERSION || ! -f $staging_dir/server/VERSION ]]; then
    echo "В бандле отсутствует VERSION" >&2
    return 1
  fi

  frontend_version=$(tr -d '\r\n' < "$staging_dir/client/VERSION")
  backend_version=$(tr -d '\r\n' < "$staging_dir/server/VERSION")
  if [[ $frontend_version != "$expected_version" || $backend_version != "$expected_version" ]]; then
    echo "Версии внутри бандлов не совпадают с release-version=$expected_version" >&2
    return 1
  fi

  node --check "$staging_dir/server/server.cjs" >/dev/null
}

install_release() {
  local bundle_dir=$1
  local version=$2
  local release_dir=$MYHEALTH_RELEASES/$version
  local staging_dir

  [[ -f $bundle_dir/frontend-$version.tar.gz ]] || { echo "Нет frontend-бандла" >&2; exit 1; }
  [[ -f $bundle_dir/backend-$version.tar.gz ]] || { echo "Нет backend-бандла" >&2; exit 1; }
  [[ ! -e $release_dir ]] || { echo "Версия $version уже установлена" >&2; exit 1; }

  staging_dir=$(mktemp -d "$MYHEALTH_RELEASES/.staging.XXXXXX")
  trap 'rm -rf -- "$staging_dir"' RETURN
  mkdir -p "$staging_dir/client" "$staging_dir/server"
  tar -xzf "$bundle_dir/frontend-$version.tar.gz" -C "$staging_dir/client" --no-same-owner
  tar -xzf "$bundle_dir/backend-$version.tar.gz" -C "$staging_dir/server" --no-same-owner

  if ! validate_staged_release "$staging_dir" "$version"; then
    rm -rf -- "$staging_dir"
    trap - RETURN
    return 1
  fi

  chown -R root:root "$staging_dir"
  find "$staging_dir" -type d -exec chmod 0755 {} +
  find "$staging_dir" -type f -exec chmod 0644 {} +
  mv "$staging_dir" "$release_dir"
  trap - RETURN

  ln -sfn "$release_dir" "$MYHEALTH_ROOT/current.next"
  mv -Tf "$MYHEALTH_ROOT/current.next" "$MYHEALTH_CURRENT"
}
