#!/usr/bin/env bash

MYHEALTH_ROOT=/opt/myhealth
MYHEALTH_RELEASES=$MYHEALTH_ROOT/releases
MYHEALTH_CURRENT=$MYHEALTH_ROOT/current
MYHEALTH_USER=myhealth
MYHEALTH_SERVICE_OVERRIDE=/etc/systemd/system/myhealth.service.d/10-config.conf
BUNDLE_TEMP_DIR=
BUNDLE_DIR=

systemd_quote() {
  local value=$1
  value=${value//\\/\\\\}
  value=${value//\"/\\\"}
  value=${value//\$/\$\$}
  value=${value//%/%%}
  printf '"%s"' "$value"
}

write_service_override() {
  local database_url=$1
  local server_host=$2
  local server_port=$3
  local database_argument host_argument port_argument

  database_argument=$(systemd_quote "$database_url")
  host_argument=$(systemd_quote "$server_host")
  port_argument=$(systemd_quote "$server_port")
  install -d -m 0755 "$(dirname -- "$MYHEALTH_SERVICE_OVERRIDE")"
  {
    echo '[Service]'
    echo 'ExecStart='
    printf 'ExecStart=/opt/myhealth/current/server/myhealth-server --host=%s --port=%s --database-url=%s --require-client-cert=true\n' \
      "$host_argument" "$port_argument" "$database_argument"
  } > "$MYHEALTH_SERVICE_OVERRIDE"
  chmod 0600 "$MYHEALTH_SERVICE_OVERRIDE"
}

require_root() {
  if (( EUID != 0 )); then
    echo "Скрипт необходимо запускать через sudo" >&2
    exit 1
  fi
}

prepare_bundle_input() {
  local input=$1
  local resolved_input
  local -a bundle_entries=()

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
  if [[ -f $BUNDLE_TEMP_DIR/release-version ]]; then
    BUNDLE_DIR=$BUNDLE_TEMP_DIR
    return
  fi

  mapfile -d '' -t bundle_entries < <(
    find "$BUNDLE_TEMP_DIR" -mindepth 1 -maxdepth 1 -print0
  )
  if (( ${#bundle_entries[@]} == 1 )) &&
      [[ -d ${bundle_entries[0]} && ! -L ${bundle_entries[0]} ]]; then
    BUNDLE_DIR=${bundle_entries[0]}
    return
  fi

  echo "Архив должен содержать один корневой каталог релиза" >&2
  rm -rf -- "$BUNDLE_TEMP_DIR"
  BUNDLE_TEMP_DIR=
  BUNDLE_DIR=
  exit 1
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
  if [[ ! -s $staging_dir/server/myhealth-server ]]; then
    echo "Backend-бандл повреждён: отсутствует бинарник myhealth-server" >&2
    return 1
  fi
  if [[ ! -f $staging_dir/client/VERSION || ! -f $staging_dir/server/VERSION || ! -f $staging_dir/server/PLATFORM ]]; then
    echo "В бандле отсутствует VERSION или PLATFORM" >&2
    return 1
  fi

  frontend_version=$(tr -d '\r\n' < "$staging_dir/client/VERSION")
  backend_version=$(tr -d '\r\n' < "$staging_dir/server/VERSION")
  if [[ $frontend_version != "$expected_version" || $backend_version != "$expected_version" ]]; then
    echo "Версии внутри бандлов не совпадают с release-version=$expected_version" >&2
    return 1
  fi

  local platform machine expected_arch
  platform=$(tr -d '\r\n' < "$staging_dir/server/PLATFORM")
  machine=$(uname -m)
  case "$machine" in
    x86_64) expected_arch=amd64 ;;
    aarch64|arm64) expected_arch=arm64 ;;
    i386|i486|i586|i686) expected_arch=386 ;;
    armv7l) expected_arch=arm ;;
    *) expected_arch=$machine ;;
  esac
  if [[ $platform != "linux/$expected_arch" ]]; then
    echo "Backend собран для $platform, сервер использует linux/$expected_arch" >&2
    return 1
  fi
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
  chmod 0755 "$staging_dir/server/myhealth-server"
  mv "$staging_dir" "$release_dir"
  trap - RETURN

  ln -sfn "$release_dir" "$MYHEALTH_ROOT/current.next"
  mv -Tf "$MYHEALTH_ROOT/current.next" "$MYHEALTH_CURRENT"
}
