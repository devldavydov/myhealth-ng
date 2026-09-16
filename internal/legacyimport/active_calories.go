package legacyimport

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const activeCaloriesFileName = "act_cal.csv"

type activeCaloriesLoader struct {
	path   string
	userID string
}

type activeCaloriesRow struct {
	dt    time.Time
	value float64
}

func newActiveCaloriesLoader(dataDirectory, userID string) activeCaloriesLoader {
	return activeCaloriesLoader{path: filepath.Join(dataDirectory, activeCaloriesFileName), userID: userID}
}

func (activeCaloriesLoader) Name() string {
	return "active calories"
}

func (loader activeCaloriesLoader) Path() string {
	return loader.path
}

func (loader activeCaloriesLoader) Import(ctx context.Context, transaction *sql.Tx) (int, error) {
	if loader.userID == "" {
		return 0, fmt.Errorf("user ID is required")
	}
	rows, err := readActiveCaloriesFile(loader.path)
	if err != nil {
		return 0, err
	}
	statement, err := transaction.PrepareContext(ctx, `
INSERT INTO act_calories (user_id, dt, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, dt) DO UPDATE SET value = EXCLUDED.value`)
	if err != nil {
		return 0, fmt.Errorf("prepare active calories: %w", err)
	}
	defer statement.Close()
	for _, row := range rows {
		if _, err := statement.ExecContext(ctx, loader.userID, row.dt, row.value); err != nil {
			return 0, fmt.Errorf("write active calories %s: %w", row.dt.Format("2006-01-02"), err)
		}
	}
	return len(rows), nil
}

func readActiveCaloriesFile(path string) ([]activeCaloriesRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readActiveCalories(file)
}

func readActiveCalories(input io.Reader) ([]activeCaloriesRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"dt", "value"}); err != nil {
		return nil, err
	}
	rows := make([]activeCaloriesRow, 0)
	dates := make(map[string]struct{})
	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		dateValue := strings.TrimSpace(record[0])
		if _, exists := dates[dateValue]; exists {
			return nil, fmt.Errorf("row %d: duplicate date %q", line, dateValue)
		}
		dates[dateValue] = struct{}{}
		dt, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, fmt.Errorf("row %d field dt: %w", line, err)
		}
		value, err := parsePositive(record[1])
		if err != nil {
			return nil, fmt.Errorf("row %d field value: %w", line, err)
		}
		rows = append(rows, activeCaloriesRow{dt: dt, value: value})
	}
	return rows, nil
}
