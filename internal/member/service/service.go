package service

import (
	"context"
	"errors"

	"family-shopping-list-api/internal/member/model"
	memberRepo "family-shopping-list-api/internal/member/repository"
)

var ErrCannotRemoveOwner = errors.New("不能移除清单所有者")

type Service struct {
	repo *memberRepo.Repository
}

func New(repo *memberRepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, listID uint64) ([]model.Member, error) {
	return s.repo.List(ctx, listID)
}

func (s *Service) Remove(ctx context.Context, listID, memberID uint64) error {
	member, err := s.repo.GetByID(ctx, listID, memberID)
	if err != nil {
		return err
	}
	if member.Role == "owner" {
		return ErrCannotRemoveOwner
	}
	return s.repo.Remove(ctx, listID, memberID)
}

func (s *Service) IsMember(ctx context.Context, listID, userID uint64) (bool, error) {
	return s.repo.IsMember(ctx, listID, userID)
}

func (s *Service) IsOwner(ctx context.Context, listID, userID uint64) (bool, error) {
	return s.repo.IsOwner(ctx, listID, userID)
}
