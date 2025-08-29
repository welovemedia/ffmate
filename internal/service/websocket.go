package service

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/yosev/debugo"
)

type websocketSvc struct {
	service
}

type Subject = string

const (
	TASK_CREATED Subject = "task:created"
	TASK_UPDATED Subject = "task:updated"
	TASK_DELETED Subject = "task:deleted"

	PRESET_CREATED Subject = "preset:created"
	PRESET_UPDATED Subject = "preset:updated"
	PRESET_DELETED Subject = "preset:deleted"

	WATCHFOLDER_CREATED Subject = "watchfolder:created"
	WATCHFOLDER_UPDATED Subject = "watchfolder:updated"
	WATCHFOLDER_DELETED Subject = "watchfolder:deleted"

	WEBHOOK_CREATED Subject = "webhook:created"
	WEBHOOK_UPDATED Subject = "webhook:updated"
	WEBHOOK_DELETED Subject = "webhook:deleted"

	BATCH_CREATED  Subject = "batch:created"
	BATCH_FINISHED Subject = "batch:finished"

	LOG Subject = "log"
)

type message struct {
	Subject string `json:"subject"`
	Payload any    `json:"payload"`
}

var debug = debugo.New("websocket:service")

var (
	conns = make(map[string]*websocket.Conn)
	mutex = sync.RWMutex{}
)

func (s *websocketSvc) AddConnection(uuid string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	conns[uuid] = conn
}

func (s *websocketSvc) RemoveConnection(uuid string, conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(conns, uuid)
}

func (s *websocketSvc) Broadcast(subject Subject, msg any) error {
	mutex.Lock()
	defer mutex.Unlock()
	for _, conn := range conns {
		conn.WriteJSON(&message{Subject: subject, Payload: msg})
	}
	return nil
}
