# ARCHITECTURE — Novels

Документ фиксирует целевую архитектуру платформы **Novels**.

Коротко: это **платформа для чтения новелл** (текстовый ридер в стиле **ranobehub**), но визуально/UX копируем/адаптируем дизайн-паттерны **Remanga** (карточки, хедер, топы, коллекции и т.п.). Монетизация: **только подписка** (без покупки глав/пакетов). SEO — важно, поэтому фронтенд проектируем под **SSR/пререндер**.

---

## 0) Цели и ограничения

### Цели
- Быстрый запуск MVP: каталог → тайтл → глава → чтение → прогресс.
- Масштабируемая доменная модель под экономику (тикеты/голосование/подписка).
- Многоязычный UI (7 языков) + локализация контента (названия/описания).
- SEO: индексируемые страницы тайтлов/глав + sitemap.

### Не-цели (на первом этапе)
- Микросервисы. Стартуем монолитом (modular monolith).
- Сложный антифрод “с первого дня” (но базовые лимиты/защита обязательны).
- Авто-перевод: **пока заглушка в админке**.

---

## 1) Технологический стек (базовый)

- **Backend**: Go + `chi` (REST API)
- **DB**: PostgreSQL
- **Frontend**: React, но с SSR/пререндером (рекомендуемо: **Next.js** как “React + SSR”)

Примечание: “React” из README сохраняется; Next.js — это способ реализовать SSR/SEO без смены парадигмы.

---

## 2) Высокоуровневая схема системы

### Компоненты
- **Web App (SSR React)**:
  - SSR/SSG для SEO-страниц (главная, каталог, тайтл, глава)
  - Клиентские интеракции (комменты, закладки, голосования)
- **API (Go/chi)**:
  - Публичные чтение/поиск
  - Пользовательские операции (закладки/прогресс/комменты)
  - Админские операции (контент, модерация, новости)
  - Экономика (тикеты, голосования, подписка)
- **PostgreSQL**: основное хранилище
- **Background Worker (в рамках бэка)**:
  - ежедневные начисления Daily Vote (00:00 UTC)
  - автоснятие победителя голосования раз в X часов
  - (позже) задачи уведомлений/индексации/генерации sitemap
- **Object Storage (опционально)**:
  - обложки/медиа (на старте можно локально, затем S3-compatible)

### Потоки
- Пользователь читает через SSR страницу главы → API для прогресса/комментов.
- Поиск/каталог: Postgres FTS + фильтры + кеш (на популярные выдачи).

---

## 3) Домен и термины

- **Тайтл (Novel)**: произведение (карточка, метаданные, статус перевода).
- **Глава (Chapter)**: текстовый контент; ридер “как ranobehub”.
- **Daily Vote**: ежедневный билет (сгорает, не накапливается), начисление всем пользователям 1 раз/сутки.
- **Novel Request**: билет на добавление тайтла в предложку.
- **Translation Ticket**: билет на заказ перевода (в MVP может быть no-op/заглушка).
- **Подписка (Premium)**: ежемесячная, даёт привилегии из README:
  - отключение рекламы
  - ежемесячный пакет Translation Ticket + Novel Request
  - увеличенный множитель Daily Vote (например, x2)
  - доступ к редактированию описания (через модерацию)
  - доступ к “повторному переводу глав” (запрос на re-translate)

---

## 4) Роли и права (RBAC)

Роли:
- **Guest**: чтение, каталог, тайтл/глава (в рамках доступных глав)
- **User**: всё Guest + комментарии/закладки/голосование/XP
- **Premium**: всё User + редактирование описания (через модерацию) + бонусы подписки
- **Moderator**: модерация (комментарии, правки описаний, новости)
- **Admin**: полный доступ (контент, пользователи, экономика)

Мини-матрица прав (пример):
- **Контент**:
  - Guest/User/Premium: read
  - Moderator: read + approve wiki edits (опционально)
  - Admin: CRUD
- **Комментарии**:
  - User/Premium: create/edit-own/delete-own
  - Moderator/Admin: delete/ban/moderate
- **Вики-правки**:
  - Premium: create edit-request
  - Moderator/Admin: approve/reject
- **Новости**:
  - Moderator/Admin: create/update/delete
  - all: read + comment

---

## 5) Архитектура backend (Go/chi)

