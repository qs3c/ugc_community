package repository

import (
	"context"

	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1
}

// AddCollectionItem implements [InteractiveRepository].
func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	panic("unimplemented")
}

// Collected implements [InteractiveRepository].
func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	panic("unimplemented")
}

// DecrLike implements [InteractiveRepository].
func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	panic("unimplemented")
}

// Get implements [InteractiveRepository].
func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	panic("unimplemented")
}

// IncrLike implements [InteractiveRepository].
func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	panic("unimplemented")
}

// IncrReadCnt implements [InteractiveRepository].
func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		c.l.Error("增加阅读数失败", logger.Field{Key: "error", Val: err})
		return err
	}
	err = c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	if err != nil {
		c.l.Error("缓存增加阅读数失败", logger.Field{Key: "error", Val: err})
	}
	return nil
}

// Liked implements [InteractiveRepository].
func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	panic("unimplemented")
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO,
	l logger.LoggerV1,
	cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache, l: l}
}
