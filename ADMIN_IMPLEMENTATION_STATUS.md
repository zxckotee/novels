# Статус реализации админ-панели

## ✅ Завершено

### База данных
- **006_authors.sql** - Таблицы для авторов с локализациями и связь с новеллами
- **007_comments_unify.sql** - Унификация схемы комментариев (target_type, target_id, body, is_deleted, root_id, depth)
- **008_admin_settings_and_audit.sql** - Настройки приложения и логи действий администраторов

### Модели Go
- **models/author.go** - Author, AuthorLocalization, NovelAuthor + DTO
- **models/admin.go** - AppSetting, AdminAuditLog, AdminStatsOverview, фильтры для админ-функций
- **models/genre_tag.go** - GenreWithLocalizations, TagWithLocalizations + DTO для CRUD

### Репозитории
- **repository/author_repository.go** - Полный CRUD для авторов с локализациями

### Сервисы
- **service/author_service.go** - Бизнес-логика для авторов

### Хендлеры
- **handlers/author_admin_handler.go** - REST API для управления авторами
- **handlers/genre_tag_admin_handler.go** - Каркас для жанров/тегов (требует реализации репозиториев)

### Фронтенд - исправления
- Исправлены все админ-страницы: заменен `isModerator` на `isAdmin`
- Использован правильный паттерн с `useEffect` для редиректа

## 🚧 TODO: Бэкенд

### 1. Репозитории и сервисы для жанров/тегов
Создать аналогично `author_repository.go`:
- `backend/internal/repository/genre_repository.go`
- `backend/internal/repository/tag_repository.go`
- `backend/internal/service/genre_service.go`
- `backend/internal/service/tag_service.go`

### 2. Репозитории и сервисы для админ-функций
```
backend/internal/repository/admin_repository.go
  - GetSettings()
  - UpdateSetting()
  - GetAuditLogs()
  - LogAction()
  - GetStats()

backend/internal/service/admin_service.go
  - Бизнес-логика для settings/logs/stats
```

### 3. Расширить UserRepository/Service
Добавить методы:
- `ListUsers(filter UsersFilter)`
- `BanUser(userID, reason)`
- `UnbanUser(userID)`
- `UpdateUserRoles(userID, roles)`

### 4. Расширить CommentRepository/Service  
Добавить админ-методы:
- `AdminListComments(filter AdminCommentsFilter)`
- `SoftDeleteComment(commentID)`
- `HardDeleteComment(commentID)`
- `GetReports(filter ReportsFilter)`
- `ResolveReport(reportID, action)`

### 5. Создать хендлеры
```
handlers/user_admin_handler.go
  - ListUsers, GetUser, BanUser, UnbanUser, UpdateRoles

handlers/comment_admin_handler.go
  - ListComments, DeleteComment (soft/hard)
  - ListReports, ResolveReport, DismissReport

handlers/admin_system_handler.go
  - GetSettings, UpdateSetting
  - GetLogs
  - GetStats
```