### 5.1 Структура модулей (рекомендация)
- `cmd/api` — запуск HTTP API
- `internal/http` — роуты, middleware, handlers
- `internal/domain` — доменные типы/интерфейсы
- `internal/service` — бизнес-логика (use-cases)
- `internal/repo` — доступ к БД (SQL)
- `internal/auth` — токены, сессии, пароли
- `internal/jobs` — cron/worker задачи
- `migrations/` — миграции БД

### 5.2 API стиль
- REST JSON.
- Единые ошибки: `{ "error": { "code": "...", "message": "...", "details": ... } }`
- Идемпотентность:
  - операции начислений/списаний — через транзакции и уникальные ключи (см. `ticket_transactions`).

### 5.3 Middleware (обязательные)
- Request ID (корреляция логов)
- CORS (для SSR может быть минимальный)
- Rate limiting (IP + user)
- Auth (jwt/cookie)
- RBAC guard (role/permission)
- Audit logging для админ/экономики

---

## 6) Архитектура frontend (React + SSR)

### 6.1 Почему SSR
- SEO на страницах тайтлов/глав/каталога.
- Быстрый TTFB на популярных страницах.
- Возможность генерировать sitemap и отдавать корректные meta/OG.

Рекомендация: **Next.js**:
- `app/` router или `pages/` (на выбор)
- SSR/SSG + ISR для каталога и страниц тайтлов/глав
- API остаётся в Go; Next.js используется только как Web (не “backend логика”)

### 6.2 i18n (7 языков)
- UI-тексты: словари по `lang` (например `ru`, `en`, …).
- Язык — в URL (рекомендуемо): `/{lang}/...` для SEO и shareable ссылок.
- Fallback: `ru` (или другой дефолт) при отсутствии перевода.

### 6.3 Критические страницы
- Главная: топы/новости/коллекции/последние обновления
- Каталог + поиск: фильтры, сортировки, пагинация
- Тайтл: обложка, метаданные, табы (описание/оглавление/комменты)
- Ридер главы (ranobehub-style): текст + навигация + прогресс
- Профиль: статистика/закладки/комменты/ачивки

---

## 7) Данные: PostgreSQL (предлагаемая схема)

Ниже — “максимально практичный” каркас, без попытки сделать финальную DDL 1-в-1.

### 7.1 Базовые сущности контента
- `novels`:
  - `id` (uuid/bigint), `slug` (unique), `cover_image_key`
  - `translation_status` (enum: ongoing/completed/paused/dropped)
  - `original_chapters_count` (int), `release_year` (int), `created_at`, `updated_at`
- `novel_localizations`:
  - `novel_id`, `lang`, `title`, `description`, `alt_titles`
  - unique(`novel_id`,`lang`)
- `chapters`:
  - `id`, `novel_id`, `number` (numeric/int), `slug` (optional), `title`
  - `published_at`, `created_at`, `updated_at`
- `chapter_contents` (чтобы отделить метаданные от текста):
  - `chapter_id`, `lang`, `content` (text), `source` (enum/manual/stub), `updated_at`
  - unique(`chapter_id`,`lang`)

### 7.2 Пользователи
- `users`: `id`, `email` unique, `password_hash`, `created_at`, `last_login_at`, `is_banned`
- `user_profiles`: `user_id`, `display_name`, `avatar_key`, `bio`
- `user_roles`: `user_id`, `role` (enum)

### 7.3 Прогресс чтения
- `reading_progress`:
  - `user_id`, `novel_id`, `chapter_id`, `updated_at`
  - unique(`user_id`,`novel_id`)

### 7.4 Закладки
- `bookmark_lists` (системные/кастомные списки):
  - `id`, `user_id`, `code` (reading/planned/dropped/read/favorites), `title`, `sort_order`
- `bookmarks`:
  - `user_id`, `novel_id`, `list_id`, `created_at`
  - unique(`user_id`,`novel_id`)

### 7.5 Комментарии (nested)
- `comments`:
  - `id`, `parent_id` (nullable), `root_id` (nullable), `depth`
  - `target_type` (novel/chapter/news/profile), `target_id`
  - `user_id`, `body`, `is_deleted`, `created_at`, `updated_at`
- `comment_votes`: `comment_id`, `user_id`, `value` (-1/+1), unique(`comment_id`,`user_id`)
- `comment_reports`: `id`, `comment_id`, `user_id`, `reason`, `created_at`

### 7.6 Новости
- `news_posts`: `id`, `slug`, `title`, `body`, `author_id`, `published_at`, `created_at`, `updated_at`

