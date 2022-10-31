package main

import (
	"go-api/internal/data"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Post("/api/login", app.Login)

	mux.Get("/api/user/all", func(w http.ResponseWriter, r *http.Request) {
		var users data.User
		all, err := users.GetAll()

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		app.writeJSON(w, http.StatusOK, all)
	})

	return mux
}
