package voting

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type Client struct {
	conn     *websocket.Conn
	memberID int
	send     chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WS bağlandı → üye %d (toplam: %d)", client.memberID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WS ayrıldı → üye %d (toplam: %d)", client.memberID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) ConnectedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// ServeWS — gin handler, WebSocket bağlantısını yönetir
func (h *Hub) ServeWS(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "oturum gerekli"})
		return
	}

	websocket.Handler(func(conn *websocket.Conn) {
		client := &Client{
			conn:     conn,
			memberID: memberID.(int),
			send:     make(chan []byte, 64),
		}
		h.register <- client

		// Gönderici goroutine
		go func() {
			for msg := range client.send {
				if _, err := conn.Write(msg); err != nil {
					break
				}
			}
		}()

		// Okuyucu — bağlantı kapanana kadar bekler
		buf := make([]byte, 512)
		for {
			if _, err := conn.Read(buf); err != nil {
				break
			}
		}
		h.unregister <- client
	}).ServeHTTP(c.Writer, c.Request)
}
