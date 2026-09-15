package legacyimport

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const foodFileName = "food.csv"

var legacyFoodNamespace = [16]byte{
	0x0a, 0x74, 0x38, 0xe6, 0xab, 0x4d, 0x4d, 0x84,
	0x83, 0x63, 0x7d, 0x5d, 0xec, 0x0f, 0x5a, 0xd1,
}

type foodLoader struct {
	path string
}

type foodRow struct {
	key     string
	name    string
	brand   string
	cal100  float64
	prot100 float64
	fat100  float64
	carb100 float64
	comment string
}

func newFoodLoader(dataDirectory string) foodLoader {
	return foodLoader{path: filepath.Join(dataDirectory, foodFileName)}
}

func (foodLoader) Name() string {
	return "food"
}

func (loader foodLoader) Path() string {
	return loader.path
}

func (loader foodLoader) Import(ctx context.Context, transaction *sql.Tx) (int, error) {
	rows, err := readFoodFile(loader.path)
	if err != nil {
		return 0, err
	}

	statement, err := transaction.PrepareContext(ctx, `
INSERT INTO food (key, name, brand, cal100, prot100, fat100, carb100, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (key) DO UPDATE SET
    name = EXCLUDED.name,
    brand = EXCLUDED.brand,
    cal100 = EXCLUDED.cal100,
    prot100 = EXCLUDED.prot100,
    fat100 = EXCLUDED.fat100,
    carb100 = EXCLUDED.carb100,
    comment = EXCLUDED.comment`)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer statement.Close()

	for _, row := range rows {
		_, err := statement.ExecContext(ctx, legacyFoodUUID(row.key), row.name, row.brand, row.cal100, row.prot100, row.fat100, row.carb100, row.comment)
		if err != nil {
			return 0, fmt.Errorf("write food %q: %w", row.key, err)
		}
	}
	return len(rows), nil
}

func readFoodFile(path string) ([]foodRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readFood(file)
}

func readFood(input io.Reader) ([]foodRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"key", "name", "brand", "cal100", "prot100", "fat100", "carb100", "comment"}); err != nil {
		return nil, err
	}

	rows := make([]foodRow, 0)
	keys := make(map[string]struct{})
	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		key := strings.TrimSpace(record[0])
		if key == "" {
			return nil, fmt.Errorf("row %d: key is required", line)
		}
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("row %d: duplicate key %q", line, key)
		}
		keys[key] = struct{}{}

		row := foodRow{
			key:     key,
			name:    strings.TrimSpace(record[1]),
			brand:   strings.TrimSpace(record[2]),
			comment: strings.TrimSpace(record[7]),
		}
		if row.name == "" {
			return nil, fmt.Errorf("row %d: name is required", line)
		}
		values := []*float64{&row.cal100, &row.prot100, &row.fat100, &row.carb100}
		for index, field := range []string{"cal100", "prot100", "fat100", "carb100"} {
			*values[index], err = parseNonNegative(record[index+3])
			if err != nil {
				return nil, fmt.Errorf("row %d field %s: %w", line, field, err)
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// The old textual key is mapped to UUID v5 so repeated imports address the
// same row while all newly created products keep using random UUIDs.
func legacyFoodUUID(key string) string {
	payload := append(append([]byte(nil), legacyFoodNamespace[:]...), []byte(key)...)
	digest := sha1.Sum(payload)
	digest[6] = (digest[6] & 0x0f) | 0x50
	digest[8] = (digest[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(digest[:16])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
