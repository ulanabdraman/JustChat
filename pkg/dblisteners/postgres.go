package dblisteners

import (
	"JustChat/pkg/streamhub"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresLogListener struct {
	db       *sqlx.DB
	listener *pq.Listener
	hubChan  streamhub.Event
}

func NewPostgresLogListener(db *sqlx.DB) (*PostgresLogListener, error) {
	// Получаем строку подключения обратно из sqlx.DB
	connStr := db.DriverName() // Это не строка подключения, нужно вручную

	// Важно: connStr должен быть сохранён отдельно при создании db
	// Пример: var connStr = "user=... dbname=... sslmode=disable"
	// Здесь должен быть connStr, переданный заранее (не получаемый из db)

	return &PostgresLogListener{
		db: db,
	}, nil
}

func (p *PostgresLogListener) Start(connStr string) error {
	// minReconnect, maxReconnect - как часто пытаться переподключиться
	minReconnect := 10 * time.Second
	maxReconnect := time.Minute

	p.listener = pq.NewListener(connStr, minReconnect, maxReconnect, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("listener event: %v, error: %v", ev, err)
		}
	})

	if err := p.listener.Listen("events"); err != nil {
		return err
	}

	go func() {
		log.Println("PostgresLogListener started listening...")
		for {
			select {
			case notification := <-p.listener.Notify:
				if notification != nil && p.hubChan != nil {
					p.hubChan <- notification.Extra
				}
			case <-time.After(90 * time.Second):
				// Периодический пинг
				go func() {
					if err := p.listener.Ping(); err != nil {
						log.Println("listener ping error:", err)
					}
				}()
			}
		}
	}()

	return nil
}

func (p *PostgresLogListener) SetStreamHub(hub *streamhub.StreamHub) {
	// Получаем канал публикации
	p.hubChan = hub.SubscribeAsProducer()
}

func (p *PostgresLogListener) Stop() {
	log.Println("Stopping PostgresLogListener...")
	if p.listener != nil {
		_ = p.listener.Close()
	}
	_ = p.db.Close()
}
