package task

import (
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"perx/internal/domain/dto"
	"perx/internal/handlers/response"
	"perx/internal/lib/logger/sl"
	"perx/internal/lib/logger/with"
)

// @Summary		Add task
// @Tags			Task
// @Accept			json
// @Produce		json
// @Param		body 	body	dto.TaskDTO		true	"Request body"
// @Success		201			{object}	response.Response		"success response"
// @Failure		400,422			{object}	response.Response	"failure response"
// @Router			/task/add [post]
func (h *TaskHandler) AddTask() http.HandlerFunc {
	const op = "handlers.task.AddTask"

	return func(w http.ResponseWriter, r *http.Request) {
		log := with.WithOp(h.log, op)

		var model dto.TaskDTO
		if err := render.Decode(r, &model); err != nil {
			log.Error("failed to decode body", sl.Err(err))
			response.ErrorResponse(w, r, err, http.StatusBadRequest)
			return
		}

		if err := model.Validate(); err != nil {
			log.Error("task validation error", sl.Err(err))
			response.ErrorResponse(w, r, err, http.StatusUnprocessableEntity)
			return
		}

		log.Debug("new task request", slog.Any("model", model))

		if err := h.taskQueue.AddTask(&model); err != nil {
			log.Error("failed to add task", sl.Err(err))
			response.ErrorResponse(w, r, err, http.StatusBadRequest)
			return
		}

		log.Info("new task successfully created")
		response.SuccessResponse(w, r, http.StatusCreated, response.Response{Detail: "new task successfully created"})
	}
}

// @Summary		Get task list
// @Tags			Task
// @Accept			json
// @Produce		json
// @Success		200			{object}	[]models.TaskModel		"success response"
// @Failure		400			{object}	response.Response	"failure response"
// @Router			/task/list [get]
func (h *TaskHandler) GetTaskList() http.HandlerFunc {
	const op = "handlers.task.GetTasksList"

	return func(w http.ResponseWriter, r *http.Request) {
		log := with.WithOp(h.log, op)

		tasks, err := h.taskQueue.GetTaskList()
		if err != nil {
			log.Error("failed to get task list", sl.Err(err))
			response.ErrorResponse(w, r, err, http.StatusBadRequest)
			return
		}

		log.Info("task list successfully retrieved", slog.Int("count", len(tasks)))
		response.SuccessResponse(w, r, http.StatusOK, tasks)
	}
}
