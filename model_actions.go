package main

import (
	"encoding/gob"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/zucenko/roader/model"
	"net/url"
	"os"
	"os/signal"
)

func NewGameSession() *GameSession {
	return &GameSession{
		Connected:   false,
		MessagesOut: make(chan model.ClientMessage, 10),
		MessagesIn:  make(chan model.ServerMessage, 10),
	}
}

// asynchro call
func (gs *GameSession) Connect(host string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: host, Path: "/play"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return err
	}
	gs.Conn = c
	go gs.LoopChannelRead()
	go gs.LoopChannelWrite()

	return nil
}

func (gs *GameSession) LoopChannelRead() {
	log.Printf("LoopChannelRead STARTED")
loop:
	for {
		messageType, r, err := gs.Conn.NextReader()
		if err != nil {
			log.Warn("LoopChannelRead err %v", err)
			gs.Errors <- struct{}{}
			break loop
		}
		log.Printf("LoopChannelRead received  message type: %d", messageType)
		dec := gob.NewDecoder(r)
		sm := &model.ServerMessage{}
		err = dec.Decode(sm)
		if err != nil {
			log.Warn("cant decode message %v", err)
			gs.Errors <- struct{}{}
			break loop
		}
		log.Infof("mess %v", sm)
		select {
		case gs.MessagesIn <- *sm:
		default:
			log.Warnf("LoopChannelRead Dropping Data red from socket but.. gs.MessagesIn full")
		}
	}
	log.Printf("LoopChannelRead ENDED")
}

// this function only consumes. no worries about full buffer stuck
func (gs *GameSession) LoopChannelWrite() {
	log.Printf("GameSession.LoopChannelWrite STARTED")
loop:
	for {
		select {
		case cm := <-gs.MessagesOut:
			log.Printf("GameSession.LoopChannelWrite have message")
			w, err := gs.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Warn("GameSession.LoopChannelWrite cant get writer")
				gs.Errors <- struct{}{}
				break loop
			}
			enc := gob.NewEncoder(w)
			err = enc.Encode(cm)
			if err != nil {
				log.Warn("GameSession.LoopChannelWrite cant encode")
				gs.Errors <- struct{}{}
				break loop
			}
			err = w.Close()
			if err != nil {
				log.Warn("GameSession.LoopChannelWrite cant Close")
				gs.Errors <- struct{}{}
				break loop
			}
		}
	}
	log.Printf("LoopChannelWrite ENDED")
}

func (gs *GameSession) Loop() {
	//log.Info("GameSession.Loop STARTING")
	//for {
	select {
	case sm := <-gs.MessagesIn:
		log.Info(sm)
		if len(sm.Setup) == 1 {
			gs.PlayerKey = sm.Setup[0].PlayerKey
			log.Info("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX  I AM: %v", gs.PlayerKey)
			gs.Model = model.NewEmptyModel(sm.Setup[0].Cols, sm.Setup[0].Rows, sm.Setup[0].Players)
		}

		for _, directionSuccess := range sm.Directions {
			if directionSuccess.Success {
				player := gs.Model.Players[directionSuccess.PlayerKey]
				var cell = gs.Model.Matrix[player.Col][player.Row]
				var newCell *model.Cell
				//nonportal
				if directionSuccess.Direction < 4 {
					newCell = cell.Paths[directionSuccess.Direction].Target
					cell.Paths[directionSuccess.Direction].Wall = false
					cell.Paths[directionSuccess.Direction].Player = player
					cell.Paths[directionSuccess.Direction].Target.Paths[(directionSuccess.Direction+2)%4].Player = player
					cell.Paths[directionSuccess.Direction].Target.Paths[(directionSuccess.Direction+2)%4].Wall = false
					newCell.Diamond = false
					newCell.Key = false
					cell.Crossings()
					newCell.Crossings()
				} else {
					newCell = gs.Model.Matrix[directionSuccess.Col][directionSuccess.Row]
				}
				//newCell.Unhook(player.Id)

				cell.Player = nil
				newCell.Player = player
				player.Row = directionSuccess.Row
				player.Col = directionSuccess.Col
			}
		}

		for _, v := range sm.Visibles {
			cell := gs.Model.Matrix[v.Col][v.Row]
			if len(v.Walls) == 4 {
				for i, p := range cell.Paths {
					if p != nil {
						p.Wall = v.Walls[i]
						p.Lock = v.Locks[i]
						if p.Target != nil {
							p.Target.Paths[(i+2)%4].Wall = v.Walls[i]
							p.Target.Paths[(i+2)%4].Lock = v.Locks[i]
						}
					}
				}
			}
			cell.Key = v.Key
			cell.Diamond = v.Diamond
			if v.Portal && cell.Portal == nil {
				cell.Portal = &model.Portal{Target: gs.Model.Matrix[v.PortalToCol][v.PortalToRow]}
				gs.Model.Matrix[v.PortalToCol][v.PortalToRow].Portal = &model.Portal{Target: cell}
			}
			if v.HasPlayer {
				cell.Player = gs.Model.Players[v.PlayerId]
			} else {
				cell.Player = nil
			}
		}
		for _, p := range sm.Picks {
			gs.Model.Players[gs.PlayerKey].Keys = p.Keys
			gs.Model.Players[gs.PlayerKey].Diamonds = p.Diamonds
		}
	default:

	}
	//}
}
