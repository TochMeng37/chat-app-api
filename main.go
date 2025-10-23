package main

import (
	"encoding/json"
	"fmt"
	"log"

	"Golang_backend/controllers"
	"Golang_backend/database"
	"Golang_backend/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

// Global chat variables
var clients = make(map[string]*websocket.Conn) // username -> connection
var broadcast = make(chan map[string]string)

func main() {
	// Connect DB
	database.ConnectDB()
	defer database.DB.Close()

	// Start broadcast goroutine
	go func() {
		for {
			msg := <-broadcast
			for username, conn := range clients {
				err := conn.WriteJSON(msg)
				if err != nil {
					fmt.Println("write error:", err)
					conn.Close()
					delete(clients, username)
				}
			}
		}
	}()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Get("/users", func(c *fiber.Ctx) error {
		return controllers.GetAllUsers(c, database.DB)
	})

	// Auth routes
	app.Post("/register", func(c *fiber.Ctx) error {
		return controllers.Register(c, database.DB)
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.Login(c, database.DB)
	})

	// WebSocket route
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// Get token from query
		token := c.Query("token")
		claims, err := middleware.ValidateJWT(token)
		if err != nil {
			c.WriteMessage(websocket.CloseMessage, []byte("invalid token"))
			c.Close()
			return
		}

		username := claims["username"].(string)
		clients[username] = c
		defer func() {
			delete(clients, username)
			c.Close()
			fmt.Println(username, "disconnected")
		}()

		fmt.Println(username, "connected")

		for {
			_, msgBytes, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read error:", err)
				break
			}

			var msg struct {
				To      string `json:"to"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				fmt.Println("unmarshal error:", err)
				continue
			}

			// Save message to DB
			_, err = database.DB.Exec(`
            INSERT INTO messages (sender, recipient, content) VALUES ($1, $2, $3)
        `, username, msg.To, msg.Message)
			if err != nil {
				fmt.Println("DB insert error:", err)
			}

			// Send only to the recipient if connected
			if recipientConn, ok := clients[msg.To]; ok {
				recipientConn.WriteJSON(map[string]string{
					"from":    username,
					"message": msg.Message,
				})
			}
		}
	}))

	// Protected chat routes
	chat := app.Group("/chat", middleware.Protect())
	chat.Get("/history", func(c *fiber.Ctx) error {
		return controllers.GetChatHistory(c, database.DB)
	})

	fmt.Println("🚀 Server running on http://localhost:5000")
	log.Fatal(app.Listen(":5000"))
}
