package entity

import "time"

type MealType string

const (
	MealBreakfast    MealType = "завтрак"
	MealBeforeLunch  MealType = "до обеда"
	MealLunch        MealType = "обед"
	MealSnack        MealType = "полдник"
	MealBeforeDinner MealType = "до ужина"
	MealDinner       MealType = "ужин"
)

var MealTypes = []MealType{
	MealBreakfast,
	MealBeforeLunch,
	MealLunch,
	MealSnack,
	MealBeforeDinner,
	MealDinner,
}

type JournalTotals struct {
	Weight  float64
	Cal     float64
	Protein float64
	Fat     float64
	Carb    float64
}

type MacroPercent struct {
	Protein float64
	Fat     float64
	Carb    float64
}

type JournalItem struct {
	Meal   MealType
	Food   Food
	Weight float64
}

type JournalItemData struct {
	FoodKey string
	Weight  float64
}

type JournalZone struct {
	Meal   MealType
	Items  []JournalItem
	Totals JournalTotals
}

type JournalDay struct {
	DT           time.Time
	Zones        []JournalZone
	Totals       JournalTotals
	MacroPercent MacroPercent
}
