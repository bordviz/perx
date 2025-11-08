package dto

import "perx/internal/lib/validator"

type TaskDTO struct {
	ElementsNumber    int     `json:"n" validate:"required" example:"5"`
	Delta             float64 `json:"d" validate:"required" example:"11.5"`
	StartValue        float64 `json:"n1" validate:"required" example:"21.33"`
	IterationInterval float64 `json:"I" validate:"required" example:"5.5"`
	TTL               float64 `json:"TTL" validate:"required" example:"120.0"`
}

func (d *TaskDTO) Validate() error {
	return validator.Validate(d)
}
