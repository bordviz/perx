package worker

import (
	"context"
	"log/slog"
	"perx/internal/domain/models"
	"perx/internal/lib/canceled"
	"perx/internal/lib/logger/sl"
	"perx/internal/lib/logger/with"
	"perx/internal/queue"
	"sync"
	"time"
)

type Pool struct {
	ctx             context.Context
	stopWorkersFunc context.CancelFunc
	wg              sync.WaitGroup
	log             *slog.Logger
	workersCount    int
	queue           *queue.Queue
}

func NewWorkerPool(ctx context.Context, log *slog.Logger, workersCount int, queue *queue.Queue) *Pool {
	return &Pool{
		ctx:          ctx,
		log:          log,
		workersCount: workersCount,
		queue:        queue,
	}
}

func (w *Pool) StartPool() {
	const op = "worker.StartPool"
	log := with.WithOp(w.log, op)

	ctx, cancel := context.WithCancel(w.ctx)
	w.stopWorkersFunc = cancel

	for idx := range w.workersCount {
		log.Debug("starting worker", slog.Int("worker_id", idx))
		w.wg.Go(func() {
			defer log.Debug("worker stopped", slog.Int("worker_id", idx))
			w.worker(ctx)
		})
	}
}

func (w *Pool) StopPool() {
	const op = "worker.StopPool"
	log := with.WithOp(w.log, op)

	w.stopWorkersFunc()
	w.wg.Wait()
	log.Debug("all workers was stopped", slog.Int("count", w.workersCount))
}

func (w *Pool) worker(ctx context.Context) {
	const op = "worker.worker"
	log := with.WithOp(w.log, op)

	for {
		select {
		case <-ctx.Done():
			log.Warn("stop worker by context", sl.Err(ctx.Err()))
			return
		case taskID, ok := <-w.queue.TaskQueue:
			if !ok {
				log.Error("stop worker by queue channel closed")
				return
			}

			log.Debug("worker received task", slog.Int("task_id", taskID))
			task, err := w.queue.GetTask(taskID)
			if err != nil {
				log.Warn("failed to get task", sl.Err(err))

				_ = w.queue.RemoveTaskFromList(taskID)
				continue
			}

			if err := w.completeTask(task); err != nil {
				log.Error("failed to complete task", sl.Err(err))
				_ = w.queue.RemoveTaskFromList(task.ID)
				continue
			}
		}
	}
}

func (w *Pool) completeTask(task *models.TaskModel) error {
	const op = "worker.completeTask"
	log := with.WithOp(w.log, op)

	task.SetStatusInProgress()
	task.QueueIndex = 0

	if err := w.queue.RemoveTaskFromList(task.ID); err != nil {
		log.Error("failed to remove task from queue list", sl.Err(err))
		return err
	}

	currentValue := task.StartValue
	currentIter := 0

	for currentIter < task.ElementsNumber {
		if err := canceled.IsContextCanceled(w.ctx); err != nil {
			log.Error("failed to complete task", sl.Err(err))
			return err
		}

		currentValue += task.Delta
		currentIter++

		task.CurrentIteration = currentIter
		task.CurrentValue = currentValue

		if currentIter < task.ElementsNumber {
			time.Sleep(time.Duration(task.IterationInterval) * time.Second)
		}
	}

	task.SetStatusCompleted()

	go func() {
		time.Sleep(time.Duration(task.TTL) * time.Second)
		w.queue.DeleteTask(task.ID)
	}()

	return nil
}
