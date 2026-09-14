package domain

// MeasurementType identifies a supported health metric.
type MeasurementType string

const (
	MeasurementTypeWeight   MeasurementType = "weight"
	MeasurementTypePressure MeasurementType = "pressure"
	MeasurementTypePulse    MeasurementType = "pulse"
)

func (measurementType MeasurementType) Valid() bool {
	switch measurementType {
	case MeasurementTypeWeight, MeasurementTypePressure, MeasurementTypePulse:
		return true
	default:
		return false
	}
}

type Measurement struct {
	ID         string          `json:"id"`
	Type       MeasurementType `json:"type"`
	Value      float64         `json:"value"`
	Unit       string          `json:"unit"`
	MeasuredAt string          `json:"measuredAt"`
}

type NewMeasurement struct {
	Type       MeasurementType
	Value      float64
	Unit       string
	MeasuredAt string
}
