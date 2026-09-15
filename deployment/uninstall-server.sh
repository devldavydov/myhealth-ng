#!/usr/bin/env bash
set -Eeuo pipefail

KEEP_PKI=false
ASSUME_YES=false

for argument in "$@"; do
  case "$argument" in
    --keep-pki) KEEP_PKI=true ;;
    --yes|-y) ASSUME_YES=true ;;
    --help|-h)
      echo "Использование: sudo $0 [--keep-pki] [--yes]"
      echo "  --keep-pki  сохранить CA и серверные ключи в /etc/myhealth/pki"
      echo "  --yes       не запрашивать интерактивное подтверждение"
      exit 0
      ;;
    *) echo "Неизвестный аргумент: $argument" >&2; exit 1 ;;
  esac
done

if (( EUID != 0 )); then
  echo "Скрипт необходимо запускать через sudo" >&2
  exit 1
fi

if [[ $ASSUME_YES != true ]]; then
  echo "Будут удалены приложение MyHealth, все релизы, systemd-сервис и конфигурация Nginx."
  if [[ $KEEP_PKI == true ]]; then
    echo "PKI будет сохранена в /etc/myhealth/pki."
  else
    echo "CA и серверные ключи также будут безвозвратно удалены."
  fi
  read -r -p "Введите DELETE для продолжения: " confirmation
  [[ $confirmation == DELETE ]] || { echo "Удаление отменено"; exit 1; }
fi

systemctl disable --now myhealth.service 2>/dev/null || true
rm -f /etc/systemd/system/myhealth.service
rm -rf -- /etc/systemd/system/myhealth.service.d
systemctl daemon-reload

rm -f /etc/nginx/sites-enabled/myhealth /etc/nginx/sites-available/myhealth
if command -v nginx >/dev/null 2>&1 && nginx -t; then
  systemctl reload nginx.service 2>/dev/null || true
fi

rm -rf -- /opt/myhealth /var/lib/myhealth
if [[ $KEEP_PKI == true ]]; then
  rm -f /etc/myhealth/myhealth.env
else
  rm -rf -- /etc/myhealth
fi

if id myhealth >/dev/null 2>&1; then
  userdel myhealth
fi

rm -f /usr/local/sbin/myhealth-upgrade \
  /usr/local/sbin/myhealth-generate-client-cert \
  /usr/local/sbin/myhealth-uninstall
rm -rf -- /usr/local/lib/myhealth-deployment

echo "MyHealth удалён. Пакеты Nginx и OpenSSL оставлены, поскольку они могут использоваться другими приложениями."
if [[ $KEEP_PKI == true ]]; then
  echo "PKI сохранена в /etc/myhealth/pki."
else
  echo "PKI удалена. Ранее выпущенные клиентские сертификаты больше нельзя использовать с новой установкой."
fi
