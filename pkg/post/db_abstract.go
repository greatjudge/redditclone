package post

import (
	context "context"

	mongo "go.mongodb.org/mongo-driver/mongo"
)

// mockgen -source=post.go -destination=repo_mock.go -package=post PostRepo
type DatabaseHelper interface {
	Collection(name string) CollectionHelper
	Client() ClientHelper
}

type CollectionHelper interface {
	Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error)
	FindOne(context.Context, interface{}) SingleResultHelper
	InsertOne(context.Context, interface{}) (interface{}, error)
	DeleteOne(ctx context.Context, filter interface{}) (int64, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
}

type SingleResultHelper interface {
	Decode(v interface{}) error
}

type ClientHelper interface {
	Database(string) DatabaseHelper
	Connect() error
	StartSession() (mongo.Session, error)
}

type MongoClient struct {
	Cl *mongo.Client
}
type MongoDatabase struct {
	DB *mongo.Database
}
type MongoCollection struct {
	Coll *mongo.Collection
}

type MongoSingleResult struct {
	Sr *mongo.SingleResult
}

type MongoSession struct {
	mongo.Session
}

func (mc *MongoClient) Database(dbName string) DatabaseHelper {
	db := mc.Cl.Database(dbName)
	return &MongoDatabase{DB: db}
}

func (mc *MongoClient) StartSession() (mongo.Session, error) {
	session, err := mc.Cl.StartSession()
	return &MongoSession{session}, err
}

func (mc *MongoClient) Connect() error {
	return mc.Cl.Connect(context.TODO())
}

func (md *MongoDatabase) Collection(colName string) CollectionHelper {
	collection := md.DB.Collection(colName)
	return &MongoCollection{Coll: collection}
}

func (md *MongoDatabase) Client() ClientHelper {
	client := md.DB.Client()
	return &MongoClient{Cl: client}
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
	return mc.Coll.Find(ctx, filter)
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{}) SingleResultHelper {
	singleResult := mc.Coll.FindOne(ctx, filter)
	return &MongoSingleResult{Sr: singleResult}
}

func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	id, err := mc.Coll.InsertOne(ctx, document)
	return id.InsertedID, err
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter interface{}) (int64, error) {
	count, err := mc.Coll.DeleteOne(ctx, filter)
	return count.DeletedCount, err
}

func (mc *MongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return mc.Coll.UpdateOne(ctx, filter, update)
}

func (sr *MongoSingleResult) Decode(v interface{}) error {
	return sr.Sr.Decode(v)
}
