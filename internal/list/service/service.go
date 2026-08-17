package service

import (
	"context"
	"errors"
	"strings"

	"family-shopping-list-api/internal/list/model"
	"family-shopping-list-api/internal/list/repository"
	memberRepository "family-shopping-list-api/internal/member/repository"
)

var ErrNotFound = repository.ErrNotFound

type Service struct {
	repo       *repository.Repository
	memberRepo *memberRepository.Repository
}

func New(repo *repository.Repository, memberRepo *memberRepository.Repository) *Service {
	return &Service{repo: repo, memberRepo: memberRepo}
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Service) Create(ctx context.Context, userID uint64, input CreateInput) (*model.ShoppingList, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("清单名称不能为空")
	}
	list := &model.ShoppingList{
		Name:        input.Name,
		Description: strings.TrimSpace(input.Description),
		OwnerID:     userID,
	}
	if err := s.repo.Create(ctx, list); err != nil {
		return nil, err
	}
	if err := s.memberRepo.Add(ctx, list.ID, userID, "owner"); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, list.ID)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (*model.ShoppingList, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID uint64) ([]model.ShoppingList, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id uint64, input UpdateInput) (*model.ShoppingList, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("清单名称不能为空")
	}
	list := &model.ShoppingList{
		ID:          id,
		Name:        input.Name,
		Description: strings.TrimSpace(input.Description),
	}
	if err := s.repo.Update(ctx, list); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
