# API Спецификация платформы Novels

## Базовая информация

**Base URL**: `/api/v1`  
**Аутентификация**: Bearer Token (JWT)  
**Формат данных**: JSON  
**Коды ошибок**: HTTP Status + структурированные ошибки

## Стандартные форматы ответов

### Успешный ответ
```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2024-02-15T10:30:00Z",
    "version": "1.0"
  }
}
```

### Ошибка
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "meta": {
    "timestamp": "2024-02-15T10:30:00Z",
    "request_id": "req_123456789"
  }
}
```

### Пагинация
```json
{
  "data": [...],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total_pages": 15,
    "total_items": 299,
    "has_next": true,
    "has_prev": false
  }
}
```

## Типы данных

### User
```typescript
interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  roles: UserRole[];
  created_at: string;
  last_login_at: string;
}

interface UserProfile extends User {
  bio?: string;
  level: number;
  xp_total: number;
  xp_to_next_level: number;
  subscription?: Subscription;
  stats: UserStats;
}

interface UserStats {
  chapters_read: number;
  reading_time_minutes: number;
  comments_count: number;
  bookmarks_count: number;
  novels_proposed: number;
  tickets_spent: number;
}
```

### Novel
```typescript
interface Novel {
  id: string;
  slug: string;
  cover_image_url: string;
  translation_status: 'ongoing' | 'completed' | 'paused' | 'dropped';
  original_chapters_count: number;
  release_year: number;
  views_daily: number;
  views_total: number;
  rating: number;
  rating_count: number;
  bookmarks_count: number;
  created_at: string;
  updated_at: string;
  // Локализованные данные (зависят от lang параметра)
  title: string;
  description: string;
  alt_titles: string[];
  author: string;
  tags: Tag[];
  genres: Genre[];
}

interface NovelDetail extends Novel {
  chapters_published: number;
  last_chapter_at: string;
  user_progress?: ReadingProgress;
  user_bookmark?: Bookmark;
}
```

### Chapter
```typescript
interface Chapter {
  id: string;
  novel_id: string;
  number: number;
  slug?: string;
  title: string;
  published_at: string;
  views: number;
  comments_count: number;
  created_at: string;
}

interface ChapterContent extends Chapter {
  content: string;
  source: 'manual' | 'auto' | 'import';
  word_count: number;
  reading_time_minutes: number;
  prev_chapter?: { id: string; number: number; title: string };
  next_chapter?: { id: string; number: number; title: string };
}
```

### Comment
```typescript
interface Comment {
  id: string;
  parent_id?: string;
  root_id?: string;
  depth: number;
  target_type: 'novel' | 'chapter' | 'news';
  target_id: string;
  user: {
    id: string;
    display_name: string;
    avatar_url?: string;
    level: number;
  };
  body: string;
  score: number;
  user_vote?: 1 | -1;
  is_edited: boolean;
  is_deleted: boolean;
  replies_count: number;
  created_at: string;
  updated_at: string;
}
```

### Ticket & Economy
```typescript
interface TicketBalance {
  type: 'daily_vote' | 'novel_request' | 'translation_ticket';
  balance: number;
  updated_at: string;
}

interface TicketTransaction {
  id: string;
  type: 'daily_vote' | 'novel_request' | 'translation_ticket';
  delta: number;
  reason: string;
  ref_type?: 'vote' | 'proposal' | 'subscription' | 'level_up';
  ref_id?: string;
  created_at: string;
}

interface NovelProposal {
  id: string;
  user: {
    id: string;
    display_name: string;
  };
  original_link: string;
  status: 'draft' | 'moderation' | 'voting' | 'rejected' | 'accepted';
  title: string;
  description: string;
  author: string;
  tags: string[];
  genres: string[];
  votes_count: number;
  created_at: string;
  updated_at: string;
}

interface VotingPoll {
  id: string;
  status: 'active' | 'closed';
  starts_at: string;
  ends_at: string;
  entries: VotingEntry[];
}

