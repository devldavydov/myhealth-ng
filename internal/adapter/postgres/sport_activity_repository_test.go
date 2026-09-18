package postgres

import (
	"testing"
	"time"
)

type sportActivityScannerStub struct{ sets string }

func (stub sportActivityScannerStub) Scan(dest ...any) error {
	*dest[0].(*time.Time) = time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
	*dest[1].(*string) = "турник"
	*dest[2].(*string) = "Турник"
	*dest[3].(*string) = "шт"
	*dest[4].(*string) = "Подтягивания"
	*dest[5].(*string) = stub.sets
	return nil
}

func TestScanSportActivityDecodesRealArrayJSON(t *testing.T) {
	item, err := scanSportActivity(sportActivityScannerStub{sets: `[5,3,3.5]`})
	if err != nil {
		t.Fatal(err)
	}
	if len(item.Sets) != 3 || item.Sets[0] != 5 || item.Sets[2] != 3.5 {
		t.Fatalf("sets=%v", item.Sets)
	}
}

func TestScanSportActivityRejectsInvalidArrayJSON(t *testing.T) {
	if _, err := scanSportActivity(sportActivityScannerStub{sets: `{5,3}`}); err == nil {
		t.Fatal("expected JSON decode error")
	}
}
