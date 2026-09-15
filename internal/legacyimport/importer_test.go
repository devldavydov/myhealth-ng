package legacyimport

import (
	"strings"
	"testing"
)

func TestReadFood(t *testing.T) {
	input := `"key","name","brand","cal100","prot100","fat100","carb100","comment"
"сыр_гауда","Сыр Гауда","Viola",338.0,26.0,26.0,0.0,"тест"
`
	rows, err := readFood(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	row := rows[0]
	if row.key != "сыр_гауда" || row.name != "Сыр Гауда" || row.cal100 != 338 || row.comment != "тест" {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestReadWeight(t *testing.T) {
	input := `"dt","value"
"2026-08-31",98.3
`
	rows, err := readWeight(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].dt.Format("2006-01-02") != "2026-08-31" || rows[0].value != 98.3 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestReadFoodRejectsInvalidNumber(t *testing.T) {
	input := `key,name,brand,cal100,prot100,fat100,carb100,comment
bad,Test,,-1,1,1,1,
`
	_, err := readFood(strings.NewReader(input))
	if err == nil || !strings.Contains(err.Error(), "row 2 field cal100") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadWeightRejectsDuplicateDate(t *testing.T) {
	input := `dt,value
2026-08-31,98.3
2026-08-31,98.2
`
	_, err := readWeight(strings.NewReader(input))
	if err == nil || !strings.Contains(err.Error(), "duplicate date") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLegacyFoodUUIDIsStableAndValid(t *testing.T) {
	first := legacyFoodUUID("сыр_гауда")
	second := legacyFoodUUID("сыр_гауда")
	other := legacyFoodUUID("сыр_другой")
	if first != second {
		t.Fatalf("UUID is not stable: %q != %q", first, second)
	}
	if first == other {
		t.Fatalf("different keys produced the same UUID: %q", first)
	}
	if len(first) != 36 || first[14] != '5' || first[19] < '8' || first[19] > 'b' {
		t.Fatalf("unexpected UUID v5: %q", first)
	}
}