interface VotingEntry {
  proposal: NovelProposal;
  score: number;
  position: number;
}
```

## Эндпоинты API

### Аутентификация

#### POST /auth/register
Регистрация нового пользователя

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "display_name": "UserName"
}
```

**Response (201):**
```json
{
  "data": {
    "user": { User },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

**Errors:**
- 400: Email уже зарегистрирован
- 422: Валидация не пройдена

#### POST /auth/login
Вход в систему

**Request Body:**
```json
{
  "email": "user@example.com", 
  "password": "securePassword123"
}
```

**Response (200):** аналогично регистрации

#### POST /auth/refresh  
Обновление токена доступа

**Request:** Refresh token в httpOnly cookie

**Response (200):**
```json
{
  "data": {
    "access_token": "newJWTtoken...",
    "expires_in": 3600
  }
}
```

#### GET /auth/me
Получение профиля текущего пользователя  

**Headers:** `Authorization: Bearer {token}`

**Response (200):**
```json
{
  "data": { UserProfile }
}
```

### Публичные эндпоинты

#### GET /novels
Получение списка новелл с фильтрацией

**Query Parameters:**
- `lang` (required): ru, en, zh, ja, ko, fr, de
- `page` (default: 1): номер страницы  
- `limit` (default: 20, max: 100): количество на странице
- `sort` (default: updated_at): updated_at, created_at, views_daily, views_total, rating, bookmarks_count
- `order` (default: desc): asc, desc
- `status`: ongoing, completed, paused, dropped
- `genres[]`: массив slug жанров
- `tags[]`: массив slug тегов  
- `search`: текстовый поиск
- `year_from`, `year_to`: диапазон года выпуска

**Response (200):**
```json
{
  "data": {
    "novels": [ Novel ],
    "filters": {
      "genres": [ Genre ],
      "tags": [ Tag ],
      "years": [2020, 2021, 2022, 2023, 2024]
    }
  },
  "pagination": { PaginationMeta }
}
```

#### GET /novels/{slug}
Детальная информация о новелле

**Path Parameters:**
- `slug`: уникальный идентификатор новеллы

**Query Parameters:**  
- `lang` (required): язык локализации

**Response (200):**
```json
{
  "data": { NovelDetail }
}
```

**Errors:**
- 404: Новелла не найдена

#### GET /novels/{slug}/chapters
Список глав новеллы 

**Query Parameters:**
- `page`, `limit`: пагинация
- `sort` (default: number): number, created_at, views
- `order` (default: asc): asc, desc

**Response (200):**
```json
{
  "data": {
    "chapters": [ Chapter ],
    "novel": { Novel }
  },
  "pagination": { PaginationMeta }
}
```

#### GET /chapters/{id}
Содержимое главы

**Path Parameters:**
- `id`: ID главы

**Query Parameters:**
- `lang` (required): язык контента

**Response (200):**
```json
{
  "data": { ChapterContent }
}
```

**Errors:**
- 404: Глава не найдена
- 403: Глава недоступна (если будет платная модель)

#### GET /tags
Список всех тегов

**Query Parameters:**
- `lang` (required): язык 

**Response (200):**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Магия",
      "slug": "magic",
      "novels_count": 127
    }
  ]
}
```

#### GET /genres  
Список всех жанров (аналогично тегам)

### Пользовательские функции

#### POST /progress
Сохранение прогресса чтения

**Headers:** `Authorization: Bearer {token}`

**Request Body:**
```json
{
  "novel_id": "uuid",
  "chapter_id": "uuid", 
  "position": 1250
}
```

**Response (200):**
```json
{
  "data": {
    "message": "Progress saved",
    "progress": { ReadingProgress }
  }
}
```

#### GET /progress/{novel_id}
Получение прогресса по новелле

**Response (200):**
```json
{
  "data": {
    "novel_id": "uuid",
    "chapter_id": "uuid",
    "chapter_number": 15,
    "position": 1250,
    "updated_at": "2024-02-15T10:30:00Z"
  }
}
```

#### GET /bookmarks  
Список закладок пользователя

