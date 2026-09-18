package legacyimport

import (
	"strings"
	"testing"
)

func TestReadSport(t *testing.T) {
	rows, err := readSport(strings.NewReader("key,name,unit,comment\nтурник,Турник,шт,Подтягивания\n"))
	if err != nil || len(rows) != 1 || rows[0].key != "турник" || rows[0].unit != "шт" {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
}

func TestReadSportActivity(t *testing.T) {
	rows, err := readSportActivity(strings.NewReader("dt,sport_key,sets\n2025-01-02,турник,\"{5,5,3.5}\"\n"))
	if err != nil || len(rows) != 1 || len(rows[0].sets) != 3 || rows[0].sets[2] != 3.5 {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
}

func TestReadSportActivityRejectsInvalidRows(t *testing.T) {
	for _, input := range []string{
		"dt,sport_key,sets\nwrong,турник,\"{5}\"\n",
		"dt,sport_key,sets\n2025-01-02,,\"{5}\"\n",
		"dt,sport_key,sets\n2025-01-02,турник,\"{}\"\n",
		"dt,sport_key,sets\n2025-01-02,турник,\"{5,0}\"\n",
		"dt,sport_key,sets\n2025-01-02,турник,\"{5}\"\n2025-01-02,турник,\"{6}\"\n",
	} {
		if _, err := readSportActivity(strings.NewReader(input)); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}
