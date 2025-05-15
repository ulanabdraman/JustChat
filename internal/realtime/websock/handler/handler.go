package handler

import (
	"context"
	"log"

	"JustChat/internal/realtime/websock/usecase"
)

type WebSockHandler interface {
	ListenAndServe(ctx context.Context)
}

type webSockHandler struct {
	messageCh      <-chan []byte
	webSockUsecase usecase.WebSockUsecase
}

func NewWebSockHandler(messageCh <-chan []byte, webSockUsecase usecase.WebSockUsecase) WebSockHandler {
	return &webSockHandler{
		messageCh:      messageCh,
		webSockUsecase: webSockUsecase,
	}
}

func (h *webSockHandler) ListenAndServe(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("websock handler shutting down")
			return
		case msg, ok := <-h.messageCh:
			if !ok {
				log.Println("message channel closed, stopping handler")
				return
			}
			if err := h.webSockUsecase.SendMessage(msg); err != nil {
				log.Printf("websock handler error: %v", err)
			}
		}
	}
}
