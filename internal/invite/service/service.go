package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"family-shopping-list-api/internal/invite/model"
	inviteRepo "family-shopping-list-api/internal/invite/repository"
	memberRepo "family-shopping-list-api/internal/member/repository"
)

var (
	ErrAlreadyMember = errors.New("你已经是该清单成员")
	ErrExpired       = errors.New("邀请码已过期")
	ErrLimitReached  = errors.New("邀请码使用次数已达上限")
	ErrInactive      = errors.New("邀请码已失效")
)

type Service struct {
	repo       *inviteRepo.Repository
	memberRepo *memberRepo.Repository
}

func New(repo *inviteRepo.Repository, memberRepo *memberRepo.Repository) *Service {
	return &Service{repo: repo, memberRepo: memberRepo}
}

type CreateInput struct {
	MaxUses   int    `json:"max_uses"`
	ExpiresAt string `json:"expires_at"`
}

type JoinInput struct {
	Code string `json:"code"`
}

func (s *Service) Create(ctx context.Context, userID, listID uint64, input CreateInput) (*model.Invite, error) {
	code, err := generateCode()
	if err != nil {
		return nil, err
	}
	if input.MaxUses <= 0 {
		input.MaxUses = 10
	}
	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			return nil, errors.New("expires_at 需使用 RFC3339 格式")
		}
		expiresAt = &parsed
	}
	invite := &model.Invite{
		ListID:    listID,
		Code:      code,
		CreatedBy: userID,
		MaxUses:   input.MaxUses,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, invite); err != nil {
		return nil, err
	}
	return s.repo.GetByCode(ctx, code)
}

func (s *Service) ListByList(ctx context.Context, listID uint64) ([]model.Invite, error) {
	return s.repo.ListByList(ctx, listID)
}

func (s *Service) Join(ctx context.Context, userID uint64, code string) (*model.Invite, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("邀请码不能为空")
	}
	invite, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if !invite.Active {
		return nil, ErrInactive
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpired
	}
	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return nil, ErrLimitReached
	}
	isMember, err := s.memberRepo.IsMember(ctx, invite.ListID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return invite, nil
	}
	if err := s.memberRepo.Add(ctx, invite.ListID, userID, "member"); err != nil {
		return nil, err
	}
	if err := s.repo.IncrementUsed(ctx, invite.ID); err != nil {
		return nil, err
	}
	return s.repo.GetByCode(ctx, code)
}

func generateCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const length = 8
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}
