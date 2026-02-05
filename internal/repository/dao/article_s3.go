package dao

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 是基于mysql数据库的
type ArticleS3DAO struct {
	ArticleGORMDAO
	oss *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, oss *s3.S3) ArticleDAO {
	return &ArticleS3DAO{
		ArticleGORMDAO: ArticleGORMDAO{db: db},
		oss:            oss,
	}
}

// 基本操作和ArticleGORMDAO一样，新增了一个同步到OSS的功能
// 函数内部操作和 GORMDAO一样，多了一个同步了文章的content到OSS

// 只有与线上库有关的两个操作函数涉及 oss
func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	// 开启db的事务
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)
		// db -> tx -> dao
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}

		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		// 现在内容不存在线上库了，content存在oss，所以有新的 publishedArticleV2 结构体
		pubArt := PulishedArticleV2{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Ctime:    now,
			Utime:    now,
			Status:   art.Status,
		}
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  pubArt.Title,
				"utime":  now,
				"status": pubArt.Status,
			}),
		}).Create(&pubArt).Error
		return err
	})
	if err != nil {
		return 0, err
	}
	// 同步content到oss
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-123"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}

// 给作者id，文章id和状态，更新两库的文章状态
func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? and author_id = ?", uid, id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PulishedArticleV2{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
	if err != nil {
		return err
	}

	// 只有设置的状态是私密的时候，删除oss上的内容
	const statusPrivate = 3
	if status == statusPrivate {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1314583317"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

// 没有 Content 的 Article
type PulishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
