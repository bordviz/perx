package queue

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"perx/internal/domain/dto"
	"perx/internal/domain/models"
	"perx/internal/lib/canceled"
	"perx/internal/lib/logger/sl"
	"perx/internal/lib/logger/with"
	"slices"
	"sync"
	"time"
)

type Queue struct {
	ctx           context.Context
	mu            sync.RWMutex
	log           *slog.Logger
	tasks         map[int]*models.TaskModel
	taskQueueList []int
	TaskQueue     chan int
	nextID        int
}

func NewQueue(ctx context.Context, log *slog.Logger) *Queue {
	return &Queue{
		ctx:           ctx,
		log:           log,
		tasks:         make(map[int]*models.TaskModel),
		taskQueueList: make([]int, 0, 100),
		TaskQueue:     make(chan int),
		nextID:        1,
	}
}

func (q *Queue) AddTask(task *dto.TaskDTO) error {
	const op = "queue.AddTask"
	log := with.WithOp(q.log, op)

	//TODO fix leak
	q.mu.Lock()
	defer q.mu.Unlock()

	if err := canceled.IsContextCanceled(q.ctx); err != nil {
		log.Error("failed to create task", sl.Err(err))
		return err
	}

	newTask := new(models.TaskModel)
	newTask.CreateTask(task, q.nextID)

	q.tasks[newTask.ID] = newTask
	q.taskQueueList = append(q.taskQueueList, newTask.ID)
	q.nextID++

	log.Debug("new task successfully created", slog.Any("task", newTask))
	log.Info("new task successfully created", slog.Int("task_id", newTask.ID))
	return nil
}

func (q *Queue) RemoveTaskFromList(taskID int) error {
	const op = "queue.RemoveTaskFromList"
	log := with.WithOp(q.log, op)

	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.taskQueueList) == 0 {
		return errors.New("task queue list is empty")
	}

	q.taskQueueList = q.taskQueueList[1:]
	log.Debug("task successfully removed from list", slog.Int("task_id", taskID))
	return nil
}

func (q *Queue) GetTask(id int) (*models.TaskModel, error) {
	const op = "queue.GetTask"
	log := with.WithOp(q.log, op)

	q.mu.RLock()
	defer q.mu.RUnlock()

	if err := canceled.IsContextCanceled(q.ctx); err != nil {
		log.Error("failed to get task", sl.Err(err))
		return nil, err
	}

	task, ok := q.tasks[id]
	if !ok {
		log.Error("task not found", slog.Int("task_id", id))
		return nil, errors.New("task not found")
	}

	return task, nil
}

func (q *Queue) GetTaskList() ([]*models.TaskModel, error) {
	const op = "queue.GetTaskList"
	log := with.WithOp(q.log, op)

	if err := canceled.IsContextCanceled(q.ctx); err != nil {
		log.Error("failed to get task list", sl.Err(err))
		return nil, err
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*models.TaskModel, 0, len(q.taskQueueList))
	keys := slices.Collect(maps.Keys(q.tasks))
	slices.Sort(keys)

	for _, key := range keys {
		task, ok := q.tasks[key]
		if !ok {
			log.Warn("task not found", slog.Int("task_id", key))
			continue
		}

		if idx := slices.Index(q.taskQueueList, task.ID); idx != -1 {
			task.QueueIndex = idx
		}

		tasks = append(tasks, task)
	}

	log.Debug("task list successfully retrieved", slog.Int("tasks_count", len(tasks)))
	return tasks, nil
}

func (q *Queue) DeleteTask(taskID int) {
	const op = "queue.DeleteTask"
	log := with.WithOp(q.log, op)

	if err := canceled.IsContextCanceled(q.ctx); err != nil {
		log.Warn("failed to delete task", sl.Err(err))
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.tasks, taskID)
	log.Debug("task successfully deleted", slog.Int("task_id", taskID))
}

func (q *Queue) StartQueue() {
	const op = "queue.StartQueue"
	log := with.WithOp(q.log, op)

	var lastSendTaskID int

	for {
		if len(q.taskQueueList) == 0 {
			log.Debug("queue has no tasks, stop cron for 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}

		var taskID int
		q.mu.RLock()
		if len(q.taskQueueList) != 0 {
			taskID = q.taskQueueList[0]
		}
		q.mu.RUnlock()

		if taskID == lastSendTaskID {
			continue
		}

		select {
		case <-q.ctx.Done():
			return
		case q.TaskQueue <- taskID:
			lastSendTaskID = taskID
			log.Debug("add new task to queue channel", slog.Int("task_id", taskID))
		}

	}
}
