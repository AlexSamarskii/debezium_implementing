package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"

	"github.com/AlexSamarskii/debezium_implementing/config"
	"github.com/AlexSamarskii/debezium_implementing/internal/es"
	"github.com/AlexSamarskii/debezium_implementing/internal/kafka"
	ht "github.com/AlexSamarskii/debezium_implementing/internal/transport/http"
)

type Http struct {
	Connection *sql.DB
	Port       int
	EsClient   *es.Client
	KafkaCfg   config.KafkaConfig
}

func (h Http) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "http",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.main(cmd.Context())
		},
	}
	cmd.Flags().IntVar(&h.Port, "port", 8080, "HTTP server port")
	return cmd
}

func (h Http) main(rootCtx context.Context) error {
	handler := ht.Handler{
		Connection: h.Connection,
	}

	app := fiber.New()

	app.Get("/api", handler.HandleGetRequests)
	app.Post("/api", handler.HandlePostRequests)

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", h.Port)); err != nil && err != http.ErrServerClosed {
			log.Printf("Fiber server error: %v", err)
		}
	}()

	kafkaConsumer := kafka.NewConsumer(
		h.KafkaCfg.Brokers,
		h.KafkaCfg.Topic,
		h.KafkaCfg.Group,
		h.EsClient,
	)

	log.Println(ctx)

	go func() {
		if err := kafkaConsumer.Consume(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
