package db

import (
	"context"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	dbtimeOut, _          = strconv.Atoi(os.Getenv("DB_CLIENT_TIMEOUT"))
	createIndexMaxTime, _ = strconv.Atoi(os.Getenv("CREATE_INDEX_MAX_TIME"))
)

// New creates client for talking to the mongodb
// NOTE: to use it in production, you most likely need to add the preffered way of
//       authentication, see https://godoc.org/go.mongodb.org/mongo-driver/mongo#Connect
func New(config Config) (*MongoDatastore, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(config.DbLocation))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	database := client.Database(config.DbName)
	mongodb := &MongoDatastore{
		Config:  config,
		db:      database,
		Session: client,
	}
	return mongodb, nil
}

// CheckCollection just check if the requested collection already exists in the database
func (mongodb *MongoDatastore) CheckCollection(collName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	names, err := mongodb.db.ListCollectionNames(ctx, bson.D{{}})
	if err != nil {
		return false, err
	}
	for _, name := range names {
		if name == collName {
			return true, nil
		}
	}
	return false, nil
}

// CreateCollection checks if the helper collection exists
// in the db, and creates them if needed
func (mongodb *MongoDatastore) CreateCollection(collName string) (MongoCollection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	err := mongodb.db.CreateCollection(ctx, collName, nil)
	if err != nil {
		return MongoCollection{}, err
	}
	coll := mongodb.GetCollection(collName)
	return coll, nil
}

// DropCollection drops collection on the server
func (mongodb *MongoDatastore) DropCollection(collectionName string) error {
	coll := mongodb.db.Collection(collectionName)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	// defer cancel()
	err := coll.Drop(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// GetCollection just wraps the default mongo collection into custom one
func (mongodb *MongoDatastore) GetCollection(collName string) MongoCollection {
	return MongoCollection{mongodb.db.Collection(collName)}
}

// Disconnect client from the context
func (mongodb *MongoDatastore) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	mongodb.Session.Disconnect(ctx)
}

// GetCollSize returns number of documents in the requested collection
func (mongodb *MongoDatastore) GetCollSize(collName string) (int64, error) {
	opts := options.EstimatedDocumentCount().SetMaxTime(5 * time.Second)
	coll := mongodb.GetCollection(collName)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	count, err := coll.EstimatedDocumentCount(ctx, opts)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreateIndexesByFields just creates the new unique ascending
// indexes based on field name (type should be int)
func (coll MongoCollection) CreateIndexesByFields(fields []string, unique bool) error {
	models := make([]mongo.IndexModel, len(fields))
	for i, field := range fields {
		models[i] = mongo.IndexModel{
			Keys: bson.D{{field, 1}},
			Options: options.MergeIndexOptions(
				options.Index().SetBackground(true), // deprecated since mongodb 4.2
				options.Index().SetUnique(unique),
				options.Index().SetSparse(true),
			),
		}
	}
	opts := options.CreateIndexes().SetMaxTime(time.Duration(createIndexMaxTime) * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := coll.Indexes().CreateMany(ctx, models, opts)
	if err != nil {
		return err
	}
	return nil
}

// DropIndexByField sends command to drop the selected index;
// Input format should be in the following format: ""Some Field_1""
func (coll MongoCollection) DropIndexByField(indexName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(createIndexMaxTime)*time.Second)
	defer cancel()
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return err
	}
	return nil
}

// GetAggregation runs prepared aggregation pipeline in mongodb
func (coll MongoCollection) GetAggregation(groupStage mongo.Pipeline) ([]bson.M, error) {
	opts := options.Aggregate().SetMaxTime(time.Duration(dbtimeOut) * time.Second)
	cursor, err := coll.Aggregate(context.TODO(), groupStage, opts)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateField updates the selected field of the doc.
// Example:
//     filter := bson.D{{"_id", id}}
//     update := bson.D{{"$set", bson.D{{"email", "newemail@example.com"}}}}
func (coll MongoCollection) UpdateField(filter, update bson.D) error {
	opts := options.Update().SetUpsert(true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

// SetRecords adds the new documents to the collection
// docs := []interface{}{
//     bson.D{{"name", "Alice"}},
//     bson.D{{"name", "Bob"}},
// }
func (coll MongoCollection) SetRecords(data []interface{}) error {
	opts := options.InsertMany().SetOrdered(false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	_, err := coll.InsertMany(ctx, data, opts)
	if err != nil {
		return err
	}
	return nil
}

// DeleteRecords deletes records from specified collection by query
func (coll MongoCollection) DeleteRecords(query bson.D) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	_, err := coll.DeleteMany(ctx, query, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetCursor returns db cursor for specified collection and query
// Example queries:
//     bson.D{{"secondaryId", bson.M{"$in": []int{1, 3}}}}
// 	   bson.D{{"Hasher", bson.D{{"$exists", true}}}}
func (coll MongoCollection) GetCursor(query FindQuery) (*mongo.Cursor, error) {
	opts := options.MergeFindOptions(
		options.Find().SetLimit(int64(query.Limit)),
		options.Find().SetProjection(query.Proj),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbtimeOut)*time.Second)
	defer cancel()
	cursor, err := coll.Find(ctx, query.Query, opts)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}
