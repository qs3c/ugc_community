package service

import (
	"context"

	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
)

type InteractiveService interface {
	IncrReadCont(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error

	Get(ctx context.Context, biz string, id, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

// CancelLike implements [InteractiveService].
func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	panic("unimplemented")
}

// Collect implements [InteractiveService].
func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	panic("unimplemented")
}

// Get implements [InteractiveService].
func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	panic("unimplemented")
}

// IncrReadCont implements [InteractiveService].
func (i *interactiveService) IncrReadCont(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx,biz,bizId)
}

// Like implements [InteractiveService].
func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	panic("unimplemented")
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}
