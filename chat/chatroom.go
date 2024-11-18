package chat

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

type Client struct {
	ID      string
	Channel chan string
	mu      sync.Mutex // Protects the channel closure
	closed  bool
}

// NewClient creates a new client
func NewClient(id string) *Client {
	return &Client{
		ID:      id,
		Channel: make(chan string, 100), // Buffered channel to prevent blocking
		closed:  false,
	}
}

type ChatRoom struct {
	db         *sql.DB
	clients    map[string]*Client
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewChatRoom(db *sql.DB) *ChatRoom {
	return &ChatRoom{
		db:         db,
		clients:    make(map[string]*Client),
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (c *ChatRoom) Run() {
	for {
		select {
		case client := <-c.register:
			c.mu.Lock()
			c.clients[client.ID] = client
			c.mu.Unlock()
			log.Printf("Client %s joined", client.ID)
			_, _ = c.db.Exec("INSERT OR IGNORE INTO clients (id) VALUES (?)", client.ID)

		case client := <-c.unregister:
			c.mu.Lock()
			delete(c.clients, client.ID)
			close(client.Channel)
			c.mu.Unlock()
			log.Printf("Client %s left", client.ID)
			_, _ = c.db.Exec("UPDATE clients SET active = 0 WHERE id = ?", client.ID)

		case message := <-c.broadcast:
			c.mu.Lock()
			for _, client := range c.clients {
				client.Channel <- message
			}
			c.mu.Unlock()
		}
	}
}

func (c *ChatRoom) RegisterClient(id string) {
	client := &Client{
		ID:      id,
		Channel: make(chan string, 100),
	}
	c.register <- client
}

func (c *ChatRoom) UnregisterClient(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if client, exists := c.clients[id]; exists && client.Channel != nil {
		c.unregister <- client
		delete(c.clients, id) // Remove client from map immediately
		client.Close()        // Close the channel gracefully
	}
}

// Close closes the client's channel safely
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		close(c.Channel)
		c.closed = true
	}
}
func (c *ChatRoom) BroadcastMessage(senderID, message string) {
	formatted := senderID + ": " + message
	c.broadcast <- formatted
	_, _ = c.db.Exec("INSERT INTO messages (sender_id, message) VALUES (?, ?)", senderID, message)
}

func (c *ChatRoom) GetMessage(clientID string, timeout time.Duration) (string, bool) {
	c.mu.Lock()
	client, exists := c.clients[clientID]
	c.mu.Unlock()
	if !exists {
		return "", false
	}
	select {
	case msg := <-client.Channel:
		return msg, true
	case <-time.After(timeout):
		return "", false
	}
}
