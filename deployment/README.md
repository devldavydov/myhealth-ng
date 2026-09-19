# Развёртывание MyHealth

Production-релиз состоит из статического Go-бинарника API, собранного React
frontend и скриптов установки. PostgreSQL запускается отдельно; Go-сервис при
старте проверяет подключение и автоматически применяет встроенные миграции.

## 1. Сборка релиза

На машине сборки нужны Go 1.26.5, Node.js 20+ и npm:

```bash
./deployment/build-bundles.sh 1.0.0
```

Скрипт запускает backend/frontend-тесты, собирает frontend и статический
Linux-бинарник. Результат — транспортный архив
`artifacts/myhealth-1.0.0.tar.gz` с frontend/backend-бандлами, deploy-скриптами,
версией и SHA-256-манифестом. На сервер переносится только этот архив; Go и
Node.js на сервере не нужны. Deployment-скрипты не копируются в системные
каталоги: запускайте их непосредственно из распакованного архива релиза.
Примеры ниже предполагают, что архив версии `1.0.0` распакован в каталог
`myhealth-release-1.0.0`; этот каталог можно сохранить или создать заново из
архива перед обслуживанием.

По умолчанию бинарник собирается для архитектуры машины сборки. Для другой
архитектуры Linux задайте `MYHEALTH_GOARCH`, например:

```bash
MYHEALTH_GOARCH=arm64 ./deployment/build-bundles.sh 1.0.0
```

Второй аргумент задаёт другой путь выходного архива:

```bash
./deployment/build-bundles.sh 1.0.0 /tmp/myhealth-release.tar.gz
```

## 2. Первая установка

До установки PostgreSQL должен быть доступен серверу по указанной строке
подключения. На новой Ubuntu извлеките deploy-скрипты из архива и запустите
установщик из извлечённого каталога:

```bash
mkdir myhealth-release-1.0.0
tar -xzf myhealth-1.0.0.tar.gz -C myhealth-release-1.0.0
sudo ./myhealth-release-1.0.0/deployment/install-server.sh ./myhealth-1.0.0.tar.gz \
  --public-host 111.88.251.114 \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
```

Архив релиза — обязательный позиционный аргумент. `--public-host` и
`--database-url` также обязательны. Дополнительные параметры:

- `--https-port` — внешний HTTPS-порт, по умолчанию `443`;
- `--server-host` — локальный адрес Go API, по умолчанию `127.0.0.1`;
- `--server-port` — локальный порт Go API, по умолчанию `3000`.

Установщик проверяет архив, устанавливает Nginx и OpenSSL, создаёт системного
пользователя `myhealth`, приватный CA, серверный сертификат и systemd-сервис.
Nginx раздаёт frontend, проксирует `/api/` в Go-сервис и требует клиентский
сертификат.

Файлы размещаются здесь:

- `/opt/myhealth/releases/<version>` — неизменяемые версии приложения;
- `/opt/myhealth/current` — ссылка на активную версию;
- `/etc/myhealth/pki` — CA и серверные ключи;
- `/etc/systemd/system/myhealth.service.d/10-config.conf` — параметры Go API;
- `/var/lib/myhealth` — рабочий каталог системного пользователя;

Systemd drop-in доступен только root, однако database URL виден в аргументах
процесса. Это ограничение текущей схемы конфигурации.

## 3. Клиентские сертификаты

```bash
# Новый пользователь: UUID создаётся автоматически
sudo ./myhealth-release-1.0.0/deployment/generate-client-cert.sh "Иван Иванов"

# Перевыпуск для той же учётной записи с сохранением UUID
sudo ./myhealth-release-1.0.0/deployment/generate-client-cert.sh \
  "Иван Иванов" 3f67c05f-7c9e-4cb5-b26a-f9ce5b065865
```

Третий необязательный аргумент задаёт корневой каталог результата. По умолчанию
комплект создаётся относительно текущего каталога:

```text
client-certificates/3f67c05f-7c9e-4cb5-b26a-f9ce5b065865_Иван_Иванов/
```

