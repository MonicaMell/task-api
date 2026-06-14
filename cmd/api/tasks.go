package main

import (
	"errors"
	"net/http"

	"github.com/MonicaMell/task-api/internal/repository"
	"github.com/MonicaMell/task-api/internal/service"
)

func (app *application) currentUserID(r *http.Request) string {
	userID, _ := r.Context().Value(userIDKey).(string)
	return userID
}

func (app *application) validTaskID(w http.ResponseWriter, id string) bool {
	if err := app.validate.Var(id, "required,uuid"); err != nil {
		app.errorJSON(w, http.StatusBadRequest, "invalid task id")
		return false
	}
	return true
}

func (app *application) createTask(w http.ResponseWriter, r *http.Request) {
	var in service.CreateTaskInput
	if err := readJSON(w, r, &in); err != nil {
		app.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := app.validate.Struct(in); err != nil {
		app.failedValidation(w, err)
		return
	}

	task, err := app.tasks.Create(r.Context(), app.currentUserID(r), in)
	if err != nil {
		app.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (app *application) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := app.tasks.List(r.Context(), app.currentUserID(r))
	if err != nil {
		app.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (app *application) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !app.validTaskID(w, id) {
		return
	}

	task, err := app.tasks.Get(r.Context(), app.currentUserID(r), id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			app.errorJSON(w, http.StatusNotFound, "task not found")
			return
		}
		app.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (app *application) updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !app.validTaskID(w, id) {
		return
	}

	var in service.UpdateTaskInput
	if err := readJSON(w, r, &in); err != nil {
		app.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := app.validate.Struct(in); err != nil {
		app.failedValidation(w, err)
		return
	}

	task, err := app.tasks.Update(r.Context(), app.currentUserID(r), id, in)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			app.errorJSON(w, http.StatusNotFound, "task not found")
			return
		}
		app.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (app *application) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !app.validTaskID(w, id) {
		return
	}

	err := app.tasks.Delete(r.Context(), app.currentUserID(r), id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			app.errorJSON(w, http.StatusNotFound, "task not found")
			return
		}
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
