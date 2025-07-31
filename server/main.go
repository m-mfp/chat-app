package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type Client struct {
	Conn *websocket.Conn
}

type WebSocketManager struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.Mutex
}

var Manager = WebSocketManager{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte, 100),
	Register:   make(chan *Client, 10),
	Unregister: make(chan *Client, 10),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return strings.HasPrefix(r.Header.Get("Origin"), "http://192.168.") ||
			strings.HasPrefix(r.Header.Get("Origin"), "http://localhost:") ||
			r.Header.Get("Origin") == "http://client:80"
	},
}

func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.Mutex.Lock()
			manager.Clients[client] = true
			manager.Mutex.Unlock()
			fmt.Println("Client registered, total clients:", len(manager.Clients))
		case client := <-manager.Unregister:
			manager.Mutex.Lock()
			if _, ok := manager.Clients[client]; ok {
				delete(manager.Clients, client)
				client.Conn.Close()
				fmt.Println("Client unregistered, total clients:", len(manager.Clients))
			}
			manager.Mutex.Unlock()
		case message := <-manager.Broadcast:
			manager.Mutex.Lock()
			for client := range manager.Clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					fmt.Println("Write error:", err)
					delete(manager.Clients, client)
					client.Conn.Close()
				}
			}
			manager.Mutex.Unlock()
		}
	}
}

func main() {
	ConnectDB()
	DB.AutoMigrate(&Message{})
	router := gin.Default()

	go Manager.Start()

	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			fmt.Println("CORS origin check:", origin)
			return strings.HasPrefix(origin, "http://192.168.") ||
				strings.HasPrefix(origin, "http://localhost:") ||
				origin == "http://client:80"
		},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Upgrade", "Connection"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path != "/ws" {
			cors.New(config)(c)
			RateLimiter()(c)
		}
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		if err := DB.Exec("SELECT 1").Error; err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "Database connection failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/api/messages", HandleMessages)
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
	var msg Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}
	Manager.Broadcast <- []byte(fmt.Sprintf(`{"type":"message","user":"%s","text":"%s"}`, msg.UserID, msg.Text))
	c.Status(http.StatusNoContent)
}

func HandleWebSocket(c *gin.Context) {
	fmt.Println("WebSocket connection attempt from:", c.Request.RemoteAddr)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}
	client := &Client{Conn: conn}
	Manager.Register <- client
	go func() {
		defer func() {
			Manager.Unregister <- client
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("WebSocket read error:", err)
				return
			}
			var msgData struct {
				Type string `json:"type"`
				User string `json:"user"`
				Text string `json:"text,omitempty"`
			}
			if err := json.Unmarshal(message, &msgData); err != nil {
				fmt.Println("JSON parse error:", err)
				continue
			}
			if msgData.Type == "typing" {
				// Broadcast typing event without saving to DB
				Manager.Broadcast <- []byte(fmt.Sprintf(`{"type":"typing","user":"%s"}`, msgData.User))
				continue
			}
			// Handle message type
			if msgData.Type == "message" {
				msg := Message{
					UserID:    msgData.User,
					MsgID:     fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000)),
					Text:      msgData.Text,
					CreatedAt: time.Now(),
				}
				if err := DB.Create(&msg).Error; err != nil {
					fmt.Println("Database save error:", err)
					continue
				}
				Manager.Broadcast <- message
			}
		}
	}()
}
