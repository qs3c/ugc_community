package dao

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBArticleDAO struct {
	node *snowflake.Node
	// 不是用 mongo的Database组成的而是Collection构成的
	col     *mongo.Collection
	liveCol *mongo.Collection
}

// 好吧其实传入的就是 mongo.Database

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) *MongoDBArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
	}
}

// MongoDBArticleDAO 要和 ArticleGROMDAO 一样，实现 ArticleDAO 接口的各个方法

var _ ArticleDAO = (*MongoDBArticleDAO)(nil)

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, art)
	return art.Id, err
}
func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()

	filter := bson.D{bson.E{Key: "id", Value: art.Id}}
	set := bson.D{bson.E{Key: "$set",Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}
func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx,art)
	} else{
		id,err = m.Insert(ctx,art)
	}
	if err != nil {
		return 0, err
	}

	// 因为如果是新插入的话 Insert 内部雪花算法产生了 id
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now

	filter := bson.D{bson.E{Key: "id",Value: art.Id},bson.E{Key: "author_id",Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set",Value: art},bson.E{Key: "$setOnInsert",Value: bson.D{bson.E{Key: "ctime",Value: now}}}}
	
	_,err = m.liveCol.UpdateOne(ctx,filter,set,options.Update().SetUpsert(true))
	return id, err
}
func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	panic("XXX")
}
func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	panic("XXX")
}
func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	panic("XXX")
}
