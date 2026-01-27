package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	// 严格来说，这个不是优秀实践
	return db.AutoMigrate(&User{})
}

// 以前这里只用 gorm 初始化了 mysql 的表
// 现在加一个初始化 mongodb 的集合

func InitCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 制作库集合
	col := mdb.Collection("articles")
	// 创建索引
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"id", 1}},
			Options: options.Index().SetUnique(true).SetName("idx_id")},
		{
			Keys: bson.D{{"author_id", 1}},
		},
	})
	if err != nil {
		return err
	}

	// 线上库集合和索引
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{"id", 1}},
			Options: options.Index().SetUnique(true).SetName("idx_id"),
		},
		{
			Keys: bson.D{{"author_id", 1}},
		},
	})
	return err
}
