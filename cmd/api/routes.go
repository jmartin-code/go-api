package main

import (
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
	mux.Post("/api/logout", app.Logout)
	mux.Get("/api/books", app.AllBooks)
	mux.Get("/api/books/{slug}", app.OneBook)

	mux.Post("/api/validate-token", app.ValidateToken)

	mux.Route("/api/admin", func(mux chi.Router) {
		// AUTHENTICATED ROUTES
		mux.Use(app.AuthTokenMiddleware)

		// Users
		mux.Post("/users", app.AllUsers)
		mux.Post("/users/save", app.EditUser)
		mux.Post("/users/get/{id}", app.GetUser)
		mux.Post("/users/delete", app.DeleteUser)
		mux.Post("/users/user-logout/{id}", app.LogUserOutAndSetInactive)

		// Books
		mux.Post("/authors", app.AllAuthors)
		mux.Post("/books/save", app.EditBook)

	})

	//static
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