**Query Parameters:**
- `list_code`: reading, planned, dropped, completed, favorites
- `page`, `limit`: пагинация
- `sort` (default: updated_at): updated_at, created_at, title, rating

**Response (200):**
```json
{
  "data": {
    "bookmarks": [
      {
        "novel": { Novel },
        "list": {
          "code": "reading",
          "title": "Читаю"
        },
        "progress": { ReadingProgress },
        "created_at": "2024-02-15T10:30:00Z",
        "updated_at": "2024-02-15T10:30:00Z"
      }
    ],
    "lists": [
      {
        "code": "reading", 
        "title": "Читаю",
        "count": 5
      }
    ]
  },
  "pagination": { PaginationMeta }
}
```

#### POST /bookmarks
Добавление в закладки

**Request Body:**
```json
{
  "novel_id": "uuid",
  "list_code": "reading"
}
```

#### PUT /bookmarks/{novel_id}
Перемещение между списками закладок  

**Request Body:**
```json
{
  "list_code": "completed"
}
```

#### DELETE /bookmarks/{novel_id}
Удаление из закладок

### Комментарии

#### GET /comments
Получение комментариев

**Query Parameters:**
- `target_type` (required): novel, chapter, news
- `target_id` (required): ID цели
- `page`, `limit`: пагинация  
- `sort` (default: created_at): created_at, score
- `order` (default: asc для created_at, desc для score)

**Response (200):**
```json
{
  "data": [ Comment ],
  "pagination": { PaginationMeta }
}
```

#### POST /comments
Создание комментария

**Headers:** `Authorization: Bearer {token}`

**Request Body:**
```json
{
  "target_type": "novel",
  "target_id": "uuid",
  "parent_id": "uuid", // опционально для ответов
  "body": "Отличная новелла!"
}
```

**Response (201):**
```json
{
  "data": { Comment }
}
```

#### PUT /comments/{id}
Редактирование комментария (только автор)

**Request Body:**
```json
{
  "body": "Исправленный комментарий"
}
```

#### DELETE /comments/{id}  
Удаление комментария (автор или модератор)

#### POST /comments/{id}/vote
Голосование за комментарий

**Request Body:**
```json
{
  "value": 1 // 1 для лайка, -1 для дизлайка
}
```

**Response (200):**
```json
{
  "data": {
    "comment_id": "uuid",
    "new_score": 15,
    "user_vote": 1
  }
}
```

#### POST /comments/{id}/report
Жалоба на комментарий

**Request Body:**
```json
{
  "reason": "spam", // spam, offensive, inappropriate, other
  "details": "Подробности жалобы"
}
```

### Экономическая система

#### GET /wallet
Баланс билетов пользователя

**Headers:** `Authorization: Bearer {token}`

**Response (200):**
```json
{
  "data": {
    "balances": [ TicketBalance ],
    "recent_transactions": [ TicketTransaction ]
  }
}
```

#### GET /wallet/transactions
История транзакций  

**Query Parameters:**
- `type`: фильтр по типу билета
- `page`, `limit`: пагинация

#### POST /votes
Голосование за предложенную новеллу

**Request Body:**
```json
{
  "proposal_id": "uuid",
  "ticket_type": "daily_vote",
  "amount": 1
}
```

**Response (200):** 
```json
{
  "data": {
    "message": "Vote cast successfully",
    "new_score": 25,
    "remaining_balance": 0
  }
}
```

#### GET /polls/current
Текущее голосование

**Response (200):**
```json
{
  "data": { VotingPoll }
}
```

#### POST /proposals
Предложение новой новеллы (требует Novel Request)

**Headers:** `Authorization: Bearer {token}`

**Request Body:**
```json
{
  "original_link": "https://example.com/novel",
  "title": "Название новеллы", 
  "description": "Описание...",
  "author": "Автор",
  "tags": ["tag1", "tag2"],
  "genres": ["fantasy", "adventure"],
  "release_year": 2023
}
```

**Response (201):**
```json
{
  "data": { NovelProposal }
}
```

