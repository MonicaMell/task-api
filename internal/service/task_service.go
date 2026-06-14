package service

import (
	"context"
	"time"

	"github.com/MonicaMell/task-api/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, userID, taskID string) (*model.Task, error)
	ListByUser(ctx context.Context, userID string) ([]model.Task, error)
	Update(ctx context.Context, t *model.Task) error
	Delete(ctx context.Context, userID, taskID string) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

type CreateTaskInput struct {
	Title       string     `json:"title" validate:"required,max=200"`
	Description string     `json:"description" validate:"max=2000"`
	Status      string     `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	DueDate     *time.Time `json:"due_date"`
}

func (s *TaskService) Create(ctx context.Context, userID string, in CreateTaskInput) (*model.Task, error) {
	status := in.Status
	if status == "" {
		status = "todo"
	}
	t := &model.Task{
		UserID:      userID,
		Title:       in.Title,
		Description: in.Description,
		Status:      status,
		DueDate:     in.DueDate,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Get(ctx context.Context, userID, taskID string) (*model.Task, error) {
	return s.repo.GetByID(ctx, userID, taskID)
}

func (s *TaskService) List(ctx context.Context, userID string) ([]model.Task, error) {
	return s.repo.ListByUser(ctx, userID)
}

type UpdateTaskInput struct {
	Title       string     `json:"title" validate:"required,max=200"`
	Description string     `json:"description" validate:"max=2000"`
	Status      string     `json:"status" validate:"required,oneof=todo in_progress done"`
	DueDate     *time.Time `json:"due_date"`
}

func (s *TaskService) Update(ctx context.Context, userID, taskID string, in UpdateTaskInput) (*model.Task, error) {
	t := &model.Task{
		ID:          taskID,
		UserID:      userID,
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		DueDate:     in.DueDate,
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	return s.repo.Delete(ctx, userID, taskID)
}
