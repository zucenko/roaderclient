package main

import (
	"github.com/gorilla/websocket"
	"github.com/zucenko/roader/model"
)

type GameSession struct {
	Connected   bool
	Conn        *websocket.Conn
	PlayerKey   int32
	PlayerIndex int
	Model       *model.Model
	Errors      chan struct{}
	MessagesOut chan model.ClientMessage
	MessagesIn  chan model.ServerMessage
}
