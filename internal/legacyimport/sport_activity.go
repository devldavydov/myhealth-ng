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

const sportActivityFileName = "sport_activity.csv"

type sportActivityLoader struct{ path, userID string }
type sportActivityRow struct {
	dt       time.Time
	sportKey string
	sets     []float64
}

func newSportActivityLoader(dataDirectory, userID string) sportActivityLoader {
	return sportActivityLoader{path: filepath.Join(dataDirectory, sportActivityFileName), userID: userID}
}
func (sportActivityLoader) Name() string        { return "sport activity" }
func (loader sportActivityLoader) Path() string { return loader.path }
func (loader sportActivityLoader) Import(ctx context.Context, tx *sql.Tx) (int, error) {
	if loader.userID == "" {
		return 0, fmt.Errorf("user ID is required")
	}
	rows, err := readSportActivityFile(loader.path)
	if err != nil {
		return 0, err
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO sport_activity (user_id,dt,sport_key,sets) VALUES ($1,$2,$3,$4) ON CONFLICT (user_id,dt,sport_key) DO UPDATE SET sets=EXCLUDED.sets`)
	if err != nil {
		return 0, fmt.Errorf("prepare sport activity: %w", err)
	}
	defer statement.Close()
	for _, row := range rows {
		if _, err := statement.ExecContext(ctx, loader.userID, row.dt, row.sportKey, row.sets); err != nil {
			return 0, fmt.Errorf("write sport activity %s/%s: %w", row.dt.Format("2006-01-02"), row.sportKey, err)
		}
	}
	return len(rows), nil
}
func readSportActivityFile(path string) ([]sportActivityRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readSportActivity(file)
}
func readSportActivity(input io.Reader) ([]sportActivityRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"dt", "sport_key", "sets"}); err != nil {
		return nil, err
	}
	rows := make([]sportActivityRow, 0)
	keys := map[string]struct{}{}
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
		sportKey := strings.TrimSpace(record[1])
		if sportKey == "" {
			return nil, fmt.Errorf("row %d field sport_key: value is required", line)
		}
		key := dtRaw + "\x00" + sportKey
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("row %d: duplicate date/sport pair", line)
		}
		keys[key] = struct{}{}
		sets, err := parseLegacySets(record[2])
		if err != nil {
			return nil, fmt.Errorf("row %d field sets: %w", line, err)
		}
		rows = append(rows, sportActivityRow{dt: dt, sportKey: sportKey, sets: sets})
	}
	return rows, nil
}
func parseLegacySets(raw string) ([]float64, error) {
	value := strings.TrimSpace(raw)
	if len(value) < 2 || value[0] != '{' || value[len(value)-1] != '}' {
		return nil, fmt.Errorf("expected PostgreSQL array, got %q", raw)
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	if inner == "" {
		return nil, fmt.Errorf("at least one set is required")
	}
	parts := strings.Split(inner, ",")
	sets := make([]float64, len(parts))
	for index, part := range parts {
		parsed, err := parsePositive(part)
		if err != nil {
			return nil, err
		}
		sets[index] = parsed
	}
	return sets, nil
}
