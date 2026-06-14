package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/MonicaMell/task-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("taskdb_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Println("start postgres container:", err)
		os.Exit(1)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Println("connection string:", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Println("connect pool:", err)
		os.Exit(1)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		fmt.Println("apply migrations:", err)
		os.Exit(1)
	}

	cfg := &config.Config{
		Addr:        "8080",
		DatabaseURL: dsn,
		JWTSecret:   "test-secret-not-for-production",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // silence logs in tests

	app := newApplication(cfg, logger, pool)
	testServer = httptest.NewServer(app.routes())

	code := m.Run()

	testServer.Close()
	pool.Close()
	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return err
		}
	}
	return nil
}

func doJSON(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req, err := http.NewRequest(method, testServer.URL+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decode(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(dst))
}

func registerAndLogin(t *testing.T, email, password string) string {
	t.Helper()
	creds := map[string]string{"email": email, "password": password}

	resp := doJSON(t, http.MethodPost, "/auth/register", "", creds)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = doJSON(t, http.MethodPost, "/auth/login", "", creds)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out map[string]string
	decode(t, resp, &out)
	require.NotEmpty(t, out["token"])
	return out["token"]
}

func TestAuthRequired(t *testing.T) {
	resp := doJSON(t, http.MethodGet, "/tasks", "", nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestValidationRejectsBadInput(t *testing.T) {
	resp := doJSON(t, http.MethodPost, "/auth/register", "", map[string]string{
		"email": "not-an-email", "password": "short",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestTaskCRUD(t *testing.T) {
	token := registerAndLogin(t, "crud@example.com", "supersecret123")

	resp := doJSON(t, http.MethodPost, "/tasks", token, map[string]any{
		"title": "write tests", "description": "integration",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decode(t, resp, &created)
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)
	require.Equal(t, "todo", created["status"])

	resp = doJSON(t, http.MethodGet, "/tasks/"+id, token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = doJSON(t, http.MethodPut, "/tasks/"+id, token, map[string]any{
		"title": "write tests", "description": "done", "status": "done",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var updated map[string]any
	decode(t, resp, &updated)
	require.Equal(t, "done", updated["status"])

	resp = doJSON(t, http.MethodDelete, "/tasks/"+id, token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	resp = doJSON(t, http.MethodGet, "/tasks/"+id, token, nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestTasksAreOwnerScoped(t *testing.T) {
	tokenA := registerAndLogin(t, "owner-a@example.com", "supersecret123")
	tokenB := registerAndLogin(t, "owner-b@example.com", "supersecret123")

	resp := doJSON(t, http.MethodPost, "/tasks", tokenA, map[string]any{"title": "A's secret task"})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created map[string]any
	decode(t, resp, &created)
	id := created["id"].(string)

	resp = doJSON(t, http.MethodGet, "/tasks/"+id, tokenB, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestNonUUIDTaskIDIsRejected(t *testing.T) {
	token := registerAndLogin(t, "uuid@example.com", "supersecret123")
	resp := doJSON(t, http.MethodGet, "/tasks/not-a-uuid", token, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
