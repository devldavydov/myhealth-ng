package cases

import (
	"context"
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Dashboard struct {
	repository port.DashboardRepository
}

func NewDashboard(repository port.DashboardRepository) *Dashboard {
	return &Dashboard{repository: repository}
}

func (dashboard *Dashboard) Get(ctx context.Context, userID string, period entity.DashboardRange) (entity.Dashboard, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if period.From.IsZero() {
		details["from"] = []string{"Начальная дата обязательна"}
	} else {
		period.From = dateOnly(period.From)
	}
	if period.To.IsZero() {
		details["to"] = []string{"Конечная дата обязательна"}
	} else {
		period.To = dateOnly(period.To)
	}
	if !period.From.IsZero() && !period.To.IsZero() && period.From.After(period.To) {
		details["from"] = []string{"Начальная дата должна быть не позже конечной"}
	}
	if len(details) > 0 {
		return entity.Dashboard{}, &entity.ValidationError{Fields: details}
	}

	source, err := dashboard.repository.Load(ctx, userID, period)
	if err != nil {
		return entity.Dashboard{}, err
	}
	result := entity.Dashboard{Activities: source.Activities}
	if result.Activities == nil {
		result.Activities = make([]entity.DashboardActivity, 0)
	}
	if source.DefaultDailyCalorieLimit != nil {
		calories := entity.DashboardCalories{Days: make([]entity.DashboardCalorieDay, 0, len(source.CalorieDays))}
		var total float64
		for _, day := range source.CalorieDays {
			limit := float64(*source.DefaultDailyCalorieLimit)
			if day.ActiveLimit != nil {
				limit = *day.ActiveLimit
			}
			balance := limit - day.Consumed
			calories.Days = append(calories.Days, entity.DashboardCalorieDay{DT: day.DT, Balance: balance})
			total += balance
		}
		if len(calories.Days) > 0 {
			average := total / float64(len(calories.Days))
			calories.Average = &average
		}
		result.Calories = &calories
	}
	if len(source.Weights) >= 2 {
		change := source.Weights[len(source.Weights)-1].Value - source.Weights[0].Value
		result.WeightChange = &change
	}
	return result, nil
}
