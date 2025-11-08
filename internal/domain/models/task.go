package models

import (
	"perx/internal/domain/dto"
	"time"
)

type TaskStatus string

const (
	statusInQueue    TaskStatus = "in queue"
	statusInProgress TaskStatus = "in progress"
	statusCompleted  TaskStatus = "completed"
)

type TaskModel struct {
	ID                int        `json:"task_id" example:"1"`
	TaskStatus        TaskStatus `json:"task_status" example:"in_queue"`
	QueueIndex        int        `json:"queue_index" example:"0"`
	ElementsNumber    int        `json:"n" example:"5"`
	Delta             float64    `json:"d" example:"11.5"`
	StartValue        float64    `json:"n1" example:"21.33"`
	IterationInterval float64    `json:"I" example:"5.5"`
	TTL               float64    `json:"TTL" example:"120.0"`
	CurrentIteration  int        `json:"current_iteration" example:"3"`
	CurrentValue      float64    `json:"current_value" example:"21.33"`
	TaskCreatedAt     time.Time  `json:"task_created_at" example:"2022-01-01T00:00:00+00:00"`
	TaskStartedAt     time.Time  `json:"task_started_at" example:"2022-01-01T00:00:00+00:00"`
	TaskFinishedAt    time.Time  `json:"task_finished_at" example:"2022-01-01T00:00:00+00:00"`
}

func (m *TaskModel) CreateTask(task *dto.TaskDTO, id int) {
	m.ID = id
	m.TaskStatus = statusInQueue
	m.ElementsNumber = task.ElementsNumber
	m.Delta = task.Delta
	m.StartValue = task.StartValue
	m.IterationInterval = task.IterationInterval
	m.TTL = task.TTL
	m.TaskCreatedAt = time.Now()
}

func (m *TaskModel) SetStatusInProgress() {
	m.TaskStatus = statusInProgress
	m.TaskStartedAt = time.Now()
}

func (m *TaskModel) SetStatusCompleted() {
	m.TaskStatus = statusCompleted
	m.QueueIndex = -1
	m.TaskFinishedAt = time.Now()
}
