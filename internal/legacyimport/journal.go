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

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

const journalFileName = "journal.csv"

type journalLoader struct {
	path   string
	userID string
}

type journalRow struct {
	dt         time.Time
	meal       entity.MealType
	foodKey    string
	foodWeight float64
}

func newJournalLoader(dataDirectory, userID string) journalLoader {
	return journalLoader{path: filepath.Join(dataDirectory, journalFileName), userID: userID}
}

func (journalLoader) Name() string {
	return "journal"
}

func (loader journalLoader) Path() string {
	return loader.path
}

func (loader journalLoader) Import(ctx context.Context, transaction *sql.Tx) (int, error) {
	if loader.userID == "" {
		return 0, fmt.Errorf("user ID is required")
	}
	rows, err := readJournalFile(loader.path)
	if err != nil {
		return 0, err
	}
	statement, err := transaction.PrepareContext(ctx, `
INSERT INTO journal (user_id, dt, meal, food_key, food_weight)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, dt, meal, food_key) DO UPDATE
SET food_weight = EXCLUDED.food_weight`)
	if err != nil {
		return 0, fmt.Errorf("prepare journal: %w", err)
	}
	defer statement.Close()
	for _, row := range rows {
		if _, err := statement.ExecContext(ctx, loader.userID, row.dt, row.meal, legacyFoodUUID(row.foodKey), row.foodWeight); err != nil {
			return 0, fmt.Errorf("write journal %s/%s/%s: %w", row.dt.Format("2006-01-02"), row.meal, row.foodKey, err)
		}
	}
	return len(rows), nil
}

func readJournalFile(path string) ([]journalRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readJournal(file)
}

func readJournal(input io.Reader) ([]journalRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"dt", "meal", "foodkey", "foodweight"}); err != nil {
		return nil, err
	}
	rows := make([]journalRow, 0)
	keys := make(map[string]struct{})
	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		dtRaw := strings.TrimSpace(record[0])
		dt, err := time.Parse("2006-01-02", dtRaw)
		if err != nil {
			return nil, fmt.Errorf("row %d field dt: %w", line, err)
		}
		meal := entity.MealType(strings.TrimSpace(record[1]))
		if !isLegacyMeal(meal) {
			return nil, fmt.Errorf("row %d field meal: unknown meal %q", line, meal)
		}
		foodKey := strings.TrimSpace(record[2])
		if foodKey == "" {
			return nil, fmt.Errorf("row %d field foodkey: value is required", line)
		}
		key := dtRaw + "\x00" + string(meal) + "\x00" + foodKey
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("row %d: duplicate date/meal/food pair", line)
		}
		keys[key] = struct{}{}
		weight, err := parsePositive(record[3])
		if err != nil {
			return nil, fmt.Errorf("row %d field foodweight: %w", line, err)
		}
		rows = append(rows, journalRow{dt: dt, meal: meal, foodKey: foodKey, foodWeight: weight})
	}
	return rows, nil
}

func isLegacyMeal(meal entity.MealType) bool {
	for _, allowed := range entity.MealTypes {
		if meal == allowed {
			return true
		}
	}
	return false
}