Пробелы и `/` в имени заменяются подчёркиваниями. Комплект содержит варианты
для Linux, Windows, iOS/iPadOS и Android, корневой сертификат CA и пароль
PKCS#12. Корневой сертификат нужно добавить в доверенные на устройстве;
пользовательский комплект следует передавать только по защищённому каналу.
Без валидного клиентского сертификата Nginx отвечает HTTP 403.

## 4. Backup и восстановление PostgreSQL

Для запуска скриптов на машине должны быть установлены `pg_dump` и
`pg_restore`; их major-версия должна совпадать с версией PostgreSQL-сервера или
быть новее. Установщик приложения эти утилиты не устанавливает.

Полный логический backup базы приложения создаётся в PostgreSQL custom format.
Он включает схему, данные и large objects, проверяется через `pg_restore` и
получает соседний SHA-256-файл. По умолчанию результат сохраняется с UTC-временем
в имени:

```bash
sudo ./myhealth-release-1.0.0/deployment/backup-database.sh \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
# /var/backups/myhealth/myhealth-YYYYmmddTHHMMSSZ.dump
```

Можно указать другой путь:

```bash
sudo ./myhealth-release-1.0.0/deployment/backup-database.sh \
  /secure/backups/myhealth-before-upgrade.dump \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
```

Строка подключения всегда передаётся явно через `--database-url`. Backup
получает согласованный снимок без остановки API.
Роли и tablespaces всего PostgreSQL-кластера в backup одной базы не входят.

Восстановление проверяет SHA-256 и структуру архива, останавливает активный
`myhealth.service`, заменяет сохранённые объекты в одной транзакции и затем
возвращает сервис в исходное состояние:

```bash
sudo ./myhealth-release-1.0.0/deployment/restore-database.sh \
  /var/backups/myhealth/myhealth-YYYYmmddTHHMMSSZ.dump \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
```

Операция требует ввести `RESTORE`. Для автоматизированного запуска доступен
`--yes`. Файлы в `/var/backups/myhealth` имеют права только для root и при
удалении MyHealth сохраняются.

## 5. Обновление

Сначала извлеките deployment-скрипты именно из нового архива, затем запустите
находящийся в нём upgrade-скрипт:

```bash
mkdir myhealth-upgrade-1.1.0
tar -xzf myhealth-1.1.0.tar.gz -C myhealth-upgrade-1.1.0
sudo ./myhealth-upgrade-1.1.0/deployment/upgrade-server.sh \
  ./myhealth-1.1.0.tar.gz
```

Архив и контрольные суммы проверяются до установки. Релиз размещается в новом
каталоге, ссылка `current` переключается атомарно, а при ошибке запуска
восстанавливаются предыдущий релиз, systemd unit и конфигурация. PKI,
PostgreSQL и существующий systemd drop-in сохраняются.

Чтобы заменить строку подключения или параметры Go API, передайте новую строку
подключения и нужные адрес/порт вместе:

```bash
sudo ./myhealth-upgrade-1.1.0/deployment/upgrade-server.sh \
  ./myhealth-1.1.0.tar.gz \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth' \
  --server-host 127.0.0.1 \
  --server-port 3000
```

Если systemd drop-in отсутствует, `--database-url` обязателен. После
обновления старый распакованный каталог можно удалить, а новый сохранить
для обслуживания этой версии либо при необходимости снова получить из архива.
Deployment-скрипты в систему не устанавливаются.

## 6. Удаление

```bash
sudo ./myhealth-release-1.1.0/deployment/uninstall-server.sh
```

Используйте `uninstall-server.sh` из распакованного архива установленной
версии. Скрипт требует ввести `DELETE`; `--yes` отключает подтверждение. Опция
`--keep-pki` сохраняет `/etc/myhealth/pki`, чтобы ранее выпущенные клиентские
сертификаты продолжили работать после переустановки. PostgreSQL, Nginx и
OpenSSL не удаляются. Пользовательские комплекты сертификатов, созданные вне
системных каталогов MyHealth, также не удаляются.

## Проверка и диагностика

```bash
systemctl status myhealth nginx
journalctl -u myhealth -n 100
nginx -t
curl --cacert root-ca.crt --cert client.crt --key client.key \
  https://111.88.251.114/api/me
```
