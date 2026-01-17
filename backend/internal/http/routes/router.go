package routes

import (
	"net/http"
	"time"

	"novels-backend/internal/config"
	"novels-backend/internal/http/handlers"
	"novels-backend/internal/http/middleware"
	"novels-backend/internal/repository"
	"novels-backend/internal/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// NewRouter создает новый роутер с настроенными маршрутами
func NewRouter(db *sqlx.DB, cfg *config.Config, log zerolog.Logger) http.Handler {
	r := chi.NewRouter()

	// Глобальные middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Logger(log))
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "Accept-Language"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	novelRepo := repository.NewNovelRepository(db)
	chapterRepo := repository.NewChapterRepository(db)
	progressRepo := repository.NewProgressRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	bookmarkRepo := repository.NewBookmarkRepository(db)
	xpRepo := repository.NewXPRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	votingRepo := repository.NewVotingRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	newsRepo := repository.NewNewsRepository(db)
	wikiEditRepo := repository.NewWikiEditRepository(db)
	authorRepo := repository.NewAuthorRepository(db)
	genreRepo := repository.NewGenreRepository(db)
	tagRepo := repository.NewTagRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg)
	xpService := service.NewXPService(xpRepo)
	novelService := service.NewNovelService(novelRepo)
	chapterService := service.NewChapterService(chapterRepo, novelRepo, progressRepo)
	commentService := service.NewCommentService(commentRepo, xpService)
	bookmarkService := service.NewBookmarkService(bookmarkRepo, novelRepo, xpService)
	ticketService := service.NewTicketService(ticketRepo, subscriptionRepo, log)
	votingService := service.NewVotingService(votingRepo, ticketRepo, log)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, ticketRepo, log)
	collectionService := service.NewCollectionService(collectionRepo, novelRepo, userRepo)
	newsService := service.NewNewsService(newsRepo, userRepo)
	wikiEditService := service.NewWikiEditService(wikiEditRepo, novelRepo, userRepo, subscriptionService)
	authorService := service.NewAuthorService(authorRepo)
	genreService := service.NewGenreService(genreRepo)
	tagService := service.NewTagService(tagRepo)
	adminService := service.NewAdminService(adminRepo)

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService)
	novelHandler := handlers.NewNovelHandler(novelService)
	chapterHandler := handlers.NewChapterHandler(chapterService)
	adminHandler := handlers.NewAdminHandler(novelService, chapterService, "./uploads")
	commentHandler := handlers.NewCommentHandler(commentService)
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkService)
	walletHandler := handlers.NewWalletHandler(ticketService, log)
	votingHandler := handlers.NewVotingHandler(votingService, log)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, log)
	collectionHandler := handlers.NewCollectionHandler(collectionService)
	newsHandler := handlers.NewNewsHandler(newsService)
	wikiEditHandler := handlers.NewWikiEditHandler(wikiEditService)
	authorHandler := handlers.NewAuthorAdminHandler(authorService)
	genreTagHandler := handlers.NewGenreTagAdminHandler(genreService, tagService)
	userAdminHandler := handlers.NewUserAdminHandler(userRepo)
	commentAdminHandler := handlers.NewCommentAdminHandler(commentRepo)
	adminSystemHandler := handlers.NewAdminSystemHandler(adminService)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, cfg.JWT)

	// Маршруты
	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok", "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		})

		// Публичные маршруты
		r.Group(func(r chi.Router) {
			// Опциональная аутентификация для получения userID
			r.Use(authMiddleware.OptionalAuth)

			// Аутентификация
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/refresh", authHandler.Refresh)

			// Каталог новелл
			r.Get("/novels", novelHandler.List)
			r.Get("/novels/{slug}", novelHandler.GetBySlug)
			r.Get("/novels/{slug}/chapters", chapterHandler.ListByNovel)
			
			// Главы
			r.Get("/chapters/{id}", chapterHandler.GetByID)

			// Теги и жанры
			r.Get("/genres", novelHandler.GetGenres)
			r.Get("/tags", novelHandler.GetTags)

			// Комментарии (публичное чтение)
			r.Get("/comments", commentHandler.List)
			r.Get("/comments/{id}", commentHandler.GetByID)
			r.Get("/comments/{id}/replies", commentHandler.GetReplies)

			// Публичные данные голосования
			r.Get("/voting/leaderboard", votingHandler.GetVotingLeaderboard)
			r.Get("/voting/proposals", votingHandler.GetVotingProposals)
			r.Get("/voting/stats", votingHandler.GetVotingStats)

			// Публичные данные подписок
			r.Get("/subscriptions/plans", subscriptionHandler.GetPlans)
			r.Get("/subscriptions/plans/{id}", subscriptionHandler.GetPlan)

			// Публичный лидерборд
			r.Get("/leaderboard", walletHandler.GetLeaderboard)

			// Коллекции (публичные)
			r.Get("/collections", collectionHandler.List)
			r.Get("/collections/popular", collectionHandler.GetPopular)
			r.Get("/collections/featured", collectionHandler.GetFeatured)
			r.Get("/collections/{id}", collectionHandler.GetByID)

			// Новости (публичные)
			r.Get("/news", newsHandler.List)
			r.Get("/news/latest", newsHandler.GetLatest)
			r.Get("/news/pinned", newsHandler.GetPinned)
			r.Get("/news/{slug}", newsHandler.GetBySlug)

			// Платформенная статистика
			r.Get("/stats/platform", wikiEditHandler.GetPlatformStats)

			// История правок для новеллы (публичная)
			r.Get("/novels/{id}/edit-history", wikiEditHandler.GetNovelEditHistory)
		})

		// Защищенные маршруты (требуют аутентификации)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// Профиль
			r.Get("/auth/me", authHandler.Me)
			r.Post("/auth/logout", authHandler.Logout)

			// Прогресс чтения (через chapterHandler)
			r.Get("/novels/{slug}/progress", chapterHandler.GetProgress)
			r.Post("/chapters/{id}/progress", chapterHandler.SaveProgress)

			// Комментарии (защищенные операции)
			r.Post("/comments", commentHandler.Create)
			r.Put("/comments/{id}", commentHandler.Update)
			r.Delete("/comments/{id}", commentHandler.Delete)
			r.Post("/comments/{id}/vote", commentHandler.Vote)
			r.Post("/comments/{id}/report", commentHandler.Report)

			// Закладки
			r.Get("/bookmarks", bookmarkHandler.List)
			r.Get("/bookmarks/lists", bookmarkHandler.GetLists)
			r.Get("/bookmarks/stats", bookmarkHandler.GetStats)
			r.Get("/bookmarks/status/{novelId}", bookmarkHandler.GetNovelStatus)
			r.Post("/bookmarks", bookmarkHandler.Add)
			r.Put("/bookmarks/{novelId}", bookmarkHandler.Update)
			r.Delete("/bookmarks/{novelId}", bookmarkHandler.Remove)

			// Кошелек и билеты
			r.Get("/wallet", walletHandler.GetWallet)
			r.Get("/wallet/transactions", walletHandler.GetTransactions)
			r.Get("/wallet/stats", walletHandler.GetStats)

			// Предложки новелл
			r.Get("/proposals", votingHandler.ListProposals)
			r.Get("/proposals/my", votingHandler.GetMyProposals)
			r.Get("/proposals/{id}", votingHandler.GetProposal)
			r.Post("/proposals", votingHandler.CreateProposal)
			r.Put("/proposals/{id}", votingHandler.UpdateProposal)
			r.Post("/proposals/{id}/submit", votingHandler.SubmitProposal)
			r.Delete("/proposals/{id}", votingHandler.DeleteProposal)

			// Голосование
			r.Post("/votes", votingHandler.CastVote)

			// Подписки пользователя
			r.Get("/subscriptions/me", subscriptionHandler.GetMySubscription)
			r.Get("/subscriptions/history", subscriptionHandler.GetMySubscriptionHistory)
			r.Get("/subscriptions/premium", subscriptionHandler.IsPremium)
			r.Get("/subscriptions/features/{feature}", subscriptionHandler.CheckFeature)
			r.Post("/subscriptions", subscriptionHandler.Subscribe)
			r.Post("/subscriptions/{id}/cancel", subscriptionHandler.CancelSubscription)

			// Коллекции (пользовательские)
			r.Post("/collections", collectionHandler.Create)
			r.Put("/collections/{id}", collectionHandler.Update)
			r.Delete("/collections/{id}", collectionHandler.Delete)
			r.Post("/collections/{id}/items", collectionHandler.AddItem)
			r.Delete("/collections/{id}/items/{novelId}", collectionHandler.RemoveItem)
			r.Put("/collections/{id}/items/reorder", collectionHandler.ReorderItems)
			r.Post("/collections/{id}/vote", collectionHandler.Vote)
			r.Get("/users/{id}/collections", collectionHandler.GetUserCollections)

			// Wiki редактирование (Premium)
			r.Post("/novels/{id}/edit-requests", wikiEditHandler.CreateEditRequest)
			r.Get("/novels/{id}/edit-requests", wikiEditHandler.GetNovelEditRequests)
			r.Get("/edit-requests/{id}", wikiEditHandler.GetEditRequest)
			r.Post("/edit-requests/{id}/cancel", wikiEditHandler.CancelEditRequest)
			r.Get("/me/edit-requests", wikiEditHandler.GetUserEditRequests)
		})

		// Маршруты модерации
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(authMiddleware.RequireRole("moderator", "admin"))

			r.Route("/moderation", func(r chi.Router) {
				// Модерация предложек
				r.Get("/proposals", votingHandler.GetPendingProposals)
				r.Post("/proposals/{id}", votingHandler.ModerateProposal)

				// Модерация wiki правок
				r.Get("/edit-requests", wikiEditHandler.GetPendingEditRequests)
				r.Post("/edit-requests/{id}/approve", wikiEditHandler.ApproveEditRequest)
				r.Post("/edit-requests/{id}/reject", wikiEditHandler.RejectEditRequest)
			})
		})

		// Административные маршруты
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(authMiddleware.RequireRole("admin"))

			r.Route("/admin", func(r chi.Router) {
				// Управление новеллами
				r.Post("/novels", adminHandler.CreateNovel)
				r.Put("/novels/{id}", adminHandler.UpdateNovel)
				r.Delete("/novels/{id}", adminHandler.DeleteNovel)

				// Управление главами
				r.Get("/chapters", adminHandler.ListChapters)
				r.Post("/chapters", adminHandler.CreateChapter)
				r.Put("/chapters/{id}", adminHandler.UpdateChapter)
				r.Delete("/chapters/{id}", adminHandler.DeleteChapter)

				// Загрузка файлов
				r.Post("/upload", adminHandler.Upload)

				// Управление билетами
				r.Post("/tickets/grant", walletHandler.GrantTickets)
				r.Get("/users/{userId}/wallet", walletHandler.GetUserWallet)

				// Управление подписками
				r.Get("/subscriptions/stats", subscriptionHandler.GetSubscriptionStats)
				r.Get("/users/{userId}/subscription", subscriptionHandler.GetUserSubscription)

				// Управление коллекциями (featured)
				r.Post("/collections/{id}/featured", collectionHandler.SetFeatured)

				// Управление новостями
				r.Get("/news", newsHandler.ListAdmin)
				r.Get("/news/{slug}", newsHandler.GetAdminBySlug)
				r.Post("/news", newsHandler.Create)
				r.Put("/news/{id}", newsHandler.Update)
				r.Delete("/news/{id}", newsHandler.Delete)
				r.Post("/news/{id}/publish", newsHandler.Publish)
				r.Post("/news/{id}/unpublish", newsHandler.Unpublish)
				r.Post("/news/{id}/pin", newsHandler.SetPinned)
				r.Put("/news/{id}/localizations/{lang}", newsHandler.SetLocalization)
				r.Delete("/news/{id}/localizations/{lang}", newsHandler.DeleteLocalization)

				// Управление авторами
				r.Get("/authors", authorHandler.ListAuthors)
				r.Post("/authors", authorHandler.CreateAuthor)
				r.Get("/authors/{id}", authorHandler.GetAuthor)
				r.Put("/authors/{id}", authorHandler.UpdateAuthor)
				r.Delete("/authors/{id}", authorHandler.DeleteAuthor)
				r.Get("/novels/{id}/authors", authorHandler.GetNovelAuthors)
				r.Put("/novels/{id}/authors", authorHandler.UpdateNovelAuthors)

				// Управление жанрами
				r.Get("/genres", genreTagHandler.ListGenres)
				r.Post("/genres", genreTagHandler.CreateGenre)
				r.Get("/genres/{id}", genreTagHandler.GetGenre)
				r.Put("/genres/{id}", genreTagHandler.UpdateGenre)
				r.Delete("/genres/{id}", genreTagHandler.DeleteGenre)

				// Управление тегами
				r.Get("/tags", genreTagHandler.ListTags)
				r.Post("/tags", genreTagHandler.CreateTag)
				r.Get("/tags/{id}", genreTagHandler.GetTag)
				r.Put("/tags/{id}", genreTagHandler.UpdateTag)
				r.Delete("/tags/{id}", genreTagHandler.DeleteTag)

				// Управление пользователями
				r.Get("/users", userAdminHandler.ListUsers)
				r.Get("/users/{id}", userAdminHandler.GetUser)
				r.Post("/users/{id}/ban", userAdminHandler.BanUser)
				r.Post("/users/{id}/unban", userAdminHandler.UnbanUser)
				r.Put("/users/{id}/roles", userAdminHandler.UpdateUserRoles)

				// Управление комментариями и жалобами
				r.Get("/comments", commentAdminHandler.ListComments)
				r.Delete("/comments/{id}", commentAdminHandler.SoftDeleteComment)
				r.Delete("/comments/{id}/hard", commentAdminHandler.HardDeleteComment)
				r.Get("/reports", commentAdminHandler.ListReports)
				r.Post("/reports/{id}/resolve", commentAdminHandler.ResolveReport)

				// Системные функции
				r.Get("/settings", adminSystemHandler.GetSettings)
				r.Get("/settings/{key}", adminSystemHandler.GetSetting)
				r.Put("/settings/{key}", adminSystemHandler.UpdateSetting)
				r.Get("/logs", adminSystemHandler.GetLogs)
				r.Get("/stats", adminSystemHandler.GetStats)
			})
		})
	})

	// Статические файлы (загруженные изображения)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	return r
}
