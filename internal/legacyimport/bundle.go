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

const bundleFileName = "bundle.csv"

var legacyBundleNamespace = [16]byte{
	0x96, 0x92, 0x56, 0x2a, 0xd2, 0x0b, 0x46, 0xbd,
	0x96, 0xb6, 0xab, 0x8c, 0x8c, 0x9a, 0x44, 0x69,
}

type bundleLoader struct {
	path string
}

type bundleRow struct {
	key     string
	foodKey string
	weight  float64
}

func newBundleLoader(dataDirectory string) bundleLoader {
	return bundleLoader{path: filepath.Join(dataDirectory, bundleFileName)}
}

func (bundleLoader) Name() string {
	return "bundle"
}

func (loader bundleLoader) Path() string {
	return loader.path
}

func (loader bundleLoader) Import(ctx context.Context, transaction *sql.Tx) (int, error) {
	rows, err := readBundleFile(loader.path)
	if err != nil {
		return 0, err
	}

	bundleStatement, err := transaction.PrepareContext(ctx, `
INSERT INTO bundle (key, name)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET name = EXCLUDED.name`)
	if err != nil {
		return 0, fmt.Errorf("prepare bundles: %w", err)
	}
	defer bundleStatement.Close()
	deleteStatement, err := transaction.PrepareContext(ctx, `DELETE FROM bundle_item WHERE bundle_key = $1`)
	if err != nil {
		return 0, fmt.Errorf("prepare bundle cleanup: %w", err)
	}
	defer deleteStatement.Close()
	itemStatement, err := transaction.PrepareContext(ctx, `
INSERT INTO bundle_item (bundle_key, food_key, weight)
VALUES ($1, $2, $3)`)
	if err != nil {
		return 0, fmt.Errorf("prepare bundle items: %w", err)
	}
	defer itemStatement.Close()

	initialized := make(map[string]struct{})
	for _, row := range rows {
		bundleKey := legacyBundleUUID(row.key)
		if _, exists := initialized[row.key]; !exists {
			if _, err := bundleStatement.ExecContext(ctx, bundleKey, row.key); err != nil {
				return 0, fmt.Errorf("write bundle %q: %w", row.key, err)
			}
			if _, err := deleteStatement.ExecContext(ctx, bundleKey); err != nil {
				return 0, fmt.Errorf("clear bundle %q: %w", row.key, err)
			}
			initialized[row.key] = struct{}{}
		}
		if _, err := itemStatement.ExecContext(ctx, bundleKey, legacyFoodUUID(row.foodKey), row.weight); err != nil {
			return 0, fmt.Errorf("write bundle %q food %q: %w", row.key, row.foodKey, err)
		}
	}
	return len(rows), nil
}

func readBundleFile(path string) ([]bundleRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()
	return readBundle(file)
}

func readBundle(input io.Reader) ([]bundleRow, error) {
	reader := csv.NewReader(input)
	if err := readHeader(reader, []string{"key", "foodkey", "weight"}); err != nil {
		return nil, err
	}

	rows := make([]bundleRow, 0)
	pairs := make(map[string]struct{})
	for line := 2; ; line++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		key := strings.TrimSpace(record[0])
		foodKey := strings.TrimSpace(record[1])
		if key == "" || foodKey == "" {
			return nil, fmt.Errorf("row %d: key and foodkey are required", line)
		}
		pair := key + "\x00" + foodKey
		if _, exists := pairs[pair]; exists {
			return nil, fmt.Errorf("row %d: duplicate bundle/food pair", line)
		}
		pairs[pair] = struct{}{}
		weight, err := parsePositive(record[2])
		if err != nil {
			return nil, fmt.Errorf("row %d field weight: %w", line, err)
		}
		rows = append(rows, bundleRow{key: key, foodKey: foodKey, weight: weight})
	}
	return rows, nil
}

func legacyBundleUUID(key string) string {
	payload := append(append([]byte(nil), legacyBundleNamespace[:]...), []byte(key)...)
	digest := sha1.Sum(payload)
	digest[6] = (digest[6] & 0x0f) | 0x50
	digest[8] = (digest[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(digest[:16])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}
