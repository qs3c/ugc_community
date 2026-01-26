package mongodb

import (
	"context"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func (suite *MongoDBTestSuite) SetupSuite() {
	t := suite.T()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			t.Logf("Command started: %s\n", evt.CommandName)
		},
	}

	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)

	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	col := client.Database("webook").Collection("articles")
	suite.col = col

	manyRes, err := col.InsertMany(ctx, []any{Article{
		Id:       1,
		AuthorId: 11,
	}, Article{
		Id:       2,
		AuthorId: 22,
	}})

	assert.NoError(t, err)
	t.Log("插入数量", len(manyRes.InsertedIDs))
}

func (suite *MongoDBTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// _, err := suite.col.DeleteMany(ctx, map[string]any{})
	_, err := suite.col.DeleteMany(ctx, bson.D{})
	assert.NoError(suite.T(), err)

	_, err = suite.col.Indexes().DropAll(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *MongoDBTestSuite) TestOr() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.A{bson.D{bson.E{Key: "id", Value: 1}}, bson.D{bson.E{Key: "id", Value: 2}}}

	res, err := suite.col.Find(ctx, bson.D{{Key: "$or", Value: filter}})
	assert.NoError(suite.T(), err)

	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(suite.T(), err)
	suite.T().Log("查询结果:", arts)
}

func (suite *MongoDBTestSuite) TestAnd() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.A{bson.D{bson.E{Key: "id", Value: 1}}, bson.D{bson.E{Key: "author_id", Value: 11}}}

	res, err := suite.col.Find(ctx, bson.D{{Key: "$and", Value: filter}})
	assert.NoError(suite.T(), err)

	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(suite.T(), err)
	suite.T().Log("查询结果:", arts)
}

func (suite *MongoDBTestSuite) TestIn() {
	ctx, cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	filter := bson.D{{Key:"id",Value:bson.D{bson.E{Key: "$in",Value: []int{1,2}}}}}

	proj := bson.M{"id":1}
	// proj := bson.M{"id":0}
	
	res,err:=suite.col.Find(ctx,filter,options.Find().SetProjection(proj))
	assert.NoError(suite.T(),err)

	var arts []Article
	err = res.All(ctx,&arts)
	assert.NoError(suite.T(),err)
	suite.T().Log("查询结果:",arts)
}

// 测试索引
func (suite *MongoDBTestSuite) TestIndexes(){
	ctx, cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	ires,err:=suite.col.Indexes().CreateOne(ctx,mongo.IndexModel{
		Keys: bson.D{{Key:"id",Value:1}},
		Options: options.Index().SetUnique(true).SetName("idx_id"),
	})
	assert.NoError(suite.T(),err)
	suite.T().Log("创建索引:",ires)
}
