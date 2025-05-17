package dblisteners

import "JustChat/pkg/streamhub"

type LogListener interface {
	Start() error                          // запуск прослушивания
	Stop() error                           // остановка
	SetStreamHub(hub *streamhub.StreamHub) // установка хаба для отправки сообщений
}
