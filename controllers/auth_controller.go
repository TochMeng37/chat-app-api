package controllers

import (
	"database/sql"
	"log"
	"time"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Register new user
func Register(c *fiber.Ctx, db *sql.DB) error {
	type RegisterInput struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var data RegisterInput
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), 12)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "password hashing failed"})
	}

	// Insert into DB
	_, err = db.Exec(`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)`,
		data.Username, data.Email, string(hash))
	if err != nil {
		log.Println("insert error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "user already exists"})
	}

	return c.JSON(fiber.Map{"message": "registered successfully"})
}

// Login user
func Login(c *fiber.Ctx, db *sql.DB) error {
	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var data LoginInput
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	var id int
	var username, passwordHash, email string
	err := db.QueryRow(`SELECT id, username, password_hash, email FROM users WHERE email = $1`, data.Email).
		Scan(&id, &username, &passwordHash, &email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}

	// Compare passwords
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(data.Password)) != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
	}

	// Create JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  id,
		"username": username,
		"email":    email,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token creation failed"})
	}

	return c.JSON(fiber.Map{
		"token":    tokenString,
		"username": username,
		"email":    email,
	})
}
