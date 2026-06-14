package service

import (
	"context"
	"strings"

	"github.com/MonicaMell/task-api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type AuthService struct {
	users UserRepository
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{users: users}
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
