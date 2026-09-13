# Развёртывание MyHealth

## 1. Сборка

На машине разработчика с Node.js 20+:

```bash
./deployment/build-bundles.sh 1.0.0
```

Результат — один транспортный архив `artifacts/myhealth-1.0.0.tar.gz`. Внутри
него находятся отдельные frontend/backend-бандлы, deploy-скрипты, версия и
SHA-256 манифест. На сервер нужно перенести только этот файл.

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
sudo ./myhealth-installer/deployment/install-server.sh ./myhealth-1.0.0.tar.gz health.example.com
```

С нестандартным HTTPS-портом передайте его третьим аргументом. Скрипт ставит
Nginx и Node.js 20+, создаёт systemd-сервис, локальный CA, серверный сертификат
и устанавливает deploy-скрипты в стандартный каталог:

- `/opt/myhealth/releases/<version>` — неизменяемые версии;
- `/opt/myhealth/current` — ссылка на активную версию;
- `/etc/myhealth` — окружение и PKI;
- `/var/lib/myhealth` — состояние приложения;
- `/usr/local/lib/myhealth-deployment` — deploy-инструменты;

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
релиз. PKI и конфигурация сохраняются; состояние in-memory сбрасывается.

## Восстановление установки с отсутствующим server.cjs

Соберите архив с новым номером версии и перенесите его на сервер. Чтобы не
использовать старый установленный upgrade-скрипт, извлеките свежий прямо из
нового архива:

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
Nginx, Node.js и OpenSSL не удаляются, поскольку могут использоваться другими
приложениями. Созданные ранее пользовательские пакеты вне стандартных путей
MyHealth также не удаляются.

## Проверка

```bash
systemctl status myhealth nginx
journalctl -u myhealth -n 100
nginx -t
```
