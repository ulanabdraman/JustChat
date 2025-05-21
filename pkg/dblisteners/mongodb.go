package dblisteners

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"

	"JustChat/pkg/streamhub"
)

type MongoNotifyListener struct {
	ctx        context.Context
	collection *mongo.Collection
	output     chan<- streamhub.Event
	input      <-chan streamhub.Event
	hub        *streamhub.StreamHub
}

func NewMongoNotifyListener(ctx context.Context, collection *mongo.Collection, hub *streamhub.StreamHub) (*MongoNotifyListener, error) {
	listener := &MongoNotifyListener{
		ctx:        ctx,
		collection: collection,
		hub:        hub,
	}

	// Сразу подключаем каналы из hub
	listener.SetOutput(hub.SubscribeAsProducer())
	listener.SetInput(hub.Subscribe(100))

	return listener, nil
}

func (l *MongoNotifyListener) SetOutput(ch chan<- streamhub.Event) {
	l.output = ch
}

func (l *MongoNotifyListener) SetInput(ch <-chan streamhub.Event) {
	l.input = ch
}

func (l *MongoNotifyListener) Start() error {
	pipeline := mongo.Pipeline{
		{{"$match", map[string]interface{}{
			"operationType": "insert",
		}}},
	}

	cs, err := l.collection.Watch(l.ctx, pipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		return err
	}

	go func() {
		defer cs.Close(l.ctx)

		for cs.Next(l.ctx) {
			var event map[string]interface{}
			if err := cs.Decode(&event); err != nil {
				log.Println("Failed to decode change event:", err)
				continue
			}

			// Логируем событие красиво отформатированным JSON
			eventJSON, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				log.Println("Failed to marshal event for logging:", err)
			} else {
				log.Println("MongoDB change event:\n" + string(eventJSON))
			}

			// Отправляем событие в output канал
			if l.output != nil {
				select {
				case l.output <- event:
				case <-l.ctx.Done():
					return
				}
			}
		}

		if err := cs.Err(); err != nil {
			log.Println("Change stream error:", err)
		}
	}()

	// Обработка входящих событий (если нужна)
	go func() {
		for {
			select {
			case <-l.ctx.Done():
				return
			case ev := <-l.input:
				// Логика обработки входящих событий, например, логирование
				log.Println("Received event from input channel:", ev)
			}
		}
	}()

	return nil
}
