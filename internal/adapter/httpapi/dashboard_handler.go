package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type dashboardHandler struct {
	dashboard port.DashboardUseCases
}

type dashboardCalorieDayResponse struct {
	DT      string  `json:"dt"`
	Balance float64 `json:"balance"`
}

type dashboardCaloriesResponse struct {
	Days    []dashboardCalorieDayResponse `json:"days"`
	Average *float64                      `json:"average"`
}

type dashboardActivityResponse struct {
	SportKey string  `json:"sportKey"`
	Name     string  `json:"name"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
	Unit     string  `json:"unit"`
}

type dashboardResponse struct {
	Calories     *dashboardCaloriesResponse  `json:"calories"`
	WeightChange *float64                    `json:"weightChange"`
	Activities   []dashboardActivityResponse `json:"activities"`
}

func newDashboardHandler(dashboard port.DashboardUseCases) *dashboardHandler {
	return &dashboardHandler{dashboard: dashboard}
}

func (handler *dashboardHandler) get(ctx *gin.Context) {
	period, err := decodeDashboardRange(ctx)
	if err != nil {
		respondDashboardError(ctx, err)
		return
	}
	result, err := handler.dashboard.Get(ctx.Request.Context(), currentUserID(ctx), period)
	if err != nil {
		respondDashboardError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toDashboardResponse(result)})
}

func decodeDashboardRange(ctx *gin.Context) (entity.DashboardRange, error) {
	details := make(map[string][]string)
	var from, to time.Time
	if ctx.Query("from") == "" {
		details["from"] = []string{"Поле обязательно"}
	} else {
		parsed, err := parseWeightDate(ctx.Query("from"), "from")
		if err != nil {
			return entity.DashboardRange{}, err
		}
		from = parsed
	}
	if ctx.Query("to") == "" {
		details["to"] = []string{"Поле обязательно"}
	} else {
		parsed, err := parseWeightDate(ctx.Query("to"), "to")
		if err != nil {
			return entity.DashboardRange{}, err
		}
		to = parsed
	}
	if len(details) > 0 {
		return entity.DashboardRange{}, &entity.ValidationError{Fields: details}
	}
	if from.After(to) {
		return entity.DashboardRange{}, &entity.ValidationError{Fields: map[string][]string{
			"from": {"Начальная дата должна быть не позже конечной"},
		}}
	}
	return entity.DashboardRange{From: from, To: to}, nil
}

func toDashboardResponse(result entity.Dashboard) dashboardResponse {
	response := dashboardResponse{
		WeightChange: result.WeightChange,
		Activities:   make([]dashboardActivityResponse, 0, len(result.Activities)),
	}
	if result.Calories != nil {
		calories := dashboardCaloriesResponse{
			Days:    make([]dashboardCalorieDayResponse, 0, len(result.Calories.Days)),
			Average: result.Calories.Average,
		}
		for _, day := range result.Calories.Days {
			calories.Days = append(calories.Days, dashboardCalorieDayResponse{DT: day.DT.Format(dateLayout), Balance: day.Balance})
		}
		response.Calories = &calories
	}
	for _, item := range result.Activities {
		response.Activities = append(response.Activities, dashboardActivityResponse{
			SportKey: item.SportKey, Name: item.Name, Count: item.Count, Total: item.Total, Unit: item.Unit,
		})
	}
	return response
}

func respondDashboardError(ctx *gin.Context, err error) {
	var validation *entity.ValidationError
	if errors.As(err, &validation) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный период дашборда", "details": validation.Fields})
		return
	}
	_ = ctx.Error(err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
}
