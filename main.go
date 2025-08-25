package main

import (
	"log"
	"os"

	"github.com/AlexSamarskii/debezium_implementing/cmd"
	"github.com/AlexSamarskii/debezium_implementing/config"
	"github.com/AlexSamarskii/debezium_implementing/internal/db"
	"github.com/AlexSamarskii/debezium_implementing/internal/es"
	"github.com/AlexSamarskii/debezium_implementing/internal/migration"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	migrator := migration.NewMigrator(cfg, "file://migrations")
	if err := migrator.Apply(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	dbConn, err := db.NewConnection(cfg)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer dbConn.Close()

	esClient, err := es.NewClient([]string{cfg.Elasticsearch.URL})
	if err != nil {
		log.Fatalf("ES client failed: %v", err)
	}

	httpCmd := cmd.Http{
		Connection: dbConn,
		Port:       cfg.Server.Port,
		EsClient:   esClient,
		KafkaCfg:   cfg.Kafka,
	}.Command()

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return migrator.Apply()
		},
	}

	root := &cobra.Command{Use: "app"}
	root.AddCommand(httpCmd, migrateCmd)

	if err := root.Execute(); err != nil {
		log.Printf("Command failed: %v", err)
		os.Exit(1)
	}
}
