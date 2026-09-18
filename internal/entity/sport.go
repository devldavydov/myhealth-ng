package entity

import "time"

type Sport struct {
	Key     string
	Name    string
	Unit    string
	Comment string
}

type SportData struct {
	Name    string
	Unit    string
	Comment string
}

type SportActivity struct {
	DT    time.Time
	Sport Sport
	Sets  []float64
}

type SportActivityData struct {
	DT       time.Time
	SportKey string
	Sets     []float64
}

type SportActivityRange struct {
	From *time.Time
	To   *time.Time
}
