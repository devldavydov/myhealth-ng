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
)

const sportFileName = "sport.csv"

type sportLoader struct{ path string }
type sportRow struct{ key, name, unit, comment string }

func newSportLoader(dataDirectory string) sportLoader {
	return sportLoader{path: filepath.Join(dataDirectory, sportFileName)}
}
func (sportLoader) Name() string        { return "sport" }
func (loader sportLoader) Path() string { return loader.path }
func (loader sportLoader) Import(ctx context.Context, transaction *sql.Tx) (int, error) {
	rows, err := readSportFile(loader.path)
	if err != nil {
		return 0, err
	}
	statement, err := transaction.PrepareContext(ctx, `INSERT INTO sport (key,name,unit,comment) VALUES ($1,$2,$3,$4) ON CONFLICT (key) DO UPDATE SET name=EXCLUDED.name,unit=EXCLUDED.unit,comment=EXCLUDED.comment`)
	if err != nil {
		return 0, fmt.Errorf("prepare sport: %w", err)
	}
	defer statement.Close()
	for _, row := range rows {
		if _, err := statement.ExecContext(ctx, row.key, row.name, row.unit, row.comment); err != nil {
			return 0, fmt.Errorf("write sport %q: %w", row.key, err)
		}
	}
	return len(rows), nil
}
func readSportFile(path string) ([]sportRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readSport(file)
}
func readSport(input io.Reader) ([]sportRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"key", "name", "unit", "comment"}); err != nil {
		return nil, err
	}
	rows := make([]sportRow, 0)
	keys := map[string]struct{}{}
	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		row := sportRow{key: strings.TrimSpace(record[0]), name: strings.TrimSpace(record[1]), unit: strings.TrimSpace(record[2]), comment: strings.TrimSpace(record[3])}
		if row.key == "" || row.name == "" || row.unit == "" {
			return nil, fmt.Errorf("row %d: key, name and unit are required", line)
		}
		if _, exists := keys[row.key]; exists {
			return nil, fmt.Errorf("row %d: duplicate key %q", line, row.key)
		}
		keys[row.key] = struct{}{}
		rows = append(rows, row)
	}
	return rows, nil
}
