package mongodb

import (
	"context"
	"fmt"
	"time"

	"JustChat/internal/messages/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepo interface {
	GetByID(ctx context.Context, id string) (*model.Message, error)
	GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	DeleteByID(ctx context.Context, id string) error
}

type messageRepo struct {
	db *mongo.Database
}

func NewMessageRepo(db *mongo.Database) MessageRepo {
	return &messageRepo{db: db}
}

// MessageRepo Methods

func (m *messageRepo) GetByID(ctx context.Context, id string) (*model.Message, error) {
	collection := m.db.Collection("messages")
	filter := bson.M{"_id": id, "deleted": false}
	var message model.Message
	err := collection.FindOne(ctx, filter).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID message repo: %w", err)
	}
	return &message, nil
}

func (m *messageRepo) GetByChatID(ctx context.Context, chatID int64) ([]model.Message, error) {
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

func (m *messageRepo) SaveMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	collection := m.db.Collection("messages")
	message.SentAt = time.Now()
	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("SaveMessage message repo: %w", err)
	}
	message.ID = result.InsertedID.(int64)
	return message, nil
}

func (m *messageRepo) DeleteByID(ctx context.Context, id string) error {
	collection := m.db.Collection("messages")
	filter := bson.M{"_id": id, "deleted": false}
	update := bson.M{"$set": bson.M{"deleted": true}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("DeleteByID message repo: %w", err)
	}
	return nil
}
