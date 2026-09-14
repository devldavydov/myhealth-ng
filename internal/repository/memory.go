package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"sort"
	"sync"

	"github.com/devldavydov/myhealth-ng/internal/domain"
)

var seedMeasurements = []domain.Measurement{
	{ID: "sample-weight", Type: domain.MeasurementTypeWeight, Value: 72.4, Unit: "кг", MeasuredAt: "2026-09-12T08:00:00.000Z"},
	{ID: "sample-pulse", Type: domain.MeasurementTypePulse, Value: 68, Unit: "уд/мин", MeasuredAt: "2026-09-13T07:30:00.000Z"},
}

type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []domain.Measurement
}

func NewInMemoryMeasurementRepository(initial []domain.Measurement) *InMemoryMeasurementRepository {
	measurements := initial
	if initial == nil {
		measurements = seedMeasurements
	}
	stored := make([]domain.Measurement, len(measurements))
	copy(stored, measurements)
	return &InMemoryMeasurementRepository{measurements: stored}
}

func (repository *InMemoryMeasurementRepository) FindAll(_ context.Context) ([]domain.Measurement, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	measurements := make([]domain.Measurement, len(repository.measurements))
	copy(measurements, repository.measurements)
	sort.Slice(measurements, func(left, right int) bool {
		return measurements[left].MeasuredAt > measurements[right].MeasuredAt
	})
	return measurements, nil
}

func (repository *InMemoryMeasurementRepository) Create(_ context.Context, data domain.NewMeasurement) (domain.Measurement, error) {
	id, err := randomUUID()
	if err != nil {
		return domain.Measurement{}, err
	}
	measurement := domain.Measurement{
		ID:         id,
		Type:       data.Type,
		Value:      data.Value,
		Unit:       data.Unit,
		MeasuredAt: data.MeasuredAt,
	}

	repository.mu.Lock()
	repository.measurements = append(repository.measurements, measurement)
	repository.mu.Unlock()
	return measurement, nil
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

type InMemoryUserRegistry struct {
	mu    sync.RWMutex
	users map[string]domain.UserIdentity
}

func NewInMemoryUserRegistry() *InMemoryUserRegistry {
	return &InMemoryUserRegistry{users: make(map[string]domain.UserIdentity)}
}

func (registry *InMemoryUserRegistry) Remember(_ context.Context, user domain.UserIdentity) error {
	registry.mu.Lock()
	registry.users[user.GUID] = user
	registry.mu.Unlock()
	return nil
}
