package es

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/AlexSamarskii/debezium_implementing/internal/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Client struct {
	es *elasticsearch.Client
}

func NewClient(addresses []string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
		// Username: "elastic",
		// Password: "password",
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES info error: %s", res.String())
	}

	log.Println("Connected to Elasticsearch")
	return &Client{es: es}, nil
}

func (c *Client) IndexUser(ctx context.Context, user models.User) error {
	data, _ := json.Marshal(user)

	req := esapi.IndexRequest{
		Index:      "users",
		DocumentID: strconv.Itoa(user.ID),
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing user %d: %s", user.ID, res.String())
	}

	log.Printf("Indexed user: %+v", user)
	return nil
}

func (c *Client) DeleteUser(ctx context.Context, id int) error {
	req := esapi.DeleteRequest{
		Index:      "users",
		DocumentID: strconv.Itoa(id),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && !strings.Contains(res.String(), "404") {
		return fmt.Errorf("error deleting user %d: %s", id, res.String())
	}

	log.Printf("Deleted user with ID=%d", id)
	return nil
}
