package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/AlexSamarskii/debezium_implementing/internal/models"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	esClient  ElasticsearchClient
	topic     string
	partition int
}

type ElasticsearchClient interface {
	IndexUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id int) error
}

func NewConsumer(brokers []string, topic, groupID string, esClient ElasticsearchClient) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1e3,
			MaxBytes:       1e6,
			MaxWait:        1 * time.Second,
			CommitInterval: 1 * time.Second,
		}),
		esClient: esClient,
		topic:    topic,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer...")
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("Error fetching message: %v", err)
				}
				return err
			}

			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("Failed to process message: %v", err)
				continue
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var event models.DebeziumEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	switch event.Payload.Op {
	case "c", "u": // create or update
		var user models.User
		if err := json.Unmarshal(event.Payload.After, &user); err != nil {
			return err
		}
		return c.esClient.IndexUser(ctx, user)

	case "d": // delete
		var user models.User
		if err := json.Unmarshal(event.Payload.Before, &user); err != nil {
			return err
		}
		return c.esClient.DeleteUser(ctx, user.ID)

	case "r": // read
		var user models.User
		if err := json.Unmarshal(event.Payload.After, &user); err != nil {
			return err
		}
		return c.esClient.IndexUser(ctx, user)

	default:
		log.Printf("Unknown operation: %s", event.Payload.Op)
	}

	return nil
}
