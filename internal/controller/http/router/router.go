package router

import (
	"log/slog"
	"net/http"
	generate "pvz-service/api"
	api "pvz-service/api/generated"
	"pvz-service/internal/controller/http/handler"
	httpMiddleware "pvz-service/internal/controller/http/middleware"
	"pvz-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Setup(
	log *slog.Logger,
	authService service.AuthService,
	pvzService service.PVZService,
) http.Handler {
	h := handler.NewHandler(log, &authService, &pvzService)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(httpMiddleware.LoggerMiddleware(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Public routes
	router.Group(func(r chi.Router) {
		// Раздача Swagger UI из embed
		r.Handle("/*", http.FileServer(http.FS(generate.APIEmbeddedFiles)))

		r.Post("/dummyLogin", h.DummyLogin())
		r.Post("/register", h.Register())
		r.Post("/login", h.Login())
	})

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(httpMiddleware.AuthMiddleware(&authService))

		// Routes for all auth users
		r.Group(func(r chi.Router) {
			r.Get("/pvz", h.GetPVZsWithReceptions())
		})

		// Routes for role='moderator'
		r.Group(func(r chi.Router) {
			r.Use(httpMiddleware.RoleMiddlewareMulti(api.UserRoleModerator))

			r.Post("/pvz", h.CreatePVZ())
		})

		// Routes for role='employee'
		r.Group(func(r chi.Router) {
			r.Use(httpMiddleware.RoleMiddlewareMulti(api.UserRoleEmployee))

			r.Post("/receptions", h.StartReception())
			r.Post("/products", h.AddProduct())
			r.Post("/pvz/{pvzId}/delete_last_product", h.DeleteLastProduct())
			r.Post("/pvz/{pvzId}/close_last_reception", h.CloseReception())
		})

	})

	return router
}
