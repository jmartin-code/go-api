package main

import (
	"errors"
	"go-api/internal/data"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type envelope map[string]interface{}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	type credentials struct {
		Username string `json:"email"`
		Password string `json:"password"`
	}

	var creds credentials
	var payload jsonResponse

	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorLog.Println(err)
		payload.Error = true
		payload.Message = "invalid json data"
		_ = app.writeJSON(w, http.StatusBadRequest, payload)
	}

	app.infoLog.Println(creds.Username, creds.Password)
	user, err := app.models.User.GetUserByEmail(creds.Username)

	if err != nil {
		println("failed here")
		app.errorJSON(w, errors.New("invalid username"))
		return
	}

	validPassword, err := user.UserPasswordMatch(creds.Password)
	if err != nil || !validPassword {
		println("or failed here")
		app.errorJSON(w, errors.New("invalid password"))
		return
	}

	token, err := app.models.Token.GenerateToken(user.ID, 24*time.Hour)
	if err != nil {
		app.errorJSON(w, err)
	}

	err = app.models.Token.InsertToken(*token, *user)
	if err != nil {
		app.errorJSON(w, err)
	}

	payload = jsonResponse{
		Error:   false,
		Message: "logged in",
		Data:    envelope{"token": token, "user": user},
	}

	err = app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Token string `json:"token"`
	}

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, errors.New("invalid json"))
		return
	}

	err = app.models.Token.DeleteByToken(requestPayload.Token)
	if err != nil {
		app.errorJSON(w, errors.New("invalid json"))
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "logged out",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)

}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("all users")
	var users data.User
	all, err := users.GetAllUsers()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "Success",
		Data:    envelope{"users": all},
	}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) EditUser(w http.ResponseWriter, r *http.Request) {
	var user data.User

	err := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if user.ID == 0 {
		// save new user
		if _, err := app.models.User.AddUser(user); err != nil {
			if err != nil {
				app.errorJSON(w, err)
				return
			}
		}
	} else {
		// update user
		u, err := app.models.User.GetUserById(user.ID)
		if err != nil {
			app.errorJSON(w, err)
			return
		}

		u.Email = user.Email
		u.FirstName = user.FirstName
		u.LastName = user.LastName

		if err := u.UpdateUser(); err != nil {
			app.errorJSON(w, err)
			return
		}

		// update password
		if user.Password != "" {
			err := u.ResetUserPassword(user.Password)
			if err != nil {
				app.errorJSON(w, err)
				return
			}
		}

	}

	payload := jsonResponse{
		Error:   false,
		Message: "Changes saved!",
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *application) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		app.errorJSON(w, err)
	}

	user, err := app.models.User.GetUserById(id)
	if err != nil {
		app.errorJSON(w, err)
	}

	_ = app.writeJSON(w, http.StatusOK, user)
}

func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var payloadId struct {
		ID int `json:"id"`
	}

	err := app.readJSON(w, r, &payloadId)
	if err != nil {
		app.errorJSON(w, err)
	}

	err = app.models.User.DeleteUserById(payloadId.ID)
	if err != nil {
		app.errorJSON(w, err)
	}

	payload := jsonResponse{
		Error:   false,
		Message: "User deleted",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}
