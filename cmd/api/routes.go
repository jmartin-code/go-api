package main

import (
	"go-api/internal/data"
	"net/http"
	"time"

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

	mux.Route("/api/admin", func(mux chi.Router) {
		// AUTHENTICATED ROUTES
		mux.Use(app.AuthTokenMiddleware)

		// All users
		mux.Post("/users", app.AllUsers)
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

	mux.Get("/api/test-generate-token", func(w http.ResponseWriter, r *http.Request) {
		token, err := app.models.Token.GenerateToken(2, 60*time.Minute)

		if err != nil {
			app.infoLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
			return
		}

		token.Email = "example@test.com"
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}

		app.writeJSON(w, http.StatusOK, payload)

	})

	mux.Get("/api/test-save-token", func(w http.ResponseWriter, r *http.Request) {
		token, err := app.models.Token.GenerateToken(2, 60*time.Minute)

		if err != nil {
			app.infoLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
			return
		}

		user, err := app.models.User.GetUserById(2)
		if err != nil {
			app.infoLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
			return
		}

		token.UserID = user.ID
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		err = token.InsertToken(*token, *user)
		if err != nil {
			app.infoLog.Println(err)
			app.errorJSON(w, err, http.StatusForbidden)
			return
		}

		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}

		app.writeJSON(w, http.StatusOK, payload)
	})

	mux.Get("/api/test-validate-token", func(w http.ResponseWriter, r *http.Request) {

		tokenToValidate := r.URL.Query().Get("token")
		valid, err := app.models.Token.ValidToken(tokenToValidate)

		if err != nil {
			app.errorJSON(w, err)
			return
		}

		payload := jsonResponse{
			Error: false,
			Data:  valid,
		}

		app.writeJSON(w, http.StatusOK, payload)
	})

	return mux
}
