package cases

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Sport struct {
	repository  port.SportRepository
	generateKey func() (string, error)
}

func NewSport(repository port.SportRepository) *Sport {
	return &Sport{repository: repository, generateKey: randomUUID}
}

func (sport *Sport) List(ctx context.Context, request entity.PageRequest) (entity.Page[entity.Sport], error) {
	request, err := normalizePageRequest(request)
	if err != nil {
		return entity.Page[entity.Sport]{}, err
	}
	return sport.repository.List(ctx, request)
}
func (sport *Sport) Get(ctx context.Context, key string) (entity.Sport, error) {
	key, err := validateSportKey(key)
	if err != nil {
		return entity.Sport{}, err
	}
	return sport.repository.Get(ctx, key)
}
func (sport *Sport) Create(ctx context.Context, data entity.SportData) (entity.Sport, error) {
	data, err := validateSport(data)
	if err != nil {
		return entity.Sport{}, err
	}
	key, err := sport.generateKey()
	if err != nil {
		return entity.Sport{}, fmt.Errorf("generate sport key: %w", err)
	}
	return sport.repository.Create(ctx, entity.Sport{Key: key, Name: data.Name, Unit: data.Unit, Comment: data.Comment})
}
func (sport *Sport) Update(ctx context.Context, key string, data entity.SportData) (entity.Sport, error) {
	key, err := validateSportKey(key)
	if err != nil {
		return entity.Sport{}, err
	}
	data, err = validateSport(data)
	if err != nil {
		return entity.Sport{}, err
	}
	return sport.repository.Update(ctx, key, data)
}
func (sport *Sport) Delete(ctx context.Context, key string) error {
	key, err := validateSportKey(key)
	if err != nil {
		return err
	}
	return sport.repository.Delete(ctx, key)
}
func validateSport(data entity.SportData) (entity.SportData, error) {
	data.Name, data.Unit, data.Comment = strings.TrimSpace(data.Name), strings.TrimSpace(data.Unit), strings.TrimSpace(data.Comment)
	details := map[string][]string{}
	if data.Name == "" {
		details["name"] = []string{"Название обязательно"}
	}
	if data.Unit == "" {
		details["unit"] = []string{"Единица измерения обязательна"}
	}
	if len(details) > 0 {
		return entity.SportData{}, &entity.ValidationError{Fields: details}
	}
	return data, nil
}
func validateSportKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", &entity.ValidationError{Fields: map[string][]string{"sportKey": {"Вид спорта обязателен"}}}
	}
	return key, nil
}

type SportActivity struct{ repository port.SportActivityRepository }

func NewSportActivity(repository port.SportActivityRepository) *SportActivity {
	return &SportActivity{repository: repository}
}
func (activity *SportActivity) List(ctx context.Context, userID string, period entity.SportActivityRange) ([]entity.SportActivity, error) {
	userID, period, err := validateSportActivityRange(userID, period)
	if err != nil {
		return nil, err
	}
	return activity.repository.List(ctx, userID, period)
}
func (activity *SportActivity) Save(ctx context.Context, userID string, data entity.SportActivityData) (entity.SportActivity, error) {
	userID, data, err := validateSportActivity(userID, data, true)
	if err != nil {
		return entity.SportActivity{}, err
	}
	return activity.repository.Save(ctx, userID, data)
}
func (activity *SportActivity) Delete(ctx context.Context, userID string, data entity.SportActivityData) error {
	userID, data, err := validateSportActivity(userID, data, false)
	if err != nil {
		return err
	}
	return activity.repository.Delete(ctx, userID, data)
}
func validateSportActivity(userID string, data entity.SportActivityData, requireSets bool) (string, entity.SportActivityData, error) {
	userID = strings.TrimSpace(userID)
	details := map[string][]string{}
	validateUserID(details, userID)
	if data.DT.IsZero() {
		details["dt"] = []string{"Дата обязательна"}
	} else {
		data.DT = dateOnly(data.DT)
	}
	data.SportKey = strings.TrimSpace(data.SportKey)
	if data.SportKey == "" {
		details["sportKey"] = []string{"Вид спорта обязателен"}
	}
	if requireSets && len(data.Sets) == 0 {
		details["sets"] = []string{"Добавьте хотя бы один подход"}
	}
	if requireSets {
		for _, set := range data.Sets {
			if math.IsNaN(set) || math.IsInf(set, 0) || set <= 0 {
				details["sets"] = []string{"Каждый подход должен быть конечным положительным числом"}
				break
			}
		}
	}
	if len(details) > 0 {
		return "", entity.SportActivityData{}, &entity.ValidationError{Fields: details}
	}
	return userID, data, nil
}
func validateSportActivityRange(userID string, period entity.SportActivityRange) (string, entity.SportActivityRange, error) {
	userID = strings.TrimSpace(userID)
	details := map[string][]string{}
	validateUserID(details, userID)
	if period.From != nil {
		value := dateOnly(*period.From)
		period.From = &value
	}
	if period.To != nil {
		value := dateOnly(*period.To)
		period.To = &value
	}
	if period.From != nil && period.To != nil && period.From.After(*period.To) {
		details["from"] = []string{"Начальная дата должна быть не позже конечной"}
	}
	if len(details) > 0 {
		return "", entity.SportActivityRange{}, &entity.ValidationError{Fields: details}
	}
	return userID, period, nil
}
