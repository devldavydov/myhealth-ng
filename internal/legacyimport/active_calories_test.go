package legacyimport

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestActiveCaloriesLoaderIsRegistered(t *testing.T) {
	loaders := datasetLoaders(Config{DataDirectory: "/legacy", UserID: "local-user"})
	for _, loader := range loaders {
		activeCalories, ok := loader.(activeCaloriesLoader)
		if !ok {
			continue
		}
		if activeCalories.path != filepath.Join("/legacy", "act_cal.csv") || activeCalories.userID != "local-user" {
			t.Fatalf("unexpected active calories loader: %#v", activeCalories)
		}
		return
	}
	t.Fatal("active calories loader is not registered")
}

func TestReadActiveCalories(t *testing.T) {
	rows, err := readActiveCalories(strings.NewReader(`"dt","value"
"2026-09-15",2450.5
"2026-09-16",2600
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].dt.Format("2006-01-02") != "2026-09-15" || rows[0].value != 2450.5 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestReadActiveCaloriesRejectsInvalidRows(t *testing.T) {
	for _, input := range []string{
		"dt,value\nwrong,2500\n",
		"dt,value\n2026-09-16,0\n",
		"dt,value\n2026-09-16,2500\n2026-09-16,2600\n",
	} {
		if _, err := readActiveCalories(strings.NewReader(input)); err == nil {
			t.Fatalf("expected validation error for %q", input)
		}
	}
}
