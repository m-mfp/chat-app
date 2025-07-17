package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// WebSocket Manager and Client structs
type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type WebSocketManager struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

var Manager = WebSocketManager{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.Clients[client] = true
		case client := <-manager.Unregister:
			if _, ok := manager.Clients[client]; ok {
				close(client.Send)
				delete(manager.Clients, client)
			}
		case message := <-manager.Broadcast:
			for client := range manager.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(manager.Clients, client)
				}
			}
		}
	}
}

func main() {
	ConnectDB()
	DB.AutoMigrate(&Message{})
	router := gin.Default()

	// Start WebSocket manager
	go Manager.Start()

	// CORS configuration
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://client:80"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	router.Use(RateLimiter())

	router.GET("/health", func(c *gin.Context) {
		if err := DB.Exec("SELECT 1").Error; err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Database connection failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/api/messages", HandleMessages)

	// WebSocket endpoint
	router.GET("/ws", HandleWebSocket)

	router.Run(":8000")
}

func RateLimiter() gin.HandlerFunc {
	limiter := rate.NewLimiter(1, 4)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Limit exceeded",
			})
		}
	}
}

func HandleMessages(c *gin.Context) {
	var input struct {
		UserID string `json:"userid"`
		MsgID  string `json:"msgid"`
		Text   string `json:"text"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	msg := Message{
		UserID: input.UserID,
		MsgID:  input.MsgID,
		Text:   input.Text,
	}

	tx := DB.Begin()
	if err := tx.Create(&msg).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	var count int64
	if err := tx.Model(&Message{}).Where("msg_id = ?", msg.MsgID).Count(&count).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify message"})
		return
	}
	if count == 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit message"})
		return
	}

	// Broadcast the message to WebSocket clients
	Manager.Broadcast <- []byte(msg.Text)

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "message received",
		"message": msg,
	})
}

func HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Restrict origins in production
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	client := &Client{Conn: conn, Send: make(chan []byte)}
	Manager.Register <- client

	// Handle sending messages to the client
	go func() {
		defer func() {
			Manager.Unregister <- client
			conn.Close()
		}()
		for message := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}()

	// Handle receiving messages from the client
	go func() {
		defer func() {
			Manager.Unregister <- client
			conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Save to database
			msg := Message{
				UserID:    "user", // Replace with actual user ID (e.g., from auth)
				MsgID:     fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000)),
				Text:      string(message),
				CreatedAt: time.Now(),
			}
			if err := DB.Create(&msg).Error; err != nil {
				continue
			}
			// Broadcast to all clients
			Manager.Broadcast <- message
		}
	}()
}
