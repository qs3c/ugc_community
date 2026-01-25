package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("Command started: %s\n", evt.CommandName)
		},
	}

	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)

	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	col := client.Database("webook").Collection("articles")

	// 增
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "mytitle",
		Content:  "mycontent",
		AuthorId: 134,
	})
	assert.NoError(t, err)

	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("insert id:", oid)

	// 查
	filter := bson.M{
		"id": 1,
	}
	findRes := col.FindOne(ctx, filter)
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var article Article
		err = findRes.Decode(&article)
		assert.NoError(t, err)
		t.Log("找到数据:", article)
	}

	// 改
	updateFilter := bson.D{bson.E{Key: "id", Value: 1}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{"title": "新标题"}}}

	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	updateManyRes, err := col.UpdateMany(ctx, updateFilter, bson.D{bson.E{Key: "$set", Value: Article{Content: "新内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	// 删
	deleteFilter := bson.D{bson.E{"id", 1}}
	deleteRes, err := col.DeleteMany(ctx, deleteFilter)
	assert.NoError(t, err)
	t.Log("删除数据:", deleteRes)

	
}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
