#!/usr/bin/env bash
set -Eeuo pipefail

if (( EUID != 0 )); then
  echo "Скрипт необходимо запускать через sudo: нужен доступ к ключу CA" >&2
  exit 1
fi

USERNAME=${1:-}
REQUESTED_UUID=${2:-}
OUTPUT_ROOT=${3:-$PWD/client-certificates}
if [[ -z $USERNAME || ${#USERNAME} -gt 64 || $USERNAME =~ [[:cntrl:]] ]]; then
  echo "Использование: sudo $0 <имя-пользователя> [uuid] [каталог-результата]" >&2
  echo "Имя должно содержать от 1 до 64 символов без управляющих символов" >&2
  exit 1
fi
if [[ -n $REQUESTED_UUID && ! $REQUESTED_UUID =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$ ]]; then
  echo "Некорректный UUID. Ожидается стандартный UUID" >&2
  exit 1
fi

PKI_DIR=/etc/myhealth/pki
[[ -r $PKI_DIR/ca.crt && -r $PKI_DIR/ca.key ]] || {
  echo "CA не найден. Сначала выполните install-server.sh" >&2
  exit 1
}

UUID=${REQUESTED_UUID:-$(< /proc/sys/kernel/random/uuid)}
UUID=${UUID,,}
DIRECTORY_USERNAME=${USERNAME// /_}
DIRECTORY_USERNAME=${DIRECTORY_USERNAME//\//_}
OUTPUT_DIR=$OUTPUT_ROOT/${UUID}_${DIRECTORY_USERNAME}
WORK_DIR=$(mktemp -d)
trap 'rm -rf -- "$WORK_DIR"' EXIT
if [[ ! -d $OUTPUT_ROOT ]]; then
  install -d -m 0755 "$OUTPUT_ROOT"
  if [[ -n ${SUDO_USER:-} && $SUDO_USER != root ]]; then
    chown "$SUDO_USER" "$OUTPUT_ROOT"
  fi
fi
umask 077
mkdir -p "$OUTPUT_DIR"/{linux,windows,ios,android}

# Экранируем специальные символы формата OpenSSL -subj.
SUBJECT_NAME=${USERNAME//\\/\\\\}
SUBJECT_NAME=${SUBJECT_NAME//\//\\/}
SUBJECT_NAME=${SUBJECT_NAME//+/\\+}

cat > "$WORK_DIR/client.ext" <<EOF
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
subjectAltName=URI:urn:myhealth:user:$UUID
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid,issuer
EOF

openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$WORK_DIR/client.key"
openssl req -new -sha256 -utf8 -key "$WORK_DIR/client.key" \
  -out "$WORK_DIR/client.csr" -subj "/CN=$SUBJECT_NAME/UID=$UUID"
openssl x509 -req -sha256 -days 397 -in "$WORK_DIR/client.csr" \
  -CA "$PKI_DIR/ca.crt" -CAkey "$PKI_DIR/ca.key" -CAcreateserial \
  -extfile "$WORK_DIR/client.ext" -out "$WORK_DIR/client.crt"

P12_PASSWORD=$(openssl rand -base64 24 | tr -d '\n')
openssl pkcs12 -export \
  -inkey "$WORK_DIR/client.key" -in "$WORK_DIR/client.crt" -certfile "$PKI_DIR/ca.crt" \
  -name "MyHealth — $USERNAME" -passout "pass:$P12_PASSWORD" -out "$WORK_DIR/client.p12"
openssl x509 -in "$PKI_DIR/ca.crt" -outform DER -out "$WORK_DIR/root-ca.cer"

install -m 0600 "$WORK_DIR/client.key" "$OUTPUT_DIR/linux/client.key"
install -m 0644 "$WORK_DIR/client.crt" "$OUTPUT_DIR/linux/client.crt"
install -m 0600 "$WORK_DIR/client.p12" "$OUTPUT_DIR/linux/client.p12"
install -m 0644 "$PKI_DIR/ca.crt" "$OUTPUT_DIR/linux/root-ca.crt"

install -m 0600 "$WORK_DIR/client.p12" "$OUTPUT_DIR/windows/myhealth-client.pfx"
install -m 0644 "$WORK_DIR/root-ca.cer" "$OUTPUT_DIR/windows/myhealth-root-ca.cer"

install -m 0600 "$WORK_DIR/client.p12" "$OUTPUT_DIR/ios/myhealth-client.p12"
install -m 0644 "$WORK_DIR/root-ca.cer" "$OUTPUT_DIR/ios/myhealth-root-ca.cer"

install -m 0600 "$WORK_DIR/client.p12" "$OUTPUT_DIR/android/myhealth-client.p12"
install -m 0644 "$PKI_DIR/ca.crt" "$OUTPUT_DIR/android/myhealth-root-ca.crt"

for platform in linux windows ios android; do
  printf '%s\n' "$P12_PASSWORD" > "$OUTPUT_DIR/$platform/password.txt"
done

cat > "$OUTPUT_DIR/metadata.txt" <<EOF
name=$USERNAME
uuid=$UUID
issued_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
certificate_sha256=$(openssl x509 -in "$WORK_DIR/client.crt" -noout -fingerprint -sha256 | cut -d= -f2)
EOF

cat > "$OUTPUT_DIR/README.txt" <<'EOF'
Linux:
  Импортируйте linux/client.p12 в браузер. При необходимости добавьте
  linux/root-ca.crt в доверенные центры сертификации.

Windows:
  Импортируйте windows/myhealth-root-ca.cer в «Доверенные корневые центры»,
  затем windows/myhealth-client.pfx в личное хранилище сертификатов.

iOS/iPadOS:
  Передайте файлы из ios на устройство и установите оба профиля. После
  установки CA включите полное доверие в настройках сертификатов.

Android:
  В настройках безопасности установите android/myhealth-root-ca.crt как CA,
  затем android/myhealth-client.p12 как сертификат VPN и приложений.

Пароль контейнера находится в password.txt соответствующего каталога.
Передавайте каталог пользователю по защищённому каналу и удалите лишние копии.
EOF

if [[ -n ${SUDO_USER:-} && $SUDO_USER != root ]]; then
  chown -R "$SUDO_USER" "$OUTPUT_DIR"
fi

echo "Сертификат создан"
echo "Пользователь: $USERNAME"
echo "UUID: $UUID"
echo "Каталог: $OUTPUT_DIR"
