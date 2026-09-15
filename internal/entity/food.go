package entity

import "fmt"

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

type ValidationError struct {
	Fields map[string][]string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", err.Fields)
}
