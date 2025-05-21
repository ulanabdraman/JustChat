package dblisteners

import (
	"JustChat/pkg/streamhub"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"log"
)

type PostgresNotifyListener struct {
	ctx     context.Context
	db      *pgx.Conn
	channel string
	output  chan<- streamhub.Event
	input   <-chan streamhub.Event
	hub     *streamhub.StreamHub
}

func NewPostgresNotifyListener(ctx context.Context, db *pgx.Conn, channel string, hub *streamhub.StreamHub) (*PostgresNotifyListener, error) {
	listener := &PostgresNotifyListener{
		ctx:     ctx,
		db:      db,
		channel: channel,
		hub:     hub,
	}

	listener.SetOutput(hub.SubscribeAsProducer())
	listener.SetInput(hub.Subscribe(100))

	return listener, nil
}

func (l *PostgresNotifyListener) SetOutput(ch chan<- streamhub.Event) {
	l.output = ch
}

func (l *PostgresNotifyListener) SetInput(ch <-chan streamhub.Event) {
	l.input = ch
}

func (l *PostgresNotifyListener) Start() error {
	_, err := l.db.Exec(l.ctx, fmt.Sprintf("LISTEN %s;", l.channel))
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-l.ctx.Done():
				return
			default:
				notification, err := l.db.WaitForNotification(l.ctx)
				if err != nil {
					log.Println("WaitForNotification error:", err)
					continue
				}

				l.output <- notification.Payload
			}
		}
	}()

	go func() {
		for {
			select {
			case <-l.ctx.Done():
				return
			case ev := <-l.input:
				// Здесь логика обработки входящих событий (например, лог, или что-то ещё)
				log.Println("Received event from input channel:", ev)
			}
		}
	}()

	return nil
}
