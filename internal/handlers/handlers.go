package handlers

import (
	"github.com/go-chi/chi/v5"
	"log/slog"
	"perx/internal/handlers/task"
)

type Handlers struct {
	log       *slog.Logger
	taskQueue task.TaskQueue
}

func NewHandlers(log *slog.Logger, taskQueue task.TaskQueue) *Handlers {
	return &Handlers{
		log:       log,
		taskQueue: taskQueue,
	}
}

func (h *Handlers) ConnectHandlers(r *chi.Mux) {
	r.Route("/task", task.NewTaskHandler(h.log, h.taskQueue).ConnectHandler())
}
