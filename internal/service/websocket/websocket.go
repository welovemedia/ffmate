package websocket

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/lib/pq"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/websocket"
)

type WebsocketSubject = string

const (
	TASK_CREATED WebsocketSubject = "task:created"
	TASK_UPDATED WebsocketSubject = "task:updated"
	TASK_DELETED WebsocketSubject = "task:deleted"

	PRESET_CREATED WebsocketSubject = "preset:created"
	PRESET_UPDATED WebsocketSubject = "preset:updated"
	PRESET_DELETED WebsocketSubject = "preset:deleted"

	WATCHFOLDER_CREATED WebsocketSubject = "watchfolder:created"
	WATCHFOLDER_UPDATED WebsocketSubject = "watchfolder:updated"
	WATCHFOLDER_DELETED WebsocketSubject = "watchfolder:deleted"

	WEBHOOK_CREATED WebsocketSubject = "webhook:created"
	WEBHOOK_UPDATED WebsocketSubject = "webhook:updated"
	WEBHOOK_DELETED WebsocketSubject = "webhook:deleted"

	WEBHOOK_EXECUTION_CREATED WebsocketSubject = "webhookExecution:created"

	SETTINGS_UPDATED WebsocketSubject = "settings:updated"

	CLIENT_UPDATED WebsocketSubject = "client:updated"

	LOG WebsocketSubject = "log:created"
)

var (
	connections = make(map[string]*websocket.Conn)
	mu          = sync.Mutex{}
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	s := &Service{
		db: db,
	}

	// process broadcast queue
	go s.processBroadcastQueue()

	return s
}

func (s *Service) Add(uuid string, c *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	connections[uuid] = c
	metrics.Gauge("websocket.connect").Inc()
}

func (s *Service) Remove(uuid string) {
	mu.Lock()
	defer mu.Unlock()
	delete(connections, uuid)
	metrics.Gauge("websocket.disconnect").Inc()
}

/**
 * Broadcast
 */

type broadcastMessage struct {
	subject WebsocketSubject
	msg     any
}

var broadcastQueue = make(chan broadcastMessage, 1000)

func (s *Service) Broadcast(subject WebsocketSubject, msg any) {
	select {
	case broadcastQueue <- broadcastMessage{subject, msg}:
	default:
		debug.Websocket.Debug("dropped local broadcast due to blocked channel (full)")
	}

	if subject != LOG && isCluster {
		select {
		case notifyQueue <- &ClusterUpdate{Subject: subject, Payload: msg, Client: session}:
		default:
			debug.Websocket.Debug("dropped cluster broadcast due to blocked channel (full)")
		}
	}
}

func (s *Service) processBroadcastQueue() {
	for b := range broadcastQueue {
		s.broadcastLocal(b.subject, b.msg)
	}
}

func (s *Service) broadcastLocal(subject WebsocketSubject, msg any) {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range connections {
		c.WriteJSON(map[string]any{"subject": subject, "payload": msg})
		metrics.Gauge("websocket.broadcast").Inc()
	}
}

/**
 * Cluster broadcasting
 */

type ClusterUpdate struct {
	Subject WebsocketSubject `json:"subject"`
	Payload any              `json:"payload"`
	Client  string           `json:"client"`
}

var notifyQueue = make(chan *ClusterUpdate, 1000)
var isCluster = false
var session = ""

func (s *Service) InitCluster() {
	session = cfg.GetString("ffmate.session")
	isCluster = true
	go s.listenCluster()
	go s.notifyCluster()
}

func (s *Service) listenCluster() {
	listener := pq.NewListener(cfg.GetString("ffmate.database"), 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			debug.Websocket.Error("listener error:", err)
		}
	})

	// Listen on the same channel used by notifyCluster
	if err := listener.Listen("ffmate"); err != nil {
		debug.Websocket.Error("failed to listen on channel:", err)
		return
	}

	debug.Log.Info("cluster listener started")
	for {
		select {
		case n := <-listener.Notify:
			if n == nil {
				continue
			}

			compressed, err := base64.StdEncoding.DecodeString(n.Extra)
			if err != nil {
				fmt.Println("failed to decode notification:", err)
				continue
			}
			decompressed, err := s.decompressPayload(compressed)
			if err != nil {
				fmt.Println("failed to decompress notification:", err)
				continue
			}

			var payload ClusterUpdate
			if err := json.Unmarshal(decompressed, &payload); err != nil {
				fmt.Println("failed to parse notification:", err)
				continue
			}

			if payload.Client != session {
				debug.Websocket.Debug("> %s from %s (size: %db)", payload.Subject, payload.Client, len(n.Extra))

				// remove self from external clients
				if payload.Subject == string(CLIENT_UPDATED) {
					delete(payload.Payload.(map[string]interface{}), "self")
				}

				select {
				case broadcastQueue <- broadcastMessage{payload.Subject, payload.Payload}:
				default:
					debug.Websocket.Warn("dropped local broadcast due to blocked channel (full)")
				}
			}

		case <-time.After(90 * time.Second):
			go listener.Ping()
		}
	}
}

func (s *Service) notifyCluster() {
	debug.Log.Info("cluster notifier started")
	go func() {
		for update := range notifyQueue {
			payloadBytes, err := json.Marshal(update)
			if err != nil {
				debug.Websocket.Error("failed to marshal message:", err)
				return
			}

			compressed, err := s.compressPayload(payloadBytes)
			if err != nil {
				debug.Websocket.Error("failed to compress message:", err)
				return
			}

			encoded := base64.StdEncoding.EncodeToString(compressed)

			sql := `SELECT pg_notify('ffmate', ?)`
			if err := s.db.Exec(sql, encoded).Error; err != nil {
				fmt.Println("failed to send cluster notification:", err)
				return
			}

			debug.Websocket.Debug("< %s from %s (size: %db/%d%s)", update.Subject, update.Client, len(encoded), len(compressed)*100/len(payloadBytes), "%")
		}
	}()
}

func (s *Service) compressPayload(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

func (s *Service) decompressPayload(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}

func (s *Service) Name() string {
	return service.Websocket
}
