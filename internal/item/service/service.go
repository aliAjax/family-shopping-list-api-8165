package service

import (
	"context"
	"errors"
	"strings"

	"family-shopping-list-api/internal/item/model"
	"family-shopping-list-api/internal/item/repository"
)

var ErrNotFound = repository.ErrNotFound

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type UpdateInput struct {
	Name      *string `json:"name"`
	Quantity  *int    `json:"quantity"`
	Purchased *bool   `json:"purchased"`
}

func (s *Service) Create(ctx context.Context, listID, userID uint64, input CreateInput) (*model.Item, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	item := &model.Item{
		ListID:    listID,
		Name:      input.Name,
		Quantity:  input.Quantity,
		Purchased: false,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, listID, item.ID)
}

func (s *Service) Get(ctx context.Context, listID, itemID uint64) (*model.Item, error) {
	return s.repo.GetByID(ctx, listID, itemID)
}

func (s *Service) List(ctx context.Context, listID uint64) ([]model.Item, error) {
	return s.repo.ListByList(ctx, listID)
}

func (s *Service) Update(ctx context.Context, listID, itemID, userID uint64, input UpdateInput) (*model.Item, error) {
	item, err := s.repo.GetByID(ctx, listID, itemID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, errors.New("商品名称不能为空")
		}
		item.Name = name
	}
	if input.Quantity != nil {
		if *input.Quantity <= 0 {
			return nil, errors.New("商品数量必须大于 0")
		}
		item.Quantity = *input.Quantity
	}
	if input.Purchased != nil {
		item.Purchased = *input.Purchased
	}
	item.UpdatedBy = userID
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, listID, itemID)
}

func (s *Service) Delete(ctx context.Context, listID, itemID uint64) error {
	return s.repo.Delete(ctx, listID, itemID)
}
