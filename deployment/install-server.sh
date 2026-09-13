#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

require_root

BUNDLE_INPUT=${1:-}
PUBLIC_HOST=${2:-}
HTTPS_PORT=${3:-443}
if [[ -z $BUNDLE_INPUT || -z $PUBLIC_HOST ]]; then
  echo "Использование: sudo $0 <архив-релиза-или-каталог> <домен-или-ip> [https-порт]" >&2
  exit 1
fi
if [[ ! $PUBLIC_HOST =~ ^[A-Za-z0-9.:-]+$ || ! $HTTPS_PORT =~ ^[0-9]+$ ]]; then
  echo "Некорректный домен/IP или порт" >&2
  exit 1
fi
SERVER_EXT=
prepare_bundle_input "$BUNDLE_INPUT"
trap 'cleanup_bundle_input; [[ -z ${SERVER_EXT:-} ]] || rm -f -- "$SERVER_EXT"' EXIT

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y ca-certificates curl gnupg nginx openssl

NODE_MAJOR=$(node -p 'Number(process.versions.node.split(".")[0])' 2>/dev/null || printf '0')
if (( NODE_MAJOR < 20 )); then
  install -d -m 0755 /etc/apt/keyrings
  curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key \
    | gpg --dearmor --yes -o /etc/apt/keyrings/nodesource.gpg
  chmod 0644 /etc/apt/keyrings/nodesource.gpg
  printf '%s\n' "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_22.x nodistro main" \
    > /etc/apt/sources.list.d/nodesource.list
  apt-get update
  apt-get install -y nodejs
fi

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

cat > /etc/myhealth/myhealth.env <<'ENV'
PORT=3000
HOST=127.0.0.1
REQUIRE_CLIENT_CERT=true
ENV
chmod 0640 /etc/myhealth/myhealth.env

cat > /etc/systemd/system/myhealth.service <<'UNIT'
[Unit]
Description=MyHealth API
After=network.target

[Service]
Type=simple
User=myhealth
Group=myhealth
WorkingDirectory=/var/lib/myhealth
EnvironmentFile=/etc/myhealth/myhealth.env
ExecStartPre=/usr/bin/test -r /opt/myhealth/current/server/server.cjs
ExecStart=/usr/bin/node /opt/myhealth/current/server/server.cjs
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/myhealth

[Install]
WantedBy=multi-user.target
UNIT

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
