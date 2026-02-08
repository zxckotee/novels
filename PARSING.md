# Парсинг/импорт новелл (архитектура)

## Идея: голосование — родитель, импортеры/парсеры — дочерние

**Бизнес‑логика голосований** (Daily Vote и Translation Tickets) живёт отдельно от парсеров.
Когда выбирается победитель, доменная логика публикует событие, а “дочерний” слой (импорт/парсеры)
подписан на эти события и выполняет импорт.

## Поток Daily Vote (добавление новой книги)

1. Пользователи голосуют `daily_vote` через `POST /api/v1/votes`.
2. Джоба выбирает победителя (каждые 6 часов): `VotingService.ProcessVotingWinner()`.
3. Победитель переводится в статус `accepted` (то есть “выбран на выпуск/импорт”) и публикуется событие:
   `daily_vote_winner_selected`.
4. `ImportOrchestrator` ловит событие и выбирает нужного импортера по `proposal.original_link`,
   запускает импорт в БД и после успеха публикует событие `proposal_released`.

## Поток Translation Tickets (перевод существующей/анонсированной книги)

1. Пользователи вкладывают билеты перевода через `POST /api/v1/translation-votes`
   (это отдельный лидерборд от daily vote).
2. Джоба выбирает победителя (каждые 6 часов): `TranslationVotingService.ProcessTranslationWinner()`.
3. Если победитель уже существует как `novel_id` → статус `translating`.
4. Если победитель — анонс (target с `proposal_id`) → статус `waiting_release`.
5. Когда daily vote импорт закончен и срабатывает `proposal_released`, target для перевода
   автоматически биндится к `novel_id` и `waiting_release -> translating`.

Важно: **голоса/билеты не сбрасываются у других** при выборе победителя.

## Как добавить новый сайт

1. Реализовать новый импортер, который умеет:
   - определить, что он “берёт” ссылку (по домену/паттерну),
   - импортировать новеллу/главы в БД.
2. Зарегистрировать импортер в `ImportOrchestrator` (список импортёров).

Сейчас пример реализован для `fanqienovel.com`.

## 69shuba (через интерактивный Playwright-браузер в Docker)

69shuba защищён Cloudflare. Мы не “обходим” защиту программно — вместо этого используем
**интерактивный браузер** (Playwright) в Docker: вы проходите проверки вручную, после чего
мы сохраняем `storage_state` (cookies + localStorage) в файл для последующего использования.

### Запуск браузера (noVNC)

Поднимите сервис (он также стартует вместе с остальными при обычном `docker compose up -d`):

```bash
sudo docker compose up -d --build shuba-browser
```

Откройте noVNC в браузере: `http://localhost:7900` (внутри будет окно Chromium).

### Сохранение cookies/storage state

Запустите интерактивную сессию (внутри контейнера):

```bash
sudo docker exec -it novels-shuba-browser \
  python /app/shuba_browser_session.py \
  --url "https://www.69shuba.com/book/90474.htm" \
  --headed \
  --out "/data/69shuba_storage.json"
```

Дальше:
- пройдите проверки/войдите в браузере (noVNC),
- вернитесь в терминал и нажмите ENTER — файл сохранится в `./cookies/69shuba_storage.json`.

### Импорт 69shuba в БД (через docker exec)

После того как `./cookies/69shuba_storage.json` сохранён:

```bash
sudo docker exec novels-backend /app/import_69shuba \
  --url "https://www.69shuba.com/book/90474.htm" \
  --chapters-limit 10 \
  --storage-state "/app/cookies/69shuba_storage.json"
```

Примечание: если путь неудобный, можно вместо `--storage-state` передать `--cookie` (Cookie header) и `--user-agent`.

## parser-service (Python + Playwright) — общий способ для сайтов с блокировками Go-клиента

Мы оставляем доменную логику (voting/orchestrator/DB-write) в Go, но переносим “скачивание и парсинг страниц”
в отдельный сервис `parser-service` на Python + Playwright (реальный Chromium).  
Go-импортёры `import_69shuba` и `import_101kks` теперь **не делают прямых HTTP запросов к сайтам**, а вызывают `parser-service`,
который возвращает распарсенный JSON.

