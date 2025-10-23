package controllers

import (
	"Golang_backend/database"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Message struct {
	ID      int    `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	SentAt  string `json:"sent_at"`
}

var Clients = make(map[string]*websocket.Conn)

func InitWebSocketController(app *fiber.App, db *sql.DB) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(handleWebSocket))
}

func handleWebSocket(c *websocket.Conn) {
	// Get token from query
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.Close()
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		c.Close()
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)
	userID := int(claims["user_id"].(float64))

	Clients[username] = c
	log.Println(username, "connected")

	defer func() {
		delete(Clients, username)
		log.Println(username, "disconnected")
		c.Close()
	}()

	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		// Save message to DB
		_, err = database.DB.Exec(
			"INSERT INTO messages (sender_id, receiver_id, content) VALUES ($1,(SELECT id FROM users WHERE username=$2),$3)",
			userID, msg.To, msg.Content,
		)
		if err != nil {
			log.Println("DB save error:", err)
		}

		// Send to recipient if online
		if conn, ok := Clients[msg.To]; ok {
			conn.WriteJSON(msg)
		}
	}
}

func GetChatHistory(c *fiber.Ctx, db *sql.DB) error {
	// Get current username from JWT middleware
	userAny := c.Locals("username")
	if userAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized: username not found in context",
		})
	}

	currentUser, ok := userAny.(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "invalid username type",
		})
	}

	// Get target username from query param
	withUser := c.Query("with")
	if withUser == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing 'with' query parameter",
		})
	}

	// Query messages between two users
	rows, err := db.Query(`
		SELECT sender, recipient, content, created_at
		FROM messages
		WHERE (sender = $1 AND recipient = $2)
		   OR (sender = $2 AND recipient = $1)
		ORDER BY created_at ASC
	`, currentUser, withUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var sender, recipient, content string
		var createdAt time.Time

		if err := rows.Scan(&sender, &recipient, &content, &createdAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		history = append(history, map[string]interface{}{
			"sender":     sender,
			"recipient":  recipient,
			"content":    content,
			"created_at": createdAt,
		})
	}

	return c.JSON(history)
}
