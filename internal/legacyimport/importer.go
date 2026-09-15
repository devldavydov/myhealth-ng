// Package legacyimport contains the temporary importer for data exported from
// the previous MyHealth implementation. It intentionally does not take part in
// the application dependency graph and can be deleted after migration.
package legacyimport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Config struct {
	DataDirectory string
	UserID        string
}

type DatasetReport struct {
	Name  string
	Count int
}

type Report struct {
	Datasets []DatasetReport
}

// datasetLoader is the extension point for future legacy sections. Each
// dataset owns its source format and SQL in a separate file; the orchestrator
// only provides one transaction and aggregates the report.
type datasetLoader interface {
	Name() string
	Import(context.Context, *sql.Tx) (int, error)
}

func Import(ctx context.Context, db *sql.DB, config Config) (Report, error) {
	config.DataDirectory = strings.TrimSpace(config.DataDirectory)
	config.UserID = strings.TrimSpace(config.UserID)
	if config.DataDirectory == "" {
		return Report{}, fmt.Errorf("legacy data directory is required")
	}

	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Report{}, fmt.Errorf("begin legacy import transaction: %w", err)
	}
	defer transaction.Rollback()

	report := Report{Datasets: make([]DatasetReport, 0)}
	for _, loader := range datasetLoaders(config) {
		count, err := loader.Import(ctx, transaction)
		if err != nil {
			return Report{}, fmt.Errorf("import legacy dataset %q: %w", loader.Name(), err)
		}
		report.Datasets = append(report.Datasets, DatasetReport{Name: loader.Name(), Count: count})
	}

	if err := transaction.Commit(); err != nil {
		return Report{}, fmt.Errorf("commit legacy import: %w", err)
	}
	return report, nil
}

// Register each future section here. Its parsing and SQL should remain in a
// dedicated file so deleting the bridge never affects production adapters.
func datasetLoaders(config Config) []datasetLoader {
	return []datasetLoader{
		newFoodLoader(config.DataDirectory),
		newWeightLoader(config.DataDirectory, config.UserID),
	}
}
