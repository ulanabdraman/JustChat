package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"JustChat/internal/messages/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type messageRepoMongoDB struct {
	db *mongo.Database
}

func NewMessageRepoMongoDB(db *mongo.Database) MessageRepo {
	return &messageRepoMongoDB{db: db}
}

// MessageRepo Methods

func (m *messageRepoMongoDB) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	collection := m.db.Collection("messages")
	filter := bson.M{"_id": id, "deleted": false}
	var message model.Message
	err := collection.FindOne(ctx, filter).Decode(&message)
	if err != nil {
		return nil, fmt.Errorf("GetByID message repo: %w", err)
	}
	return &message, nil
}

func (m *messageRepoMongoDB) GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error) {
	collection := m.db.Collection("messages")
	filter := bson.M{"chat_id": chatID, "deleted": false}
	var messages []model.Message

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("GetByChatID message repo: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var message model.Message
		if err := cursor.Decode(&message); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return messages, nil
}

func (m *messageRepoMongoDB) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	collection := m.db.Collection("messages")

	id, err := m.getNextSequence(ctx, m.db, "messages")
	if err != nil {
		return nil, fmt.Errorf("generate ID: %w", err)
	}
	message.ID = id
	message.SentAt = time.Now()
	_, err = collection.InsertOne(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("SaveMessage message repo: %w", err)
	}
	return message, nil
}

func (m *messageRepoMongoDB) DeleteByID(ctx context.Context, id int64) error {
	collection := m.db.Collection("messages")
	filter := bson.M{"_id": id, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("DeleteByID message repo: %w", err)
	}
	return nil
}
func (m *messageRepoMongoDB) getNextSequence(ctx context.Context, db *mongo.Database, name string) (int64, error) {
	var result struct {
		Seq int64 `bson:"seq"`
	}
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).                 // создаёт документ, если не существует
		SetReturnDocument(options.After) // вернёт новое значение seq

	err := db.Collection("counters").FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("getNextSequence: %w", err)
	}
	return result.Seq, nil
}
