package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

type ErrorDuplicateKey struct {
}

func (e *ErrorDuplicateKey) Error() string {
	return "ErrorDuplicateKey"
}

type HxMongo struct {
	client *mongo.Client
}

// ConnectMongoDb https://docs.mongodb.com/manual/reference/connection-string/
func ConnectMongoDb(mongoURI string, maxPoolSize uint64) (*HxMongo, error) {
	// 启用优先读取从节点的配置
	opt := options.Client()
	if rpf, err := readpref.New(readpref.SecondaryPreferredMode); err != nil {
		return nil, err
	} else {
		opt.ReadPreference = rpf
	}
	client, err := mongo.Connect(context.Background(),
		opt.ApplyURI(mongoURI).SetMaxPoolSize(maxPoolSize))

	if err != nil {
		log.Printf("connect mongodb fail!")
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		log.Printf("ping mongo fail err(%v)/n", err)
		return nil, err
	}
	return &HxMongo{client: client}, nil
}

func (cli *HxMongo) GetDatabase(dbName string) *mongo.Database {
	return cli.client.Database(dbName)
}

func (cli *HxMongo) GetCollection(dbName, collName string) *mongo.Collection {
	return cli.client.Database(dbName).Collection(collName)
}

func (cli *HxMongo) CreateCollection(dbName, collName string) error {
	return cli.client.Database(dbName).CreateCollection(context.Background(), collName)
}

func (cli *HxMongo) CreateCollectionIndex(dbName, collName string, indexModels []mongo.IndexModel) ([]string, error) {
	return cli.client.Database(dbName).Collection(collName).Indexes().CreateMany(context.Background(), indexModels)
}

func (cli *HxMongo) GetCollectionNames(dbName string) ([]string, error) {
	return cli.client.Database(dbName).ListCollectionNames(context.Background(), bson.D{})
}

func (cli *HxMongo) HasCollection(dbName, collName string) (bool, error) {
	collList, err := cli.client.Database(dbName).ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return false, err
	}
	find := false
	for _, v := range collList {
		if collName == v {
			find = true
			break
		}
	}
	return find, nil
}

// InsertOne 这种需要创建唯一索引
func (cli *HxMongo) InsertOne(dbName, collName string, document interface{}) (interface{}, error) {
	result, err := cli.client.Database(dbName).Collection(collName).InsertOne(context.Background(), document)
	if err != nil {
		log.Printf("insert one err(%v)", err)
		we, ok := err.(mongo.WriteException)
		if ok && len(we.WriteErrors) == 1 {
			if we.WriteErrors[0].Code == 11000 { // 文档已存在
				return nil, &ErrorDuplicateKey{}
			}
		}
		return nil, err
	}
	return result.InsertedID, nil
}

func (cli *HxMongo) InsertMany(dbName string, collName string, documents []interface{}) ([]interface{}, error) {
	result, err := cli.client.Database(dbName).Collection(collName).InsertMany(context.Background(), documents)
	if err != nil {
		log.Printf("insert many err(%v)", err)
		we, ok := err.(mongo.WriteException)
		if ok && len(we.WriteErrors) == 1 {
			if we.WriteErrors[0].Code == 11000 { // 文档已存在
				return nil, &ErrorDuplicateKey{}
			}
		}
		return nil, err
	}
	return result.InsertedIDs, nil
}

func (cli *HxMongo) Update(dbName string, collName string, filter interface{}, update interface{}, bMany bool) (interface{}, int64, error) {
	var err error
	var result *mongo.UpdateResult
	collection := cli.client.Database(dbName).Collection(collName)
	if bMany {
		result, err = collection.UpdateMany(context.Background(), filter, update)
	} else {
		result, err = collection.UpdateOne(context.Background(), filter, update)
	}
	if err != nil {
		log.Print(err)
		return nil, 0, err
	}
	return result.UpsertedID, result.ModifiedCount + result.UpsertedCount, nil
}

func (cli *HxMongo) Replace(dbName string, collName string, filter interface{}, replacement interface{}) (interface{}, error) {
	result, err := cli.client.Database(dbName).Collection(collName).ReplaceOne(context.Background(), filter, replacement)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return result.UpsertedID, nil
}

func (cli *HxMongo) Delete(dbName string, collName string, filter interface{}, bMany bool) (int64, error) {
	var err error
	var result *mongo.DeleteResult
	collection := cli.client.Database(dbName).Collection(collName)
	if bMany {
		result, err = collection.DeleteMany(context.Background(), filter)
	} else {
		result, err = collection.DeleteOne(context.Background(), filter)
	}
	if err != nil {
		log.Print(err)
		return 0, err
	}
	return result.DeletedCount, nil
}

func (cli *HxMongo) Find(dbName string, collName string, filter interface{}, opts *options.FindOptions) (*mongo.Cursor, error) {
	var err error
	var cur *mongo.Cursor
	collection := cli.client.Database(dbName).Collection(collName)
	if opts != nil {
		cur, err = collection.Find(context.Background(), filter, opts)
	} else {
		cur, err = collection.Find(context.Background(), filter)
	}
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	if cur.Err() != nil {
		return nil, err
	}
	return cur, err
}

func (cli *HxMongo) FindOne(dbName string, collName string, filter interface{}, opts *options.FindOneOptions) (*mongo.SingleResult, error) {
	var rest *mongo.SingleResult
	collection := cli.client.Database(dbName).Collection(collName)
	if opts != nil {
		rest = collection.FindOne(context.Background(), filter, opts)
	} else {
		rest = collection.FindOne(context.Background(), filter)
	}
	if rest.Err() != nil {
		return nil, rest.Err()
	}
	return rest, nil
}

func (cli *HxMongo) Aggregate(dbName string, collName string, pipeline interface{}) (*mongo.Cursor, error) {
	var err error
	var cursor *mongo.Cursor
	if cursor, err = cli.client.Database(dbName).Collection(collName).Aggregate(context.Background(), pipeline); err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	if cursor.Err() != nil {
		return nil, err
	}
	return cursor, err
}

func (cli *HxMongo) Count(dbName string, collName string, filter interface{}, opts *options.CountOptions) (int64, error) {
	var err error
	var count int64
	collection := cli.client.Database(dbName).Collection(collName)
	if opts != nil {
		count, err = collection.CountDocuments(context.Background(), filter, opts)
	} else {
		count, err = collection.CountDocuments(context.Background(), filter)
	}
	if err != nil {
		log.Printf("%v", err)
		return 0, err
	}
	return count, err
}

func (cli *HxMongo) DeleteCollection(dbName, collName string) error {
	return cli.client.Database(dbName).Collection(collName).Drop(context.Background())
}

func (cli *HxMongo) Transaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return cli.client.UseSession(ctx, fn)
}
