package legacyimport

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func readHeader(reader *csv.Reader, expected []string) error {
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	if len(header) != len(expected) {
		return fmt.Errorf("unexpected header %q", header)
	}
	for index := range expected {
		if header[index] != expected[index] {
			return fmt.Errorf("unexpected header %q", header)
		}
	}
	return nil
}

func parseNonNegative(raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, fmt.Errorf("expected a finite non-negative number, got %q", raw)
	}
	return value, nil
}

func parsePositive(raw string) (float64, error) {
	value, err := parseNonNegative(raw)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("expected a finite positive number, got %q", raw)
	}
	return value, nil
}
