package workers_app

import (
	"context"
)

type Worker interface {
	Run(ctx context.Context) error
	Stop()
}

type WorkersApp struct {
	workers []Worker
}

func New() *WorkersApp {
	return &WorkersApp{
		workers: make([]Worker, 0),
	}
}

func (app *WorkersApp) AddWorker(worker Worker) {
	app.workers = append(app.workers, worker)
}

func (app *WorkersApp) Run() error {
	for _, worker := range app.workers {
		if err := worker.Run(context.Background()); err != nil {
			app.Stop()
			return err
		}
	}

	return nil
}

func (app *WorkersApp) Stop() {
	for _, worker := range app.workers {
		worker.Stop()
	}
}
