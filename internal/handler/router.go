package handler

import (
	"github.com/churilovmn1/workout-tracker/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter sets up all HTTP routes.
func NewRouter(
	authService *service.AuthService,
	exerciseService *service.ExerciseService,
	workoutService *service.WorkoutService,
	templateService *service.TemplateService,
	webDir string,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	authHandler := NewAuthHandler(authService)
	exerciseHandler := NewExerciseHandler(exerciseService)
	workoutHandler := NewWorkoutHandler(workoutService)
	templateHandler := NewTemplateHandler(templateService, workoutService)
	webHandler := NewWebHandler(webDir)

	r.Get("/", webHandler.Index)
	r.Handle("/static/*", webHandler.StaticHandler())

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(authService))

			r.Route("/exercises", func(r chi.Router) {
				r.Get("/", exerciseHandler.List)
				r.Get("/{id}", exerciseHandler.GetByID)

				r.Group(func(r chi.Router) {
					r.Use(AdminOnly)
					r.Post("/", exerciseHandler.Create)
					r.Put("/{id}", exerciseHandler.Update)
					r.Delete("/{id}", exerciseHandler.Delete)
				})
			})

			r.Route("/workouts", func(r chi.Router) {
				r.Get("/", workoutHandler.List)
				r.Post("/", workoutHandler.Create)
				r.Get("/{id}", workoutHandler.GetByID)
				r.Put("/{id}", workoutHandler.Update)
				r.Delete("/{id}", workoutHandler.Delete)
				r.Post("/{id}/copy", workoutHandler.Copy)
			})

			r.Route("/templates", func(r chi.Router) {
				r.Get("/", templateHandler.List)
				r.Post("/", templateHandler.Create)
				r.Get("/{id}", templateHandler.GetByID)
				r.Put("/{id}", templateHandler.Update)
				r.Delete("/{id}", templateHandler.Delete)
				r.Post("/{id}/start", templateHandler.Start)
			})

			r.Route("/stats", func(r chi.Router) {
				r.Get("/pr", workoutHandler.PersonalRecords)
				r.Get("/volume", workoutHandler.WeeklyVolume)
			})
		})
	})

	return r
}
