package main

import (
	"errors"
	"net/http"

	"github.com/MonicaMell/task-api/internal/repository"
	"github.com/MonicaMell/task-api/internal/service"
)

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	var in service.RegisterInput
	if err := readJSON(w, r, &in); err != nil {
		app.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := app.auth.Register(r.Context(), in)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			app.errorJSON(w, http.StatusConflict, "email already registered")
			return
		}
		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var in service.LoginInput
	if err := readJSON(w, r, &in); err != nil {
		app.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := app.auth.Login(r.Context(), in)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			app.errorJSON(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		app.serverError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