### Запуск

```bash
cd /home/fsociety/novels
sudo docker compose up -d --build parser-service
```

Порт наружу: `http://localhost:8010`  
Healthcheck: `GET /health`

### storage_state (cookies) — где лежит

`parser-service` монтирует `./cookies` как `/data`.  
По умолчанию он будет использовать:
- `/data/101kks_storage.json`
- `/data/69shuba_storage.json`

Если надо явно — передаём `--storage-state`, но **в Go он будет преобразован в `/data/<basename>`** внутри parser-service.

### Импорт 101kks через parser-service

```bash
sudo docker exec novels-backend /app/import_101kks \
  --url "https://101kks.com/book/12544.html" \
  --chapters-limit 10 \
  --user-agent "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36" \
  --referer "https://101kks.com/booklist/detail/8.html"
```

### Импорт 69shuba через parser-service

```bash
sudo docker exec novels-backend /app/import_69shuba \
  --url "https://www.69shuba.com/book/90474.htm" \
  --chapters-limit 10 \
  --user-agent "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36"
```

## 101kks (по `parser_101.md`)

Сайт `101kks.com` может требовать Cloudflare cookie (`cf_clearance`) для запросов из Docker.
Мы не “обходим” защиту программно — используем тот же интерактивный браузер `shuba-browser` для
ручного прохождения/получения cookies и сохраняем `storage_state`.

### Сохранение storage state

Открой noVNC: `http://localhost:7900`

```bash
sudo docker exec -it novels-shuba-browser \
  python /app/shuba_browser_session.py \
  --url "https://101kks.com/book/12544.html" \
  --headed \
  --out "/data/101kks_storage.json"
```

Файл появится в `./cookies/101kks_storage.json`.

### Импорт 101kks в БД

```bash
sudo docker exec novels-backend /app/import_101kks \
  --url "https://101kks.com/book/12544.html" \
  --chapters-limit 10 \
  --storage-state "/app/cookies/101kks_storage.json" \
  --user-agent "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 YaBrowser/25.6.0.0 Safari/537.36"
```

Если нужно как в `parser_101.md`, можно передать `--referer "https://101kks.com/booklist/detail/8.html"`.

## Тесты через docker exec

После сборки контейнеров:

```bash
sudo docker compose up -d --build
```

Запуск импорта напрямую (пишет в БД: `novels`, `novel_localizations`, `chapters`, `chapter_contents`):

```bash
sudo docker exec novels-backend /app/import_fanqie \
  --url "https://fanqienovel.com/page/7276384138653862966" \
  --chapters-limit 10
```

### Placeholder "parser" для пропущенных полей

Для тестовых импортов поля, которые обычно приходят из “предложки”, но не извлекаются парсером,
заполняются значением `"parser"`:

- `novels.author = "parser"`
- `novel_localizations.alt_titles = ["parser"]`
- при отсутствии описания: `novel_localizations.description = "parser"`
- дополнительно создаются и привязываются `genre/tag` со slug `"parser"` (чтобы метаданные не были пустыми)

## Pause/Resume для импортов (парсинг -> запись в БД)

Импорты победителей (daily vote) запускаются асинхронно через `ImportOrchestrator` и пишутся в таблицу `import_runs`.
Теперь импорт можно **поставить на паузу** и **продолжить с места остановки**:

- Импорт идёт **по главам**, каждая глава коммитится отдельно
- Прогресс хранится в `import_runs.progress_current / progress_total`
- Чекпоинт хранится в `import_runs.checkpoint` (novelId/slug/nextIndex)

### Эндпоинты (admin-only)

- **Список запусков**:

```bash
curl -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs?limit=50"
```

- **Пауза** (для активного run):```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/pause"
```

- **Продолжить**:

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/resume"
```

- **Отмена**:

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/cancel"
```