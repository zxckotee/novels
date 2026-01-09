# План технической реализации платформы Novels

## Обзор проекта

**Цель**: Создание многоязычной платформы для чтения новелл с экономической системой тикетов, голосования и комьюнити-функциями.

**Технологический стек**:
- Backend: Go + Chi router
- Database: PostgreSQL 
- Frontend: Next.js (React с SSR)
- Deploy: Docker + Docker Compose

**Ключевые особенности**:
- Поддержка 7 языков интерфейса
- SEO-оптимизированные страницы (SSR)
- Экономическая система с тикетами и голосованиями
- Комьюнити-функции (комментарии, коллекции, вики-редактирование)

## Архитектурная диаграмма

```mermaid
graph TB
    subgraph "Frontend (Next.js)"
        A[Web App SSR]
        B[Admin Panel]
        C[Mobile Responsive]
    end
    
    subgraph "Backend (Go + Chi)"
        D[API Gateway]
        E[Auth Service]
        F[Content Service]
        G[Economy Service]
        H[Comment Service]
        I[Background Jobs]
    end
    
    subgraph "Data Layer"
        J[(PostgreSQL)]
        K[Object Storage]
        L[Redis Cache]
    end
    
    A --> D
    B --> D
    C --> D
    
    D --> E
    D --> F
    D --> G
    D --> H
    
    E --> J
    F --> J
    G --> J
    H --> J
    
    F --> K
    D --> L
    
    I --> J
    I --> L
```

## ЭТАП A: MVP 1 - База платформы

### 1. Проектная структура

```
novels/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── routes/
│   │   ├── domain/
│   │   │   ├── models/
│   │   │   ├── repositories/
│   │   │   └── services/
│   │   ├── database/
│   │   │   ├── migrations/
│   │   │   └── queries/
│   │   ├── auth/
│   │   ├── config/
│   │   └── jobs/
│   ├── pkg/
│   │   ├── logger/
│   │   ├── validator/
│   │   └── utils/
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── components/
│   │   ├── ui/
│   │   ├── layout/
│   │   ├── novel/
│   │   └── reader/
│   ├── pages/
│   │   ├── api/ (если нужны Next.js API routes)
│   │   ├── [locale]/
│   │   │   ├── index.js
│   │   │   ├── catalog.js
│   │   │   ├── novel/
│   │   │   └── read/
│   │   └── _app.js
│   ├── styles/
│   ├── public/
│   │   └── locales/
│   │       ├── ru/
│   │       ├── en/
│   │       └── ...
│   ├── utils/
│   ├── hooks/
│   ├── context/
│   ├── package.json
│   └── Dockerfile
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   └── nginx/
└── docs/
    ├── api/
    └── db-schema/
```

### 2. Схема базы данных PostgreSQL

```mermaid
erDiagram
    novels {
        uuid id PK
        varchar slug UK
        varchar cover_image_key
        enum translation_status
        int original_chapters_count
        int release_year
        timestamp created_at
        timestamp updated_at
    }
    
    novel_localizations {
        uuid novel_id PK,FK
        varchar lang PK
        varchar title
        text description
        text alt_titles
        timestamp created_at
        timestamp updated_at
    }
    
    tags {
        uuid id PK
        varchar name UK
        varchar slug UK
    }
    
    genres {
        uuid id PK
        varchar name UK
        varchar slug UK
    }
    
    novel_tags {
        uuid novel_id PK,FK
        uuid tag_id PK,FK
    }
    
    novel_genres {
        uuid novel_id PK,FK
        uuid genre_id PK,FK
    }
    
    chapters {
        uuid id PK
        uuid novel_id FK
        int number
        varchar slug
        varchar title
        timestamp published_at
        timestamp created_at
        timestamp updated_at
    }
    
    chapter_contents {
        uuid chapter_id PK,FK
        varchar lang PK
        text content
        varchar source
        timestamp updated_at
    }
    
    users {
        uuid id PK
        varchar email UK
        varchar password_hash
        boolean is_banned
        timestamp last_login_at
        timestamp created_at
        timestamp updated_at
    }
    
    user_profiles {
        uuid user_id PK,FK
        varchar display_name
        varchar avatar_key
        text bio
        timestamp created_at
        timestamp updated_at
    }
    
    user_roles {
        uuid user_id PK,FK
        enum role PK
        timestamp created_at
    }
    
    reading_progress {
        uuid user_id PK,FK
        uuid novel_id PK,FK
        uuid chapter_id FK
        int position
        timestamp updated_at
    }
    
    novels ||--o{ novel_localizations : "has"
    novels ||--o{ novel_tags : "has"
    novels ||--o{ novel_genres : "has"
    novels ||--o{ chapters : "contains"
    chapters ||--|| chapter_contents : "has"
    users ||--|| user_profiles : "has"
    users ||--o{ user_roles : "has"
    users ||--o{ reading_progress : "tracks"
    tags ||--o{ novel_tags : "tagged"
    genres ||--o{ novel_genres : "categorized"
```

