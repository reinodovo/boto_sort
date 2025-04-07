package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	client *mongo.Client
}

const (
	databaseName = "boto_sort"
)

func NewMongoDatabase() *MongoDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Panic(err)
	}
	return &MongoDatabase{client: client}
}

func (db *MongoDatabase) GetObject(collectionName, key string, object any) error {
	collection := db.client.Database(databaseName).Collection(collectionName)

	var result bson.M

	err := collection.FindOne(context.Background(), bson.D{{Key: "_id", Value: key}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return ErrKeyNotFound
	} else if err != nil {
		return err
	}
	if value, ok := result["value"]; ok {
		bsonType, data, err := bson.MarshalValue(value)
		if err != nil {
			return err
		}
		rawData := bson.RawValue{Type: bsonType, Value: data}
		err = rawData.Unmarshal(object)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MongoDatabase) SaveObject(collectionName, key string, object any) error {
	collection := db.client.Database(databaseName).Collection(collectionName)

	result := struct {
		Value interface{}
	}{
		Value: object,
	}

	_, err := collection.ReplaceOne(context.Background(), bson.D{{Key: "_id", Value: key}}, result, options.Replace().SetUpsert(true))
	return err
}

func (db *MongoDatabase) Close() error {
	return db.client.Disconnect(context.Background())
}
