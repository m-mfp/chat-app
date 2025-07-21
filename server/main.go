package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
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
}

var Manager = WebSocketManager{
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		fmt.Println("WebSocket origin:", r.Header.Get("Origin"))
		return true // Restrict origins in production
	},
}

func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.Clients[client] = true
			fmt.Println("Client registered")
		case client := <-manager.Unregister:
			if _, ok := manager.Clients[client]; ok {
				delete(manager.Clients, client)
				fmt.Println("Client unregistered")
			}
		case message := <-manager.Broadcast:
			for client := range manager.Clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					fmt.Println("Write error:", err)
					client.Conn.Close()
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

	go Manager.Start()

	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://192.168.2.") || origin == "http://localhost:3000" || origin == "http://client:80"
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
	Manager.Broadcast <- []byte(fmt.Sprintf(`{"user":"%s","text":"%s"}`, msg.UserID, msg.Text))
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
			conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("WebSocket read error:", err)
				return
			}
			var msgData struct {
				User string `json:"user"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(message, &msgData); err != nil {
				fmt.Println("JSON parse error:", err)
				continue
			}
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
	}()
}
