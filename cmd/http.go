package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"

	"github.com/AlexSamarskii/debezium_implementing/internal/transport/http"
)

type Http struct {
	Connection *sql.DB
	Port       int
}

func (h Http) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "http",
		RunE: func(_ *cobra.Command, _ []string) error {
			return h.main()
		},
	}

	cmd.Flags().IntVar(&h.Port, "port", 8080, "HTTP server port")
	return cmd
}

func (h Http) main() error {
	handler := http.Handler{
		Connection: h.Connection,
	}

	app := fiber.New()

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", h.Port)); err != nil && err != http.ErrServerClosed {
			log.Printf("Fiber server error: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