### 7.6.1 Рейтинг/отзывы (если включаем)
- `novel_ratings`:
  - `novel_id`, `user_id`, `value` (1..5), `created_at`, `updated_at`
  - unique(`novel_id`,`user_id`)
- `novel_reviews` (опционально отдельный тип “рецензий”):
  - `id`, `novel_id`, `user_id`, `title`, `body`, `created_at`, `updated_at`

### 7.6.2 Просмотры и “тренды”
- `novel_views_daily`:
  - `novel_id`, `date` (UTC), `views` int
  - unique(`novel_id`,`date`)
- `chapter_views_daily` (опционально):
  - `chapter_id`, `date` (UTC), `views` int
  - unique(`chapter_id`,`date`)

### 7.7 Подписка
- `subscription_plans`:
  - `id`, `code`, `title`, `price`, `period` (monthly), `is_active`
- `subscriptions`:
  - `id`, `user_id`, `plan_id`, `status` (active/canceled/past_due), `starts_at`, `ends_at`, `created_at`

### 7.8 Тикеты и транзакции
- `ticket_balances`:
  - `user_id`, `type` (daily_vote/novel_request/translation_ticket), `balance` int, `updated_at`
  - unique(`user_id`,`type`)
- `ticket_transactions`:
  - `id`, `user_id`, `type`, `delta` (int), `reason`, `ref_type`, `ref_id`, `created_at`
  - `idempotency_key` unique (критично для кронов/повторов)

### 7.9 Голосование и предложка
- `novel_proposals`:
  - `id`, `user_id`, `original_link`, `status` (draft/moderation/voting/rejected/accepted)
  - черновые метаданные (title/description/tags/genres…), `created_at`, `updated_at`
- `voting_polls`:
  - `id`, `status` (active/closed), `starts_at`, `ends_at`, `created_at`
- `voting_entries`:
  - `poll_id`, `novel_id` (или proposal_id), `score` (int), unique(`poll_id`,`novel_id`)
- `votes`:
  - `id`, `poll_id`, `user_id`, `novel_id`, `ticket_type`, `amount`, `created_at`

### 7.10 Уровни, XP, ачивки
- `user_xp`:
  - `user_id`, `xp_total` bigint, `level` int, `updated_at`
- `xp_events`:
  - `id`, `user_id`, `type` (read_chapter/comment/...), `delta`, `ref_type`, `ref_id`, `created_at`
- `achievements`:
  - `id`, `code`, `title`, `description`, `icon_key`
- `user_achievements`:
  - `user_id`, `achievement_id`, `unlocked_at`
  - unique(`user_id`,`achievement_id`)

### 7.11 Коллекции пользователей
- `collections`:
  - `id`, `user_id`, `slug`, `title`, `description`, `created_at`, `updated_at`
- `collection_items`:
  - `collection_id`, `novel_id`, `position`, `added_at`
  - unique(`collection_id`,`novel_id`)
- `collection_votes`:
  - `collection_id`, `user_id`, `value` (+1), `created_at`
  - unique(`collection_id`,`user_id`)

---

## 8) Поиск и каталог

### Полнотекстовый поиск (Postgres FTS)
- Индексировать:
  - `novel_localizations.title`, `novel_localizations.description`
  - автор/теги/жанры (либо через join, либо денормализованный search_doc)
- Учитывать `lang` (поиск в текущем языке UI).

### Фильтры/сортировки
- Жанры (many-to-many)
- Статус перевода
- Сортировки: просмотры, закладки, отзывы (если появятся), дата добавления, дата обновления

Кеш:
- “популярное/тренды/новинки/последние обновления” кешировать на короткое время (например 30–120 сек).

---

## 9) Экономика: правила и инварианты

### 9.1 Daily Vote (сгорают)
- Начисление: 1 раз в сутки всем пользователям в 00:00 UTC.
- **Сгорание**: неиспользованные Daily Vote “обнуляются” при следующем начислении.
  - Практический вариант реализации:
    - хранить баланс в `ticket_balances(user_id,type='daily_vote')`
    - при начислении: `balance = base_amount` (или `base_amount * multiplier`), независимо от прошлых остатков
    - транзакция начисления — идемпотентна по ключу `{date}:{user_id}:daily_vote` (через `ticket_transactions.idempotency_key`)

