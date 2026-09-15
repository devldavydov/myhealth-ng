# Развёртывание MyHealth

## 1. Сборка

На машине разработчика с Go 1.26.5 и Node.js 20+:

```bash
./deployment/build-bundles.sh 1.0.0
```

Результат — один транспортный архив `artifacts/myhealth-1.0.0.tar.gz`. Внутри
него находятся отдельные frontend/backend-бандлы, deploy-скрипты, версия и
SHA-256 манифест. На сервер нужно перенести только этот файл.

Backend компилируется в статический Linux-бинарник, поэтому Go и Node.js на
сервере не нужны. PostgreSQL разворачивается отдельно и должен быть доступен до
установки приложения. По умолчанию используется архитектура машины сборки. Для
кросс-компиляции, например под ARM64, задайте `MYHEALTH_GOARCH=arm64`:

```bash
MYHEALTH_GOARCH=arm64 ./deployment/build-bundles.sh 1.0.0
```

Вторым аргументом можно задать другое имя выходного архива:

```bash
./deployment/build-bundles.sh 1.0.0 /tmp/myhealth-release.tar.gz
```

## 2. Первая установка

На новой Ubuntu сначала извлеките из архива deploy-скрипты, затем передайте
сам архив установщику:

```bash
mkdir myhealth-installer
tar -xzf myhealth-1.0.0.tar.gz -C myhealth-installer
sudo ./myhealth-installer/deployment/install-server.sh ./myhealth-1.0.0.tar.gz \
  --public-host health.example.com \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
```

Необязательные параметры --https-port, --server-host и --server-port меняют
соответственно внешний HTTPS-порт и адрес backend. Скрипт ставит Nginx и
OpenSSL, создаёт systemd-сервис, локальный CA, серверный сертификат и
устанавливает deploy-скрипты в стандартный каталог:

- `/opt/myhealth/releases/<version>` — неизменяемые версии;
- `/opt/myhealth/current` — ссылка на активную версию;
- `/etc/myhealth` — окружение и PKI;
- `/var/lib/myhealth` — состояние приложения;
- `/usr/local/lib/myhealth-deployment` — deploy-инструменты;

Параметры запуска backend записываются в
`/etc/systemd/system/myhealth.service.d/10-config.conf`. Файл доступен только
root, однако строка подключения также видна в аргументах процесса.

Серверный сертификат подписан приватным MyHealth CA. Поэтому корневой
сертификат из пользовательского комплекта необходимо сделать доверенным на
устройстве. Без клиентского сертификата Nginx отвечает HTTP 403.

## 3. Пользовательский сертификат

```bash
# Новый пользователь: GUID будет создан автоматически
sudo /usr/local/lib/myhealth-deployment/generate-client-cert.sh "Иван Иванов"

# Перевыпуск сертификата с сохранением прежнего GUID
sudo /usr/local/lib/myhealth-deployment/generate-client-cert.sh "Иван Иванов" 3f67c05f-7c9e-4cb5-b26a-f9ce5b065865
```

Без второго аргумента создаются новая ключевая пара и новый UUID. Если передать
существующий GUID вторым аргументом, сертификат будет перевыпущен для той же
учётной записи. Необязательный третий аргумент задаёт каталог результата.
Комплект содержит форматы для Linux, Windows, iOS/iPadOS и Android. Пароль
PKCS#12 находится рядом; передавайте комплект только по защищённому каналу.

## 4. Обновление

```bash
sudo /usr/local/lib/myhealth-deployment/upgrade-server.sh /path/to/myhealth-1.1.0.tar.gz
```

Архив распаковывается во временный каталог, а внутренние контрольные суммы
проверяются до установки. Новая версия размещается отдельно, затем атомарно
переключается `current`. При ошибке запуска сервис возвращается на предыдущий
релиз. PKI, PostgreSQL и systemd drop-in с конфигурацией сохраняются.

При первом переходе с версии, использовавшей environment-конфигурацию,
необходимо передать строку подключения:

```bash
sudo /usr/local/lib/myhealth-deployment/upgrade-server.sh /path/to/myhealth-1.1.0.tar.gz \
  --database-url 'postgresql://myhealth:password@db.example.com/myhealth'
```

Те же параметры можно использовать позднее, чтобы заменить database URL,
backend host или backend port.

## Переход с Node.js-бэкенда на Go

Соберите архив с новым номером версии и перенесите его на сервер. Старый
установленный upgrade-скрипт ожидает `server.cjs`, поэтому при первом переходе
извлеките новый скрипт прямо из свежего архива:

```bash
mkdir myhealth-upgrade-1.0.1
tar -xzf myhealth-1.0.1.tar.gz -C myhealth-upgrade-1.0.1
sudo ./myhealth-upgrade-1.0.1/deployment/upgrade-server.sh ./myhealth-1.0.1.tar.gz
```

Существующий релиз удалять не нужно. После успешного обновления
обычные скрипты в `/usr/local/lib/myhealth-deployment` также обновятся.

## 5. Удаление

```bash
sudo /usr/local/lib/myhealth-deployment/uninstall-server.sh
```

Скрипт требует ввести `DELETE`. Для автоматического запуска используйте
`--yes`. Опция `--keep-pki` сохраняет CA и ключи для последующей переустановки.
Nginx и OpenSSL не удаляются, поскольку могут использоваться другими
приложениями. Созданные ранее пользовательские пакеты вне стандартных путей
MyHealth также не удаляются.

## Проверка

```bash
systemctl status myhealth nginx
journalctl -u myhealth -n 100
nginx -t
```
