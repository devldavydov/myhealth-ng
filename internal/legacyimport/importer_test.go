package legacyimport

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDatasetLoadersUseCurrentExportNames(t *testing.T) {
	loaders := datasetLoaders(Config{
		DataDirectory: "/legacy",
		UserID:        "local-user",
	})

	food, ok := loaders[0].(foodLoader)
	if !ok || food.path != filepath.Join("/legacy", "food.csv") {
		t.Fatalf("unexpected food loader: %#v", loaders[0])
	}
	weight, ok := loaders[1].(weightLoader)
	if !ok || weight.path != filepath.Join("/legacy", "weight.csv") {
		t.Fatalf("unexpected weight loader: %#v", loaders[1])
	}
	bundle, ok := loaders[2].(bundleLoader)
	if !ok || bundle.path != filepath.Join("/legacy", "bundle.csv") {
		t.Fatalf("unexpected bundle loader: %#v", loaders[2])
	}
}

func TestReadBundle(t *testing.T) {
	input := `"key","foodkey","weight"
"завтрак","сыр_гауда",48.0
"завтрак","хлеб_бел",40.0
`
	rows, err := readBundle(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].key != "завтрак" || rows[0].foodKey != "сыр_гауда" || rows[0].weight != 48 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
	if legacyBundleUUID("завтрак") != legacyBundleUUID("завтрак") || legacyBundleUUID("завтрак") == legacyBundleUUID("обед") {
		t.Fatal("legacy bundle UUID must be stable and key-specific")
	}
}

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
