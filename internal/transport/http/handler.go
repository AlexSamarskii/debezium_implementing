package http

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/AlexSamarskii/debezium_implementing/internal/transport/http/models"
	"github.com/gofiber/fiber/v2"
)

const (
	insertQuery = `INSERT INTO debezium.users (name, email) VALUES ($1, $2)`
	selectQuery = `SELECT id, name, email FROM debezium.users`
)

type Handler struct {
	Connection *sql.DB
}

func (h *Handler) HandlePostRequests(c *fiber.Ctx) error {
	var req models.Request

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	_, err := h.Connection.ExecContext(ctx, insertQuery, req.Name, req.Email)
	if err != nil {
		log.Printf("Insert failed: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save user",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "User added successfully",
	})
}

func (h *Handler) HandleGetRequests(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.Connection.QueryContext(ctx, selectQuery)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	defer rows.Close()

	var users []models.Response

	for rows.Next() {
		var user models.Response
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			log.Printf("Scan failed: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to scan user",
			})
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error reading rows",
		})
	}

	return c.JSON(users)
}
