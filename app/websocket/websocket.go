package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WSMessage struct {
	Namespace string `json:"namespace"`
	Event     string `json:"event"`
	Payload   string `json:"payload"`
}

type Client struct {
	conn      *websocket.Conn
	namespace string
	send      chan WSMessage
}

var (
	upgrader     = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients      = make(map[string]map[*Client]bool)
	clientsMutex sync.Mutex
)

func HandleWebSocket(c echo.Context) error {
	namespace := c.Param("namespace")
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return err
	}

	client := &Client{
		conn:      conn,
		namespace: namespace,
		send:      make(chan WSMessage),
	}

	registerClient(client)

	go readPump(client)
	go writePump(client)

	return nil
}

func registerClient(client *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if clients[client.namespace] == nil {
		clients[client.namespace] = make(map[*Client]bool)
	}
	clients[client.namespace][client] = true
	log.Printf("Client connected to namespace: %s\n", client.namespace)
}

func unregisterClient(client *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if clients[client.namespace] != nil {
		delete(clients[client.namespace], client)
		if len(clients[client.namespace]) == 0 {
			delete(clients, client.namespace)
		}
	}
	log.Printf("Client disconnected from namespace: %s\n", client.namespace)
	client.conn.Close()
}

func readPump(client *Client) {
	defer unregisterClient(client)

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Println("invalid message format:", err)
			continue
		}

		log.Printf("[%s] %s: %s", wsMsg.Namespace, wsMsg.Event, wsMsg.Payload)

		// Echo back to sender
		client.send <- WSMessage{
			Namespace: wsMsg.Namespace,
			Event:     "echo",
			Payload:   fmt.Sprintf("You said: %s", wsMsg.Payload),
		}

		// Broadcast to others
		broadcastToNamespace(wsMsg, client)
	}
}

func writePump(client *Client) {
	for msg := range client.send {
		if err := client.conn.WriteJSON(msg); err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func broadcastToNamespace(msg WSMessage, sender *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients[msg.Namespace] {
		if client != sender {
			client.send <- msg
		}
	}
}
