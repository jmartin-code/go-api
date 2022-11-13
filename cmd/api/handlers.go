package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"go-api/internal/data"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mozillazg/go-slugify"
)

var staticPath = "./static/"

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
		app.errorJSON(w, errors.New("invalid password"))
		return
	}

	if user.Active == 0 {
		app.errorJSON(w, errors.New("inactive User"))
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
		u.Active = user.Active

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

func (app *application) LogUserOutAndSetInactive(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	user, err := app.models.User.GetUserById(userId)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	user.Active = 0
	err = user.UpdateUser()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.models.Token.DeleteTokenForUser(userId)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "User is logged out and set to inactive",
	}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Token string `json:"token"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	valid := false

	valid, _ = app.models.Token.ValidToken(requestPayload.Token)

	payload := jsonResponse{
		Error: false,
		Data:  valid,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

// Books
func (app *application) AllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := app.models.Book.GetAll()

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "success",
		Data:    envelope{"books": books},
	}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) OneBook(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	book, err := app.models.Book.GetOneBySlug(slug)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "success",
		Data:    envelope{"book": book},
	}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllAuthors(w http.ResponseWriter, r *http.Request) {
	authors, err := app.models.Author.All()

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	type selectData struct {
		Value int    `json:"value"`
		Text  string `json:"text"`
	}

	var results []selectData

	for _, author := range authors {
		results = append(results, selectData{Value: author.ID, Text: author.AuthorName})
	}

	payload := jsonResponse{
		Error:   false,
		Message: "success",
		Data:    results,
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) EditBook(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		ID              int    `json:"id"`
		Title           string `json:"title"`
		AuthorID        int    `json:"author_id"`
		PublicationYear int    `json:"publication_year"`
		Description     string `json:"description"`
		CoverBase64     string `json:"cover"`
		genreIDs        []int  `json:"genre_ids"`
	}

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err)
		return
	}

	book := data.Book{
		ID:              requestPayload.ID,
		Title:           requestPayload.Title,
		AuthorID:        requestPayload.AuthorID,
		PublicationYear: requestPayload.PublicationYear,
		Description:     requestPayload.Description,
		Slug:            slugify.Slugify(requestPayload.Title),
		GenreIDs:        requestPayload.genreIDs,
	}

	if len(requestPayload.CoverBase64) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(requestPayload.CoverBase64)
		if err != nil {
			app.errorJSON(w, err)
			return
		}

		if err := os.WriteFile(fmt.Sprintf("%s/covers/%s.jpg", staticPath, book.Slug), decoded, 0666); err != nil {
			app.errorJSON(w, err)
			return
		}
	}

	if book.ID == 0 {
		// add book
		_, err := app.models.Book.Insert(book)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	} else {
		// update book
		if err := book.Update(); err != nil {
			app.errorJSON(w, err)
			return
		}

	}

	payload := jsonResponse{
		Error:   false,
		Message: "Changes Saved",
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}
