# Статус работ

Дата фиксации: 13 сентября 2026 года.

## Реализовано

- React + Vite SPA со страницами обзора и измерений.
- Express JSON API для проверки состояния и работы с измерениями.
- Получение имени и GUID пользователя из клиентского сертификата.
- Отображение имени пользователя в шапке SPA.
- In-memory хранение пользователей и измерений.
- Nginx с обязательной взаимной TLS-аутентификацией (mTLS).
- UTF-8 для JSON-ответа Nginx при отсутствии клиентского сертификата.
- Генерация клиентских сертификатов для Linux, Windows, iOS и Android.
- Перевыпуск сертификата с сохранением существующего GUID.
- Сборка frontend и backend в единый транспортный архив.
- Первоначальная установка, обновление и удаление приложения на Ubuntu.
- Проверка SHA-256, версии и структуры бандлов перед установкой.
- Откат на предыдущий релиз при неудачном обновлении.

## Формат релиза

```bash
./deployment/build-bundles.sh 1.0.3
```

Результат:

```text
artifacts/myhealth-1.0.3.tar.gz
```

Архив содержит отдельные frontend/backend-бандлы, deploy-скрипты,
`release-version` и `SHA256SUMS`.

## Стандартные пути на сервере

- `/opt/myhealth/releases/<version>` — установленные релизы;
- `/opt/myhealth/current` — активный релиз;
- `/etc/myhealth` — конфигурация и PKI;
- `/var/lib/myhealth` — рабочий каталог сервиса;
- `/usr/local/lib/myhealth-deployment` — deploy-скрипты.

## Команды

Обновление:

```bash
sudo /usr/local/lib/myhealth-deployment/upgrade-server.sh \
  /path/to/myhealth-1.0.3.tar.gz
```

Новый пользовательский сертификат:

```bash
sudo /usr/local/lib/myhealth-deployment/generate-client-cert.sh "Иван Иванов"
```

Перевыпуск с прежним GUID:

```bash
sudo /usr/local/lib/myhealth-deployment/generate-client-cert.sh \
  "Иван Иванов" 3f67c05f-7c9e-4cb5-b26a-f9ce5b065865
```

Удаление приложения:

```bash
sudo /usr/local/lib/myhealth-deployment/uninstall-server.sh
```

## Исправленные проблемы

- Backend собирается в CommonJS-файл `server.cjs`; ESM-бандлы с
  `server.mjs` отклоняются до установки.
- Каталоги релиза получают права `0755`, файлы — `0644`. Это позволяет
  systemd-сервису, работающему от пользователя `myhealth`, читать backend и
  позволяет Nginx читать frontend.
- Ошибка mTLS 403 возвращается как `application/json` в UTF-8.
- Symlink-команды в `/usr/local/sbin` не создаются; используются обычные
  deploy-скрипты.

## Проверки

- Все 7 автоматических тестов проходят.
- Production-сборка клиента и сервера проходит.
- Единый архив успешно распаковывается и проходит проверку SHA-256.
- Backend-бандл проходит `node --check` и smoke-тест API.
- Shell-скрипты проходят `bash -n`.
- Production-зависимости не имеют известных уязвимостей по `npm audit --omit=dev`.

Текущая установка версии `1.0.2` после исправления прав успешно работает.