### 6. Обновить роутер
В `backend/internal/http/routes/router.go` добавить маршруты:
```go
r.Route("/admin", func(r chi.Router) {
    r.Use(middleware.RequireRole("admin"))
    
    // Authors
    r.Get("/authors", authorAdminHandler.ListAuthors)
    r.Post("/authors", authorAdminHandler.CreateAuthor)
    r.Get("/authors/{id}", authorAdminHandler.GetAuthor)
    r.Put("/authors/{id}", authorAdminHandler.UpdateAuthor)
    r.Delete("/authors/{id}", authorAdminHandler.DeleteAuthor)
    
    // Novels X Authors
    r.Get("/novels/{id}/authors", authorAdminHandler.GetNovelAuthors)
    r.Put("/novels/{id}/authors", authorAdminHandler.UpdateNovelAuthors)
    
    // Genres
    r.Get("/genres", genreTagHandler.ListGenres)
    r.Post("/genres", genreTagHandler.CreateGenre)
    r.Put("/genres/{id}", genreTagHandler.UpdateGenre)
    r.Delete("/genres/{id}", genreTagHandler.DeleteGenre)
    
    // Tags
    r.Get("/tags", genreTagHandler.ListTags)
    r.Post("/tags", genreTagHandler.CreateTag)
    r.Put("/tags/{id}", genreTagHandler.UpdateTag)
    r.Delete("/tags/{id}", genreTagHandler.DeleteTag)
    
    // Users
    r.Get("/users", userAdminHandler.ListUsers)
    r.Get("/users/{id}", userAdminHandler.GetUser)
    r.Post("/users/{id}/ban", userAdminHandler.BanUser)
    r.Post("/users/{id}/unban", userAdminHandler.UnbanUser)
    r.Put("/users/{id}/roles", userAdminHandler.UpdateRoles)
    
    // Comments & Reports
    r.Get("/comments", commentAdminHandler.ListComments)
    r.Delete("/comments/{id}", commentAdminHandler.SoftDeleteComment)
    r.Delete("/comments/{id}/hard", commentAdminHandler.HardDeleteComment)
    r.Get("/reports", commentAdminHandler.ListReports)
    r.Post("/reports/{id}/resolve", commentAdminHandler.ResolveReport)
    
    // System
    r.Get("/settings", adminSystemHandler.GetSettings)
    r.Put("/settings/{key}", adminSystemHandler.UpdateSetting)
    r.Get("/logs", adminSystemHandler.GetLogs)
    r.Get("/stats", adminSystemHandler.GetStats)
})
```

## 🚧 TODO: Фронтенд

### 1. Обновить API типы
В `frontend/src/lib/api/types.ts` добавить:
```typescript
// Authors
export interface Author { ... }
export interface CreateAuthorRequest { ... }

// Genres/Tags
export interface Genre { ... }
export interface Tag { ... }

// Admin
export interface AdminUser { ... }
export interface AdminComment { ... }
export interface AdminReport { ... }
export interface AdminSetting { ... }
export interface AdminLog { ... }
export interface AdminStats { ... }
```

### 2. Создать API хуки
```
frontend/src/lib/api/hooks/useAdminAuthors.ts
frontend/src/lib/api/hooks/useAdminGenres.ts
frontend/src/lib/api/hooks/useAdminTags.ts
frontend/src/lib/api/hooks/useAdminUsers.ts
frontend/src/lib/api/hooks/useAdminComments.ts
frontend/src/lib/api/hooks/useAdminSettings.ts
```

### 3. Создать админ-страницы
```
frontend/src/app/[locale]/admin/authors/page.tsx
frontend/src/app/[locale]/admin/genres/page.tsx
frontend/src/app/[locale]/admin/tags/page.tsx
frontend/src/app/[locale]/admin/users/page.tsx
frontend/src/app/[locale]/admin/comments/page.tsx
frontend/src/app/[locale]/admin/reports/page.tsx
frontend/src/app/[locale]/admin/news/page.tsx (использовать существующий API)
frontend/src/app/[locale]/admin/settings/page.tsx
frontend/src/app/[locale]/admin/logs/page.tsx
frontend/src/app/[locale]/admin/stats/page.tsx
frontend/src/app/[locale]/admin/popular/page.tsx
```

Каждая страница должна:
- Использовать `isAdmin` guard
- React Query для данных и мутаций
- Таблицу с пагинацией
- Поиск/фильтры
- Модалки Create/Edit
- Действия (delete/ban/resolve)

## 📝 Примечания

- Все миграции БД готовы к применению
- Авторы полностью реализованы (модели, репозиторий, сервис, хендлеры)
- Жанры/теги требуют реализации репозиториев (по аналогии с авторами)
- Существующие эндпоинты для новелл/глав/новостей уже в роутере
- Комментарии после миграции 007 готовы к админ-управлению

## 🔄 Следующие шаги

1. Применить миграции: `docker-compose exec backend ./api migrate`
2. Завершить репозитории для жанров/тегов
3. Создать админ-хендлеры для users/comments/settings
4. Обновить роутер
5. Создать фронтенд-страницы с UI компонентами
