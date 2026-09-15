#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

require_root

BUNDLE_INPUT=${1:-}
[[ -n $BUNDLE_INPUT ]] && shift
PUBLIC_HOST=
HTTPS_PORT=443
SERVER_HOST=127.0.0.1
SERVER_PORT=3000
DATABASE_URL=
while (( $# > 0 )); do
  case "$1" in
    --public-host) PUBLIC_HOST=${2:-}; shift 2 ;;
    --https-port) HTTPS_PORT=${2:-}; shift 2 ;;
    --server-host) SERVER_HOST=${2:-}; shift 2 ;;
    --server-port) SERVER_PORT=${2:-}; shift 2 ;;
    --database-url) DATABASE_URL=${2:-}; shift 2 ;;
    --help|-h)
      echo "Использование: sudo $0 <архив-релиза-или-каталог> --public-host <домен-или-ip> --database-url <url> [--https-port 443] [--server-host 127.0.0.1] [--server-port 3000]"
      exit 0
      ;;
    *) echo "Неизвестный аргумент: $1" >&2; exit 1 ;;
  esac
done
if [[ -z $BUNDLE_INPUT || -z $PUBLIC_HOST || -z $DATABASE_URL ]]; then
  echo "Обязательны архив, --public-host и --database-url" >&2
  exit 1
fi
if [[ ! $PUBLIC_HOST =~ ^[A-Za-z0-9.:-]+$ || ! $SERVER_HOST =~ ^[A-Za-z0-9.:-]+$ ||
      ! $HTTPS_PORT =~ ^[0-9]+$ || ! $SERVER_PORT =~ ^[0-9]+$ ||
      $HTTPS_PORT -lt 1 || $HTTPS_PORT -gt 65535 || $SERVER_PORT -lt 1 || $SERVER_PORT -gt 65535 ||
      $DATABASE_URL == *$'\n'* || $DATABASE_URL == *$'\r'* ]]; then
  echo "Некорректный домен/IP, адрес или порт" >&2
  exit 1
fi
SERVER_EXT=
prepare_bundle_input "$BUNDLE_INPUT"
trap 'cleanup_bundle_input; [[ -z ${SERVER_EXT:-} ]] || rm -f -- "$SERVER_EXT"' EXIT

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y ca-certificates nginx openssl

id "$MYHEALTH_USER" >/dev/null 2>&1 || useradd --system --home-dir /var/lib/myhealth --shell /usr/sbin/nologin "$MYHEALTH_USER"
install -d -m 0755 "$MYHEALTH_ROOT" "$MYHEALTH_RELEASES" /etc/myhealth /etc/myhealth/pki
install -d -o "$MYHEALTH_USER" -g "$MYHEALTH_USER" -m 0750 /var/lib/myhealth

PKI_DIR=/etc/myhealth/pki
if [[ ! -f $PKI_DIR/ca.key || ! -f $PKI_DIR/ca.crt ]]; then
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:4096 -out "$PKI_DIR/ca.key"
  openssl req -x509 -new -sha256 -days 3650 -key "$PKI_DIR/ca.key" \
    -out "$PKI_DIR/ca.crt" -subj "/CN=MyHealth Private CA"
fi
chmod 0600 "$PKI_DIR/ca.key"
chmod 0644 "$PKI_DIR/ca.crt"

SAN_TYPE=DNS
[[ $PUBLIC_HOST == *:* || $PUBLIC_HOST =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] && SAN_TYPE=IP
SERVER_EXT=$(mktemp)
printf '%s\n' \
  'basicConstraints=critical,CA:FALSE' \
  'keyUsage=critical,digitalSignature,keyEncipherment' \
  'extendedKeyUsage=serverAuth' \
  "subjectAltName=$SAN_TYPE:$PUBLIC_HOST" > "$SERVER_EXT"
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$PKI_DIR/server.key"
openssl req -new -sha256 -key "$PKI_DIR/server.key" -out "$PKI_DIR/server.csr" -subj "/CN=$PUBLIC_HOST"
openssl x509 -req -sha256 -days 397 -in "$PKI_DIR/server.csr" \
  -CA "$PKI_DIR/ca.crt" -CAkey "$PKI_DIR/ca.key" -CAcreateserial \
  -extfile "$SERVER_EXT" -out "$PKI_DIR/server.crt"
rm -f "$PKI_DIR/server.csr"
chmod 0600 "$PKI_DIR/server.key"
chmod 0644 "$PKI_DIR/server.crt"

verify_bundles "$BUNDLE_DIR"
VERSION=$(read_release_version "$BUNDLE_DIR")
install_release "$BUNDLE_DIR" "$VERSION"

install -m 0644 "$SCRIPT_DIR/templates/myhealth.service" /etc/systemd/system/myhealth.service
write_service_override "$DATABASE_URL" "$SERVER_HOST" "$SERVER_PORT"

sed \
  -e "s|__PUBLIC_HOST__|$PUBLIC_HOST|g" \
  -e "s|__HTTPS_PORT__|$HTTPS_PORT|g" \
  "$SCRIPT_DIR/templates/nginx.conf" > /etc/nginx/sites-available/myhealth
ln -sfn /etc/nginx/sites-available/myhealth /etc/nginx/sites-enabled/myhealth
rm -f /etc/nginx/sites-enabled/default

systemctl daemon-reload
systemctl enable --now myhealth.service
nginx -t
systemctl enable --now nginx.service
systemctl reload nginx.service

"$SCRIPT_DIR/install-tools.sh"
echo "MyHealth $VERSION установлен: https://$PUBLIC_HOST:$HTTPS_PORT"
echo "Теперь выпустите клиентский сертификат: sudo $SCRIPT_DIR/generate-client-cert.sh \"Имя пользователя\""
