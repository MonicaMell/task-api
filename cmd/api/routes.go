package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", app.healthz)
	mux.HandleFunc("POST /auth/register", app.register)
	mux.HandleFunc("POST /auth/login", app.login)

	mux.Handle("POST /tasks", app.requireAuth(http.HandlerFunc(app.createTask)))
	mux.Handle("GET /tasks", app.requireAuth(http.HandlerFunc(app.listTasks)))
	mux.Handle("GET /tasks/{id}", app.requireAuth(http.HandlerFunc(app.getTask)))
	mux.Handle("PUT /tasks/{id}", app.requireAuth(http.HandlerFunc(app.updateTask)))
	mux.Handle("DELETE /tasks/{id}", app.requireAuth(http.HandlerFunc(app.deleteTask)))

	return mux
}
