# Job control (pause/resume) — parsing/import/translation

## Импорт (парсинг -> запись в БД)

Импорт победителя (daily vote) запускается асинхронно через `ImportOrchestrator` и пишется в таблицу `import_runs`.

Поддерживается:

- **Pause**: останавливает текущий run, сохраняя прогресс
- **Resume**: продолжает с сохранённого `checkpoint`
- **Cancel**: отмена run

Технически:

- импорт идёт **по главам**, каждая глава коммитится отдельно
- прогресс: `import_runs.progress_current` / `import_runs.progress_total`
- чекпоинт (JSONB): `import_runs.checkpoint` (включает `novelId`, `slug`, `nextIndex`)

## Admin API

Все эндпоинты ниже находятся под `admin` роутами и требуют `Authorization: Bearer <ADMIN_JWT>`.

### List runs

```bash
curl -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs?limit=50"
```

### Pause

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/pause"
```

### Resume

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/resume"
```

### Cancel

```bash
curl -X POST -H "Authorization: Bearer <ADMIN_JWT>" \
  "http://localhost:8080/api/v1/admin/ops/import-runs/<RUN_ID>/cancel"
```

