// Command legacy-import is a temporary development-only bridge from the
// previous MyHealth CSV exports to the current PostgreSQL schema.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	postgresadapter "github.com/devldavydov/myhealth-ng/internal/adapter/postgres"
	"github.com/devldavydov/myhealth-ng/internal/legacyimport"
)

const operationTimeout = time.Minute

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("legacy-import", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	databaseURL := flags.String("database-url", "", "PostgreSQL connection string")
	dataDirectory := flags.String("data-dir", "legacy_data", "legacy CSV directory")
	userID := flags.String("user-id", "", "user ID assigned to user-scoped legacy entries")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if strings.TrimSpace(*databaseURL) == "" {
		return fmt.Errorf("--database-url is required")
	}
	if strings.TrimSpace(*dataDirectory) == "" {
		return fmt.Errorf("--data-dir is required")
	}
	if strings.TrimSpace(*userID) == "" {
		return fmt.Errorf("--user-id is required")
	}

	database, err := sql.Open("pgx", *databaseURL)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("connect PostgreSQL: %w", err)
	}
	if err := postgresadapter.Migrate(ctx, database); err != nil {
		return err
	}
	report, err := legacyimport.Import(ctx, database, legacyimport.Config{
		DataDirectory: *dataDirectory,
		UserID:        *userID,
	})
	if err != nil {
		return err
	}
	total := 0
	imported := 0
	for _, dataset := range report.Datasets {
		if dataset.Skipped {
			log.Printf("Legacy dataset skipped (file is absent): %s", dataset.Name)
			continue
		}
		log.Printf("Legacy dataset imported: %s=%d", dataset.Name, dataset.Count)
		imported++
		total += dataset.Count
	}
	log.Printf("Legacy import complete: datasets=%d, rows=%d", imported, total)
	return nil
}
