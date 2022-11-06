package main

import (
	"net/http"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Println("admin")
		_, err := app.models.Token.AuthenticateToken(r)
		if err != nil {
			payload := jsonResponse{
				Error:   true,
				Message: "user is not logged in, invalid credentials",
			}
			_ = app.writeJSON(w, http.StatusUnauthorized, payload)
			return
		}
		next.ServeHTTP(w, r)
	})
}
