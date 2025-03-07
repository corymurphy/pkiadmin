package scheduling

import (
	"bytes"
	"fmt"
	"log"
)

type Worker struct {
	Id       int
	results  chan *Result
	closed   chan bool
	ready    bool
	busy     bool
	registry *ProcRegistry
}

func (w *Worker) Close() {
	w.closed <- true
}

func (w *Worker) IsReady() bool {
	return w.ready
}

func (w *Worker) IsBusy() bool {
	return w.busy
}

func (w *Worker) RegisteredJobs() []string {
	var jobs []string
	// jobs = append(jobs, "SignCertificate")
	// jobs = append(jobs, "RenewCertificate")
	// jobs = append(jobs, "InstallCertificate")
	// for jobName := range w {
	// 	jobNames = append(jobNames, jobName)
	// }
	return jobs
}

func (w *Worker) Run(jobs chan *RunableJob, workerReady chan bool) {
	fmt.Printf("worker %d starting...\n", w.Id)

	for {
		select {
		case <-w.closed:
			fmt.Printf("stopping worker %d\n", w.Id)
			return
		case workerReady <- true:

		case job := <-jobs:

			fmt.Println("running job", (*job).Metadata().Id, "on worker", w.Id)

			var logBuffer bytes.Buffer
			logger := log.New(&logBuffer, fmt.Sprintf("[Worker %d] ", w.Id), log.LstdFlags)

			next, err := (*job).Run(logger)
			if err != nil {
				logger.Println("error while running job", err)
			} else {
				logger.Println("job completed successfully")
			}

			result := &Result{
				err:       err,
				completed: err == nil,
				worker:    w,
				metadata:  (*job).Metadata(),
				output:    logBuffer.String(),
				args:      (*job).Args(),
				next:      next,
			}

			// TODO remove from inprogress set

			w.results <- result
		}
	}
}