### 3. API спецификации MVP 1

**Base URL**: `/api/v1`

#### Публичные эндпоинты

```yaml
GET /novels:
  description: Получение списка новелл с фильтрами
  parameters:
    - lang: string (required)
    - page: int (default: 1)
    - limit: int (default: 20, max: 100)
    - sort: enum [updated_at, created_at, views, rating]
    - order: enum [asc, desc] (default: desc)
    - status: enum [ongoing, completed, paused, dropped]
    - genres: string[] (comma-separated genre slugs)
    - tags: string[] (comma-separated tag slugs)
    - search: string (full-text search)
  response:
    novels: Novel[]
    pagination: PaginationMeta
    total: int

GET /novels/{slug}:
  description: Получение детальной информации о новелле
  parameters:
    - slug: string (required)
    - lang: string (required)
  response: NovelDetail

GET /novels/{slug}/chapters:
  description: Получение списка глав новеллы
  parameters:
    - slug: string (required)
    - page: int (default: 1)
    - limit: int (default: 50)
  response:
    chapters: Chapter[]
    pagination: PaginationMeta

GET /chapters/{id}:
  description: Получение содержимого главы
  parameters:
    - id: uuid (required)
    - lang: string (required)
  response: ChapterContent

GET /tags:
  description: Получение списка тегов
  parameters:
    - lang: string (required)
  response: Tag[]

GET /genres:
  description: Получение списка жанров
  parameters:
    - lang: string (required)
  response: Genre[]
```

#### Аутентификация

```yaml
POST /auth/register:
  body:
    email: string (required)
    password: string (required, min: 8)
    display_name: string (required)
  response:
    user: User
    access_token: string
    refresh_token: string (httpOnly cookie)

POST /auth/login:
  body:
    email: string (required)
    password: string (required)
  response:
    user: User
    access_token: string
    refresh_token: string (httpOnly cookie)

POST /auth/refresh:
  headers:
    Cookie: refresh_token
  response:
    access_token: string

POST /auth/logout:
  headers:
    Authorization: Bearer {access_token}
  response:
    message: string

GET /auth/me:
  headers:
    Authorization: Bearer {access_token}
  response: UserProfile
```

#### Пользовательские функции

```yaml
POST /progress:
  headers:
    Authorization: Bearer {access_token}
  body:
    novel_id: uuid (required)
    chapter_id: uuid (required)
    position: int (default: 0)
  response:
    message: string

GET /progress/{novel_id}:
  headers:
    Authorization: Bearer {access_token}
  parameters:
    - novel_id: uuid (required)
  response: ReadingProgress
```

#### Админские эндпоинты

```yaml
POST /admin/novels:
  headers:
    Authorization: Bearer {admin_token}
  body:
    slug: string (required)
    translation_status: enum (required)
    original_chapters_count: int
    release_year: int
    cover_image: file
    localizations: LocalizationData[]
    tags: string[] (tag slugs)
    genres: string[] (genre slugs)
  response: Novel

PUT /admin/novels/{id}:
  headers:
    Authorization: Bearer {admin_token}
  body: (same as POST /admin/novels)
  response: Novel

POST /admin/chapters:
  headers:
    Authorization: Bearer {admin_token}
  body:
    novel_id: uuid (required)
    number: int (required)
    title: string (required)
    slug: string
    contents: ChapterContentData[]
  response: Chapter

POST /admin/upload:
  headers:
    Authorization: Bearer {admin_token}
  body:
    file: multipart/form-data
    type: enum [cover, avatar]
  response:
    key: string
    url: string
```

### 4. Конфигурация Next.js с i18n

