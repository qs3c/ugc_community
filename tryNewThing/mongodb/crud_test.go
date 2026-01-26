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

	// 监控 mongoDB 执行的命令
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("Command started: %s\n", evt.CommandName)
		},
	}

	// 配置 mongoDB 客户端的连接地址，并挂载监控器
	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)

	// 建立连接得到 mongoDB 客户端
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	// 选择数据库和集合
	// 获取到 webook 数据库下的 articles 集合
	col := client.Database("webook").Collection("articles")

	// 增（如果数据库或集合不存在，MongoDB 通常会在第一次写入数据时自动创建）
	// database => collection => document => field
	// 自动序列化成 bson 数据存入 MongoDB
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "mytitle",
		Content:  "mycontent",
		AuthorId: 134,
	})
	assert.NoError(t, err)

	// mongodb 中有自己的 12 字节的一个 objectID
	// 没有 mysql 那种自增 id，而是随机的 12 字节对象 ID
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
		// bson 数据反序列化
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

	// 所以第三个参数是指令文档，指令文档中只有一个指令就是 set，set 的内容是对应 article 文档
	updateManyRes, err := col.UpdateMany(ctx, updateFilter, bson.D{bson.E{Key: "$set", Value: Article{Content: "新内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	// 删
	deleteFilter := bson.D{bson.E{Key: "id", Value: 1}}
	// bson.D{bson.E{"id", 1}}
	// bson.D{{"id", 1}}
	// bson.M{"id": 1,}
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
