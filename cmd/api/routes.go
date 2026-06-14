package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", app.healthz)
	mux.HandleFunc("POST /tasks", app.createTask)
	mux.HandleFunc("GET /tasks", app.listTasks)
	mux.HandleFunc("GET /tasks/{id}", app.getTask)
	mux.HandleFunc("PUT /tasks/{id}", app.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", app.deleteTask)

	return mux
}
