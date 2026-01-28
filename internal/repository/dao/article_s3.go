package dao

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
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

func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {

}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {

}
