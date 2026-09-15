package entity

import "time"

type Weight struct {
	DT    time.Time
	Value float64
}

type WeightRange struct {
	From *time.Time
	To   *time.Time
}