### 9.2 Множитель Daily Vote у Premium
- При начислении Daily Vote применяем множитель плана подписки (например x2).
- Множитель влияет только на “вес голоса” или на “кол-во билетов” — выбрать один вариант и везде соблюдать (предпочтение: **выдавать больше daily_vote билетов**, проще для UX).

### 9.3 Translation Ticket и Novel Request
- Выдача: по подписке (ежемесячно) + (позже) за уровни.
- Списание: только через транзакции, баланс не может уйти < 0.

---

## 10) Фоновые задачи (cron/worker)

### 10.1 Grant Daily Vote (ежедневно)
Триггер: 00:00 UTC.
- Для каждого активного пользователя:
  - определить multiplier (premium?)
  - `ticket_balances(daily_vote).balance = 1 * multiplier`
  - записать `ticket_transactions` с idempotency key

### 10.2 Pick winner from voting poll (раз в X часов)
- Найти активный poll
- Взять entry с max score
- Перевести тайтл/предложку в статус “В переводе” или отправить задачу переводчикам
- Зафиксировать событие (audit log)

### 10.3 Sitemap/SEO (периодически)
- Генерация sitemap для:
  - тайтлы
  - главы (если допустимо индексировать)
  - новости

---

## 11) Редактирование описаний (wiki через модерацию)

### Поток
1) Premium нажимает “Редактировать описание”
2) Создаётся “edit request” (черновик изменений) со статусом `pending`
3) Moderator/Admin:
   - approve → применяем изменения к `novels`/`novel_localizations` (и фиксируем историю)
   - reject → сохраняем причину

Рекомендованные таблицы:
- `novel_edit_requests` + `novel_edit_request_items` или одна JSON-таблица “diff”
- `novel_edit_history` (кто/когда/что поменял)

---

## 12) Комментарии “как Reddit/ranobehub”

### Nested комментарии (MVP)
Варианты хранения:
- adjacency list (`parent_id`) + `root_id` + `depth` (простая реализация)
- (позже) materialized path / ltree для быстрых выборок больших тредов

### Абзацные комментарии (позже)
Потребуется стабильный “якорь”:
- `anchor` = `{paragraph_index}` (если контент стабилен)
- или `{quote_hash + offset}` (если контент может меняться)

---

## 13) Безопасность

### Auth
- Пароли: `bcrypt` или `argon2id`.
- Сессии:
  - access token (короткий)
  - refresh token (длинный) в httpOnly cookie
- CSRF защита, если используем cookie (double-submit или same-site + CSRF token).

### Anti-abuse (минимум)
- Rate-limit на:
  - login/register
  - create comment
  - vote
- CAPTCHA при подозрении (позже), fingerprint (позже).

### RBAC
- Проверка роли в middleware.
- Админские эндпоинты + аудит.

---

## 14) Наблюдаемость и эксплуатация

- Структурные логи (json) + request_id
- Метрики:
  - latency p50/p95/p99
  - error rate
  - cron job durations + failures
- Трассировка (позже): OpenTelemetry
- Алерты (позже): по ошибкам/крон-фейлам

---

## 15) SEO детали

### URL структура (рекомендация)
- `/{lang}/` — главная
- `/{lang}/catalog`
- `/{lang}/novel/{slug}`
- `/{lang}/novel/{slug}/chapter/{chapterSlugOrNumber}`
- `/{lang}/news/{slug}`

### Meta/OG
- Тайтл/описание из локализации
- OG:image = обложка

### Sitemap
- отдельные sitemap по типам (novels, chapters, news) + index

---

## 16) API (черновые контракты)

Префикс: `/api/v1`

