package task

import (
	"github.com/go-chi/chi/v5"
	"log/slog"
	"perx/internal/domain/dto"
	"perx/internal/domain/models"
)

type TaskHandler struct {
	log       *slog.Logger
	taskQueue TaskQueue
}

type TaskQueue interface {
	AddTask(task *dto.TaskDTO) error
	GetTaskList() ([]*models.TaskModel, error)
}

func NewTaskHandler(log *slog.Logger, queue TaskQueue) *TaskHandler {
	return &TaskHandler{
		log:       log,
		taskQueue: queue,
	}
}

func (h *TaskHandler) ConnectHandler() func(chi.Router) {
	return func(r chi.Router) {
		r.Post("/add", h.AddTask())
		r.Get("/list", h.GetTaskList())
	}
}
