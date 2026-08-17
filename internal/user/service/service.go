package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"family-shopping-list-api/internal/config"
	"family-shopping-list-api/internal/user/model"
	"family-shopping-list-api/internal/user/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUsernameTaken      = errors.New("用户名已存在")
)

type Service struct {
	repo *repository.Repository
	cfg  config.Config
}

func New(repo *repository.Repository, cfg config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Nickname = strings.TrimSpace(input.Nickname)
	if len(input.Username) < 3 || len(input.Password) < 8 {
		return nil, errors.New("用户名至少 3 位，密码至少 8 位")
	}
	if input.Nickname == "" {
		input.Nickname = input.Username
	}

	if _, err := s.repo.GetByUsername(ctx, input.Username); err == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     input.Username,
		Nickname:     input.Nickname,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	created, err := s.repo.GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: *created}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	user, err := s.repo.GetByUsername(ctx, strings.TrimSpace(input.Username))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *Service) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) issueToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", user.ID),
		"user_id":  user.ID,
		"username": user.Username,
		"iat":      now.Unix(),
		"exp":      now.Add(s.cfg.TokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
