package service

import (
	"context"
	"errors"
	"strings"

	"github.com/MonicaMell/task-api/internal/auth"
	"github.com/MonicaMell/task-api/internal/model"
	"github.com/MonicaMell/task-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type AuthService struct {
	users  UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

var ErrInvalidCredentials = errors.New("invalid email or password")

type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
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

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (string, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.tokens.Generate(user.ID)
}
