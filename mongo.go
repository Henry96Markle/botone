package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Record struct {
	ChatID int64     `bson:"chat_id" json:"chat_id"`
	Notes  []string  `bson:"notes" json:"notes"`
	Date   time.Time `bson:"date" json:"date"`
}

type User struct {
	ID         primitive.ObjectID    `bson:"_id" json:"_id"`
	Names      []string              `bson:"names" json:"names"`
	Usernames  []string              `bson:"usernames" json:"usernames"`
	TelegramID int64                 `bson:"tg_id" json:"tg_id"`
	AliasIDs   []int64               `bson:"alias_ids" json:"alias_ids"`
	Permission int                   `bson:"permission_level" json:"permission_level"`
	Records    map[string]([]Record) `bson:"records" json:"records"`
}

type Database struct {
	client     *mongo.Client
	database   *mongo.Database
	collection string
}

func NewDatabase(connectionString string, databaseName string, collectionName string) (database *Database, err error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionString))

	if err != nil {
		return nil, err
	}

	return &Database{
		client:     client,
		database:   client.Database(databaseName),
		collection: collectionName,
	}, nil
}

func (d Database) Disconnect() error {
	return d.client.Disconnect(context.TODO())
}

func (d Database) Collection() *mongo.Collection {
	return d.database.Collection(d.collection)
}

func (d Database) Find(filter bson.D) (user User, err error) {
	user = User{}

	err = d.Collection().FindOne(context.TODO(), filter).Decode(&user)

	return
}

func (d Database) Filter(filter bson.D) (users []User, err error) {
	collection, err := d.Collection().Find(context.TODO(), filter)

	if err != nil {
		return nil, err
	}

	users = make([]User, 0)

	r := true

	for _ = collection.Current; r; r = collection.Next(context.TODO()) {
		record := User{}

		decode_err := collection.Decode(&record)

		if decode_err == nil {
			users = append(users, record)
		}
	}

	return
}

func (d Database) GetRecords() (records []User, err error) {
	return d.Filter(bson.D{{}})
}

// Looks up using only tg_id field.
func (d Database) FindByID(id int64) (user User, err error) {
	user = User{}

	err = d.Collection().FindOne(context.TODO(), bson.D{{Key: "tg_id", Value: id}}).Decode(&user)

	return
}

func (d Database) Add(users ...any) error {
	var err error = nil

	if len(users) == 1 {
		_, err = d.Collection().InsertOne(context.TODO(), users[0])
	} else {
		_, err = d.Collection().InsertMany(context.TODO(), users)
	}

	return err
}

func (d Database) RemoveByID(id int64) error {
	_, err := d.Collection().DeleteOne(context.TODO(), bson.D{{Key: "tg_id", Value: id}})

	return err
}

func (d Database) Aliases(pull bool, id int64, ids ...int64) error {
	modifier := "$push"

	if pull {
		modifier = "$pull"
	}

	operator := "$each"

	if pull {
		operator = "$in"
	}

	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{{
			Key: "tg_id", Value: id,
		}},
		bson.D{{
			Key: modifier,
			Value: bson.D{{
				Key: "alias_ids",
				Value: bson.D{{
					Key: operator, Value: ids,
				}},
			}},
		}},
	)

	return err
}

func (d Database) Record(id int64, category string, record Record, remove bool) error {

	modifier := "$push"

	if remove {
		modifier = "$pull"
	}

	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{{Key: "tg_id", Value: id}}, bson.D{{Key: modifier, Value: bson.D{{Key: "records." + category, Value: record}}}},
	)

	if err != nil {
		return fmt.Errorf("error accessing records from ID \"%d\": %w", id, err)
	} else {
		return nil
	}
}

func (d Database) Names(pull bool, id int64, names ...string) error {
	modifier := "$push"

	if pull {
		modifier = "$pull"
	}

	operator := "$each"

	if pull {
		operator = "$in"
	}

	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{{Key: "tg_id", Value: id}},
		bson.D{{Key: modifier, Value: bson.D{{Key: "names", Value: bson.D{{Key: operator, Value: names}}}}}},
	)

	return err
}

func (d Database) Usernames(pull bool, id int64, usernames ...string) error {
	modifier := "$push"

	if pull {
		modifier = "$pull"
	}

	operator := "$each"

	if pull {
		operator = "$in"
	}

	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{
			{Key: "tg_id", Value: id}},
		bson.D{{Key: modifier, Value: bson.D{{Key: "usernames", Value: bson.D{{Key: operator, Value: usernames}}}}}})
	return err
}

func (d Database) ReplaceByID(id int64, user User) error {
	_, err := d.Collection().ReplaceOne(context.TODO(), bson.D{{Key: "tg_id", Value: id}}, user)

	return err
}

// Generic API

func (d Database) Field(id int64, field string, value string) error {
	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{{Key: "tg_id", Value: id}}, bson.D{{Key: "$set", Value: bson.D{{Key: field, Value: value}}}})

	return err
}

func (d Database) Array(pull bool, id int64, field string, values ...any) error {
	modifier := "$push"

	if pull {
		modifier = "$pull"
	}

	_, err := d.Collection().UpdateOne(
		context.TODO(),
		bson.D{
			{Key: "tg_id", Value: id}},
		bson.D{{
			Key: modifier,
			Value: bson.D{{
				Key: field,
				Value: bson.D{{
					Key:   "$each",
					Value: values,
				}},
			}},
		}},
	)

	return err
}
