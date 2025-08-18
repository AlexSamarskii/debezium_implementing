package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type Migrate struct {
	Connection *sql.DB
	Source     string
}

func (m Migrate) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrate",
		RunE: func(_ *cobra.Command, _ []string) error {
			return m.main()
		},
	}

	cmd.Flags().StringVar(&m.Source, "source", "", "Path to migration SQL file")
	_ = cmd.MarkFlagRequired("source")

	return cmd
}

func (m Migrate) main() error {
	data, err := os.ReadFile(m.Source)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if len(data) == 0 {
		log.Println("Migration file is empty, skipping")
		return nil
	}

	log.Printf("Executing migration from %s", m.Source)

	if _, err := m.Connection.Exec(string(data)); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migration applied successfully")
	return nil
}
