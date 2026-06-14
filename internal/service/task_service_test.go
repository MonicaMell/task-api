package service

import (
	"context"
	"testing"

	"github.com/MonicaMell/task-api/internal/model"
	"github.com/stretchr/testify/require"
)

type fakeTaskRepo struct {
	created *model.Task
}

func (f *fakeTaskRepo) Create(ctx context.Context, t *model.Task) error {
	t.ID = "fake-id"
	f.created = t
	return nil
}
func (f *fakeTaskRepo) GetByID(ctx context.Context, userID, taskID string) (*model.Task, error) {
	return nil, nil
}
func (f *fakeTaskRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.Task, error) {
	return nil, nil
}
func (f *fakeTaskRepo) Update(ctx context.Context, t *model.Task) error         { return nil }
func (f *fakeTaskRepo) Delete(ctx context.Context, userID, taskID string) error { return nil }

func TestCreateDefaultsStatusToTodo(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := NewTaskService(repo)

	task, err := svc.Create(context.Background(), "user-1", CreateTaskInput{Title: "no status given"})

	require.NoError(t, err)
	require.Equal(t, "todo", task.Status)
	require.Equal(t, "user-1", repo.created.UserID)
}
