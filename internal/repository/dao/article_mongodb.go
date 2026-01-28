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

// 制作库插入一篇文章【雪花算法生成id】
func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, art)
	return art.Id, err
}

// 制作库更新文章，通过id
func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()

	filter := bson.D{bson.E{Key: "id", Value: art.Id}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
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

// 查询制作库文章列表（根据作者id分页）
func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	filter := bson.D{bson.E{Key: "author_id", Value: uid}}
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{bson.E{Key: "utime", Value: -1}}) // DESC
	res, err := m.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	// 一堆就 All
	err = res.All(ctx, &arts)
	return arts, err
}

// 查询制作库某篇文章（根据文章id）
func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	filter := bson.D{bson.E{Key: "id", Value: id}}
	res := m.col.FindOne(ctx, filter)
	// 单个就 Decode
	err := res.Decode(&art)
	return art, err
}

// 查询线上库某篇已发布文章（根据文章id）
func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pubArt PublishedArticle
	filter := bson.M{"id": id}
	res := m.liveCol.FindOne(ctx, filter)
	err := res.Decode(&pubArt)
	return pubArt, err
}

// 同步文章到线上库
// 【可能是制作库也没有的，都是insert|
// 可能是制作库已有制作库更新update，线上库insert|
// 可能是制作库和线上库都有的，都是update】
func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// 因为如果是新插入的话 Insert 内部雪花算法产生了 id
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now

	filter := bson.D{bson.E{Key: "id", Value: art.Id}, bson.E{Key: "author_id", Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set", Value: art}, bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	_, err = m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	return id, err
}

// 同步文章状态到线上库和制作库【制作和线上都是 update】
func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	// 双 id 锁定某篇具体文章【配所有权】！
	filter := bson.D{bson.E{Key: "id", Value: id}, bson.E{Key: "author_id", Value: uid}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"status": status,
	}}}

	// 制作库更新
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID 不对或者创作者不对")
	}
	// 线上库更新，并且不用check
	_, err = m.liveCol.UpdateOne(ctx, filter, set)
	return err
}
