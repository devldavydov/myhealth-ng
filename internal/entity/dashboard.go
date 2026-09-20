package entity

import "time"

type DashboardRange struct {
	From time.Time
	To   time.Time
}

type DashboardCalorieSourceDay struct {
	DT          time.Time
	Consumed    float64
	ActiveLimit *float64
}

type DashboardActivity struct {
	SportKey string
	Name     string
	Count    int
	Total    float64
	Unit     string
}

type DashboardSource struct {
	DefaultDailyCalorieLimit *int
	CalorieDays              []DashboardCalorieSourceDay
	Weights                  []Weight
	Activities               []DashboardActivity
}

type DashboardCalorieDay struct {
	DT      time.Time
	Balance float64
}

type DashboardCalories struct {
	Days    []DashboardCalorieDay
	Average *float64
}

type Dashboard struct {
	Calories     *DashboardCalories
	WeightChange *float64
	Activities   []DashboardActivity
}
