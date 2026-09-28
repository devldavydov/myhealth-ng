package entity

import (
	"fmt"
	"time"
)

const FoodStatisticsExcludedName = "Много еды"

type Food struct {
	Key     string
	Name    string
	Brand   string
	Cal100  float64
	Prot100 float64
	Fat100  float64
	Carb100 float64
	Comment string
}

type FoodData struct {
	Name    string
	Brand   string
	Cal100  float64
	Prot100 float64
	Fat100  float64
	Carb100 float64
	Comment string
}

type FoodStatistics struct {
	TotalWeight          float64
	FirstConsumedDate    *time.Time
	LastConsumedDate     *time.Time
	AveragePortionWeight *float64
}

type FoodDetails struct {
	Food       Food
	Statistics *FoodStatistics
}

type ValidationError struct {
	Fields map[string][]string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", err.Fields)
}