**next.config.js**:
```javascript
const { i18n } = require('./next-i18next.config')

module.exports = {
  i18n,
  experimental: {
    esmExternals: false
  },
  images: {
    domains: ['localhost', process.env.API_BASE_URL]
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.API_BASE_URL}/api/:path*`
      }
    ]
  }
}
```

**next-i18next.config.js**:
```javascript
module.exports = {
  i18n: {
    defaultLocale: 'ru',
    locales: ['ru', 'en', 'zh', 'ja', 'ko', 'fr', 'de'],
    localePath: './public/locales',
    localeDetection: false
  }
}
```

### 5. Middleware и безопасность

```go
// internal/http/middleware/auth.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        claims, err := validateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_role", claims.Role)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole := r.Context().Value("user_role").(string)
            
            for _, role := range allowedRoles {
                if userRole == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            
            http.Error(w, "Forbidden", http.StatusForbidden)
        })
    }
}

func RateLimitMiddleware(requests int, duration time.Duration) func(http.Handler) http.Handler {
    // Реализация rate limiting с использованием Redis или in-memory store
}
```

## ЭТАП B: MVP 2 - Комментарии, закладки, уровни

### 1. Расширение схемы БД

```mermaid
erDiagram
    bookmark_lists {
        uuid id PK
        uuid user_id FK
        varchar code
        varchar title
        int sort_order
        timestamp created_at
    }
    
    bookmarks {
        uuid user_id PK,FK
        uuid novel_id PK,FK
        uuid list_id FK
        timestamp created_at
        timestamp updated_at
    }
    
    comments {
        uuid id PK
        uuid parent_id FK
        uuid root_id FK
        int depth
        enum target_type
        uuid target_id
        uuid user_id FK
        text body
        boolean is_deleted
        timestamp created_at
        timestamp updated_at
    }
    
    comment_votes {
        uuid comment_id PK,FK
        uuid user_id PK,FK
        int value
        timestamp created_at
    }
    
    user_xp {
        uuid user_id PK,FK
        bigint xp_total
        int level
        timestamp updated_at
    }
    
    xp_events {
        uuid id PK
        uuid user_id FK
        enum type
        int delta
        enum ref_type
        uuid ref_id
        timestamp created_at
    }
    
    users ||--o{ bookmark_lists : "owns"
    users ||--o{ bookmarks : "creates"
    novels ||--o{ bookmarks : "bookmarked"
    bookmark_lists ||--o{ bookmarks : "contains"
    
    users ||--o{ comments : "writes"
    comments ||--o{ comments : "replies"
    comments ||--o{ comment_votes : "voted"
    
    users ||--|| user_xp : "has"
    users ||--o{ xp_events : "earns"
```

### 2. API эндпоинты для MVP 2

```yaml
# Закладки
GET /bookmarks:
  headers:
    Authorization: Bearer {access_token}
  parameters:
    - list_code: string (reading, planned, dropped, read, favorites)
    - page: int (default: 1)
    - limit: int (default: 20)
    - sort: enum [updated_at, created_at, title]
  response:
    bookmarks: BookmarkItem[]
    lists: BookmarkList[]

POST /bookmarks:
  headers:
    Authorization: Bearer {access_token}
  body:
    novel_id: uuid (required)
    list_code: string (required)
  response: BookmarkItem

PUT /bookmarks/{novel_id}:
  headers:
    Authorization: Bearer {access_token}
  body:
    list_code: string (required)
  response: BookmarkItem

DELETE /bookmarks/{novel_id}:
  headers:
    Authorization: Bearer {access_token}
  response:
    message: string

# Комментарии
GET /comments:
  parameters:
    - target_type: enum [novel, chapter] (required)
    - target_id: uuid (required)
    - page: int (default: 1)
    - limit: int (default: 50)
    - sort: enum [created_at, score] (default: created_at)
  response:
    comments: Comment[]
    pagination: PaginationMeta

POST /comments:
  headers:
    Authorization: Bearer {access_token}
  body:
    target_type: enum [novel, chapter] (required)
    target_id: uuid (required)
    parent_id: uuid (optional)
    body: string (required)
  response: Comment

PUT /comments/{id}:
  headers:
    Authorization: Bearer {access_token}
  body:
    body: string (required)
  response: Comment

DELETE /comments/{id}:
  headers:
    Authorization: Bearer {access_token}
  response:
    message: string

POST /comments/{id}/vote:
  headers:
    Authorization: Bearer {access_token}
  body:
    value: int (1 or -1)
  response:
    score: int
```

## ЭТАП C: Economy Update - Тикеты и экономика

### 1. Схема экономической системы

```mermaid
erDiagram
    subscription_plans {
        uuid id PK
        varchar code UK
        varchar title
        decimal price
        enum period
        boolean is_active
        json benefits
        timestamp created_at
    }
    
    subscriptions {
        uuid id PK
        uuid user_id FK
        uuid plan_id FK
        enum status
        timestamp starts_at
        timestamp ends_at
        timestamp created_at
        timestamp updated_at
    }
    
    ticket_balances {
        uuid user_id PK,FK
        enum type PK
        int balance
        timestamp updated_at
    }
    
    ticket_transactions {
        uuid id PK
        uuid user_id FK
        enum type
        int delta
        varchar reason
        enum ref_type
        uuid ref_id
        varchar idempotency_key UK
        timestamp created_at
    }
    
    novel_proposals {
        uuid id PK
        uuid user_id FK
        varchar original_link
        enum status
        json metadata
        timestamp created_at
        timestamp updated_at
    }
    
    voting_polls {
        uuid id PK
        enum status
        timestamp starts_at
        timestamp ends_at
        timestamp created_at
        timestamp updated_at
    }
    
    voting_entries {
        uuid poll_id PK,FK
        uuid proposal_id PK,FK
        int score
        timestamp updated_at
    }
    
    votes {
        uuid id PK
        uuid poll_id FK
        uuid user_id FK
        uuid proposal_id FK
        enum ticket_type
        int amount
        timestamp created_at
    }
    
    users ||--o{ subscriptions : "subscribes"
    subscription_plans ||--o{ subscriptions : "provides"
    
    users ||--o{ ticket_balances : "owns"
    users ||--o{ ticket_transactions : "makes"
    
    users ||--o{ novel_proposals : "proposes"
    novel_proposals ||--o{ voting_entries : "entered"
    voting_polls ||--o{ voting_entries : "contains"
    voting_polls ||--o{ votes : "includes"
```

### 2. Background Jobs (Cron задачи)

```go
// internal/jobs/daily_vote_granter.go
type DailyVoteGranter struct {
    userRepo   repositories.UserRepository
    ticketRepo repositories.TicketRepository
    logger     logger.Logger
}

func (d *DailyVoteGranter) Run(ctx context.Context) error {
    // Запускается каждый день в 00:00 UTC (03:00 MSK)
    users, err := d.userRepo.GetActiveUsers(ctx)
    if err != nil {
        return fmt.Errorf("failed to get active users: %w", err)
    }
    
    for _, user := range users {
        multiplier := d.getVoteMultiplier(user)
        dailyVotes := 1 * multiplier
        
        idempotencyKey := fmt.Sprintf("daily_vote_%s_%s", 
            time.Now().Format("2006-01-02"), user.ID)
        
        err := d.ticketRepo.GrantTickets(ctx, &domain.TicketTransaction{
            UserID:           user.ID,
            Type:             domain.TicketTypeDailyVote,
            Delta:            dailyVotes,
            Reason:           "Daily vote grant",
            IdempotencyKey:   idempotencyKey,
        })
        
        if err != nil {
            d.logger.Error("Failed to grant daily votes", 
                "user_id", user.ID, "error", err)
        }
    }
    
    return nil
}

// internal/jobs/voting_winner_picker.go
type VotingWinnerPicker struct {
    proposalRepo repositories.ProposalRepository
    novelRepo    repositories.NovelRepository
    pollRepo     repositories.PollRepository
    logger       logger.Logger
}

func (v *VotingWinnerPicker) Run(ctx context.Context) error {
    // Запускается каждые X часов
    activePoll, err := v.pollRepo.GetActivePoll(ctx)
    if err != nil || activePoll == nil {
        return nil // Нет активного голосования
    }
    
    winner, err := v.pollRepo.GetWinnerEntry(ctx, activePoll.ID)
    if err != nil {
        return fmt.Errorf("failed to get winner: %w", err)
    }
    
    if winner == nil {
        return nil // Нет голосов
    }
    
    // Создаем новеллу из предложения
    novel, err := v.createNovelFromProposal(ctx, winner.ProposalID)
    if err != nil {
        return fmt.Errorf("failed to create novel: %w", err)
    }
    
    // Закрываем голосование
    err = v.pollRepo.ClosePoll(ctx, activePoll.ID)
    if err != nil {
        return fmt.Errorf("failed to close poll: %w", err)
    }
    
    v.logger.Info("Voting winner selected", 
        "proposal_id", winner.ProposalID, 
        "novel_id", novel.ID)
    
    return nil
}
```

## ЭТАП D: Community Update

### 1. Коллекции и вики-редактирование

```mermaid
erDiagram
    collections {
        uuid id PK
        uuid user_id FK
        varchar slug UK
        varchar title
        text description
        timestamp created_at
        timestamp updated_at
    }
    
    collection_items {
        uuid collection_id PK,FK
        uuid novel_id PK,FK
        int position
        text user_comment
        timestamp added_at
    }
    
    collection_votes {
        uuid collection_id PK,FK
        uuid user_id PK,FK
        int value
        timestamp created_at
    }
    
    novel_edit_requests {
        uuid id PK
        uuid novel_id FK
        uuid user_id FK
        enum status
        json changes
        text moderator_notes
        timestamp created_at
        timestamp updated_at
    }
    
    news_posts {
        uuid id PK
        varchar slug UK
        varchar title
        text body
        uuid author_id FK
        timestamp published_at
        timestamp created_at
        timestamp updated_at
    }
    
    users ||--o{ collections : "creates"
    novels ||--o{ collection_items : "included"
    collections ||--o{ collection_items : "contains"
    collections ||--o{ collection_votes : "voted"
    
    novels ||--o{ novel_edit_requests : "edited"
    users ||--o{ novel_edit_requests : "submits"
    
    users ||--o{ news_posts : "writes"
```

## Docker конфигурация

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: novels
      POSTGRES_USER: novels
      POSTGRES_PASSWORD: novels_dev_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    environment:
      DATABASE_URL: postgres://novels:novels_dev_password@postgres:5432/novels?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev_jwt_secret_change_in_production
      PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    volumes:
      - ./uploads:/app/uploads

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    environment:
      NEXT_PUBLIC_API_BASE_URL: http://localhost:8080
      NODE_ENV: development
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
```

## Критерии готовности MVP

### MVP 1 (База платформы)
- ✅ Пользователь может найти новеллу в каталоге
- ✅ Открыть страницу новеллы с описанием  
- ✅ Выбрать и читать главы в ридере
- ✅ Прогресс чтения сохраняется
- ✅ Админ может добавлять новеллы и главы
- ✅ Поддерживается 7 языков интерфейса
- ✅ SEO-страницы корректно индексируются

### MVP 2 (Комьюнити)
- ✅ Nested комментарии работают стабильно
- ✅ Закладки с фильтрацией и сортировкой
- ✅ XP и уровни корректно начисляются и отображаются

### Economy Update
- ✅ Daily Vote начисляется автоматически в 03:00 МСК
- ✅ Система предложки новелл → голосование → автовыбор работает
- ✅ Подписки активируют все привилегии
- ✅ Защита от накрутки базово реализована

### Community Update
- ✅ Коллекции создаются и голосуются
- ✅ Вики-редактирование проходит модерацию
- ✅ Новости публикуются модераторами
- ✅ Производительность не страдает при нагрузке

## Безопасность и производительность

### Обязательные меры безопасности
- Хеширование паролей с bcrypt
- JWT токены с коротким TTL + refresh токены
- RBAC проверки на всех эндпоинтах
- Rate limiting на критичных операциях
- SQL injection защита через prepared statements
- XSS защита через санитизацию HTML
- CSRF защита для cookie-based аутентификации

### Оптимизации производительности
- Индексы на часто запрашиваемые поля
- Пагинация для всех списков
- Кеширование популярных запросов
- Compression для статических ресурсов
- CDN для изображений
- Database connection pooling

## Процесс деплоя

### Development
```bash
docker-compose up -d
cd frontend && npm run dev
cd backend && go run cmd/api/main.go
```

### Production
```bash
docker-compose -f docker-compose.prod.yml up -d
# + настройка Nginx для SSL termination  
# + мониторинг (Prometheus/Grafana)
# + бэкапы PostgreSQL
```

Этот план обеспечивает поэтапную разработку с четкими критериями готовности каждого этапа и возможностью масштабирования в будущем.