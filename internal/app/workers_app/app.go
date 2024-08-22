package workers_app

import "context"

type Worker interface {
	Run(ctx context.Context) error
}

type WorkersApp struct {
	ctx     context.Context
	workers []Worker
}

func New() *WorkersApp {
	return &WorkersApp{
		ctx:     context.Background(),
		workers: make([]Worker, 0),
	}
}

func (app *WorkersApp) AddWorker(worker Worker) {
	app.workers = append(app.workers, worker)
}

func (app *WorkersApp) Run() error {

	for _, worker := range app.workers {
		if err := worker.Run(app.ctx); err != nil {
			app.Stop()
			return err
		}
	}

	return nil
}

func (app *WorkersApp) Stop() {
	app.ctx.Done()
}