### Public
- `GET /novels` (filters/sort/pagination, lang)
- `GET /novels/{slug}`
- `GET /novels/{slug}/chapters`
- `GET /chapters/{id}` (или по slug)
- `GET /news`
- `GET /news/{slug}`

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/refresh`
- `GET /me`

### User features
- `POST /progress` (novel_id/chapter_id)
- `GET /bookmarks`
- `POST /bookmarks` / `PUT /bookmarks/{novel_id}`
- `GET /comments` (by target)
- `POST /comments`
- `POST /comments/{id}/vote`
- `POST /comments/{id}/report`
- `POST /ratings` (1..5)
- `POST /reviews` (если есть рецензии)

### Economy
- `GET /wallet` (balances)
- `POST /votes` (spend daily_vote/translation_ticket)
- `POST /proposals` (requires novel_request)
- `GET /proposals` (list/status)
- `GET /polls/current`

### Wiki edits
- `POST /novels/{id}/edit-requests` (premium)
- `GET /moderation/edit-requests` (moder/admin)
- `POST /moderation/edit-requests/{id}/approve|reject`

### Admin
- `POST /admin/novels` / `PUT /admin/novels/{id}`
- `POST /admin/chapters` / `PUT /admin/chapters/{id}`
- `POST /admin/news`
- (заглушка) `POST /admin/translate` (no-op/placeholder)

---

## 17) Транзакционность и консистентность

Ключевые места, где **обязательно** транзакция в БД:
- Списание тикетов + запись `votes`/`proposals`.
- Начисления по cron (идемпотентность через `idempotency_key`).
- Модерация вики-правок (apply diff + история).

Правило: баланс тикетов меняется **только** через `ticket_transactions` (или строго согласованную транзакцию), чтобы не ловить гонки.

---

## 18) Производительность (база)

- Пагинация везде.
- Индексы:
  - `novels.slug`
  - `novel_localizations(novel_id, lang)`
  - `chapters(novel_id, number)`
  - `comments(target_type, target_id, created_at)`
  - `ticket_transactions(user_id, created_at)`
  - FTS индекс по search_doc
- Кешировать “топы” и “главную”.

---

## 19) Деплой и окружения (рекомендация)

Окружения:
- `dev` (локально docker compose)
- `staging` (похоже на prod)
- `prod`

Компоненты деплоя:
- `api` (Go)
- `web` (Next.js SSR)
- `db` (Postgres)
- `object_storage` (опционально)

---

## 20) Открытые решения (ADR backlog)

Нужно зафиксировать отдельными ADR (в `docs/adr/`), когда будем кодить:
- Выбор SSR фреймворка (Next.js vs Remix vs SPA+prerender)
- Стратегия хранения текста глав (в БД vs в файловом хранилище)
- Модель комментов (ltree/materialized path) при росте
- Антифрод: captcha/fingerprint и пороги включения

---

## 21) Важные сценарии (sequence “словами”)

### 21.1 Чтение главы + прогресс
- SSR рендерит страницу главы (получает контент главы с API/DB).
- Клиент:
  - по событию “перешёл на главу” отправляет `POST /progress`
  - (позже) может отправлять “position” (процент/параграф) раз в N секунд.

### 21.2 Голосование Daily Vote
- Пользователь нажимает “Голосовать” → `POST /votes`
- Backend:
  - транзакция: проверить `ticket_balances(daily_vote)` → списать → записать `votes` → обновить `voting_entries.score`
  - вернуть новый баланс и новый score.

### 21.3 Premium редактирует описание (модерация)
- Premium отправляет `POST /novels/{id}/edit-requests`
- Moderator approves:
  - транзакция: применить diff → записать историю → закрыть request.

---

## 22) Конфигурация и секреты (env vars)

Минимальный список (рекомендация):
- `APP_ENV` = dev/staging/prod
- `API_BASE_URL`
- `DB_DSN` (Postgres)
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`
- `COOKIE_DOMAIN`, `COOKIE_SECURE`, `COOKIE_SAMESITE`
- `CORS_ALLOWED_ORIGINS`
- `RATE_LIMIT_*` (пороги)
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` (если включено)
- `PAYMENT_PROVIDER_*` (заглушка/позже)

Принцип: секреты не коммитим; в prod — через secret manager.

---

## 23) Миграции и бэкапы

- Миграции: строго версионированные, применяются автоматически при деплое (или отдельным job).
- Бэкапы Postgres:
  - ежедневный full + (опционально) WAL
  - регулярная проверка восстановления.

---

## 24) Биллинг (как заглушка на старте)

Так как “только подписка”, интеграция биллинга должна быть изолирована:
- `billing` модуль с интерфейсом (provider-agnostic)
- `payment_webhooks` эндпоинт для подтверждения оплаты/продления
- на MVP можно:
  - руками активировать Premium через админку (feature flag)
  - хранить `subscriptions` как “источник истины”.

---

## 25) Реклама (только выключение по подписке)

Архитектурно:
- В web: “ad slots” рендерятся только если `!isPremium`.
- В API: `me` включает признак premium; SSR может использовать его для server-side ветвления.

---

## 26) CI/CD (рекомендация)

- PR checks:
  - backend: `go test`, линтер
  - frontend: typecheck, lint, build
- Deploy:
  - staging на merge в main
  - prod по тэгу/релизу



