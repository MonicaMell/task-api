package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MonicaMell/task-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) error {
	query := `
		INSERT INTO tasks (user_id, title, description, status, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, t.UserID, t.Title, t.Description, t.Status, t.DueDate).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, userID, taskID string) (*model.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2`

	var t model.Task
	err := r.db.QueryRow(ctx, query, taskID, userID).Scan(&t.ID, &t.UserID,
		&t.Title, &t.Description, &t.Status, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &t, nil
}

func (r *TaskRepository) ListByUser(ctx context.Context, userID string) ([]model.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, due_date, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	tasks := make([]model.Task, 0)

	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status,
			&t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *model.Task) error {
	query := `UPDATE tasks
		SET title = $1, description = $2, status = $3, due_date = $4, updated_at = now()
		WHERE id = $5 AND user_id = $6
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query, t.ID, t.UserID).Scan(&t.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("update ask: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, userID, taskID string) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`

	tag, err := r.db.Exec(ctx, query, taskID, userID)

	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}
