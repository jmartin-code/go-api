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
		all, err := users.GetAllUsers()

		if err != nil {
			app.errorLog.Println(err)
			return
		}

		app.writeJSON(w, http.StatusOK, all)
	})

	mux.Get("/api/user/add", func(w http.ResponseWriter, r *http.Request) {
		var u = data.User{
			Email:     "testing@testing.com",
			FirstName: "John",
			LastName:  "Martin",
			Password:  "test",
		}

		app.infoLog.Println("Adding a new user")

		id, err := app.models.User.AddUser(u)
		if err != nil {
			app.errorLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
		}

		app.infoLog.Println("got the id of user as", id)
		newUser, _ := app.models.User.GetUserById(id)

		app.writeJSON(w, http.StatusOK, newUser)

	})

	return mux
}