#### GET /proposals
Мои предложения

**Query Parameters:**
- `status`: фильтр по статусу
- `page`, `limit`: пагинация

### Коллекции (Community Update)

#### GET /collections
Публичные коллекции

**Query Parameters:**
- `sort` (default: votes): votes, created_at, updated_at
- `page`, `limit`: пагинация

#### POST /collections
Создание коллекции

**Request Body:**
```json
{
  "title": "Лучшее фэнтези 2024",
  "description": "Моя подборка топовых фэнтези новелл",
  "novels": [
    {
      "novel_id": "uuid",
      "comment": "Отличная магическая система"
    }
  ]
}
```

#### GET /collections/{id}
Детали коллекции

#### POST /collections/{id}/vote  
Голосование за коллекцию

**Request Body:**
```json
{
  "value": 1 // только положительные голоса
}
```

### Административные функции

#### POST /admin/novels
Создание новеллы

**Headers:** `Authorization: Bearer {admin_token}`

**Request Body:**
```json
{
  "slug": "novel-slug",
  "translation_status": "ongoing",
  "original_chapters_count": 100,
  "release_year": 2023,
  "cover_image": "base64_encoded_image",
  "localizations": [
    {
      "lang": "ru",
      "title": "Название",
      "description": "Описание...",
      "alt_titles": ["Альт название"],
      "author": "Автор"
    }
  ],
  "tags": ["tag1", "tag2"],
  "genres": ["fantasy"]
}
```

#### POST /admin/chapters
Добавление главы

**Request Body:**
```json
{
  "novel_id": "uuid",
  "number": 1,
  "title": "Глава 1",
  "slug": "chapter-1",
  "contents": [
    {
      "lang": "ru", 
      "content": "Текст главы...",
      "source": "manual"
    }
  ]
}
```

#### GET /admin/moderation/comments
Очередь модерации комментариев

#### POST /admin/moderation/comments/{id}/action
Действие модерации

**Request Body:**
```json
{
  "action": "delete", // delete, ban_user, warn_user, approve
  "reason": "Причина действия",
  "ban_duration": 7 // дни, только для ban_user
}
```

## Коды ошибок

### Стандартные HTTP коды
- 200: OK
- 201: Created  
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 422: Validation Error
- 429: Rate Limit Exceeded
- 500: Internal Server Error

### Кастомные коды ошибок
- `VALIDATION_ERROR`: Ошибка валидации данных
- `INSUFFICIENT_BALANCE`: Недостаточно билетов  
- `DUPLICATE_VOTE`: Повторное голосование
- `NOVEL_NOT_FOUND`: Новелла не найдена
- `CHAPTER_NOT_FOUND`: Глава не найдена  
- `ACCESS_DENIED`: Нет прав доступа
- `RATE_LIMIT_EXCEEDED`: Превышен лимит запросов
- `ACCOUNT_BANNED`: Аккаунт заблокирован
- `FEATURE_REQUIRES_PREMIUM`: Функция требует подписку

## Rate Limiting

### Публичные эндпоинты
- Поиск/каталог: 60 запросов/минуту на IP
- Чтение глав: 120 запросов/минуту на IP  

### Авторизованные пользователи  
- Комментарии: 10 создания/час
- Голосования: 50 голосов/день
- Закладки: 100 операций/час

### Административные
- Создание контента: 100 запросов/час

## Версионирование

API использует семантическое версионирование в URL (`/api/v1`). При breaking changes создается новая версия API с поддержкой предыдущей версии минимум 6 месяцев.

## Безопасность

### Аутентификация
- JWT токены с коротким TTL (1 час)
- Refresh токены в httpOnly cookies
- Автоматический logout при бездействии

### Валидация
- Все входные данные валидируются на бэкенде
- SQL injection защита через prepared statements  
- XSS защита через экранирование HTML
- Rate limiting для предотвращения злоупотреблений

### Логирование
- Все административные действия логируются
- IP адреса сохраняются для аудита
- Чувствительные данные не логируются