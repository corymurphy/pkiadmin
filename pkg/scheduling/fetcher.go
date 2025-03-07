package scheduling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/repo"
)

type Fetcher struct {
	queue     *Queue
	stopFetch chan bool
	queries   *repo.Queries
	registry  *ProcRegistry
	*sync.WaitGroup
}

func NewFetcher(queue *Queue, registry *ProcRegistry) *Fetcher {
	return &Fetcher{
		queue,
		make(chan bool),
		queue.queries,
		registry,
		&sync.WaitGroup{},
	}
}

func (f *Fetcher) Run(jobs chan *RunableJob, workerReady chan bool) {
	f.Add(1)

	inprogress, err := f.queries.ListInProgressSet(context.Background())

	if err != nil {
		fmt.Println("error while getting inprogress jobs", err)
	}

	for _, job := range inprogress {
		f.queries.CreateSchedulerQueueJob(context.Background(), repo.CreateSchedulerQueueJobParams(job))
		f.queries.DeleteInProgressSet(context.Background(), job.ID)
	}

	go f.fetchJobs(jobs, workerReady)
}

func (f *Fetcher) Close() {
	f.stopFetch <- true
}

func (f *Fetcher) fetchJobs(jobs chan *RunableJob, workerReady chan bool) {
	for {
		select {
		case <-f.stopFetch:
			fmt.Println("Gracefully shutting fetcher")
			f.Done()
			return
		case <-workerReady:
			// TODO this should happen atomicly
			queue, err := f.queries.GetOneSchedulerQueueJob(context.Background())
			if err != nil {
				// TODO backoff?
				// fmt.Println("error while getting one job from the queue", err)
				time.Sleep(1 * time.Second)
				continue
			}
			_, err = f.queries.CreateInProgressSet(context.Background(), repo.CreateInProgressSetParams(queue))
			if err != nil {
				// TODO backoff?
				fmt.Println("error while transitioning the job to the inprogress queue", err)
				continue
			}
			err = f.queries.DeleteSchedulerQueueJob(context.Background(), queue.ID)
			if err != nil {
				// TODO backoff?
				fmt.Println("error while deleting job from the queue", err)
				continue
			}

			metadata := Metadata{
				Id:         queue.ID,
				Retry:      queue.Retry,
				RetryCount: queue.RetryCount,
				CreatedAt:  queue.CreatedAt,
			}
			job, err := f.registry.Get(queue.Processor, queue.Arguments, metadata)
			if err != nil {
				fmt.Println("error getting job from job registry", err)
				continue
			}

			fmt.Println("fetching job", job.Metadata().Id, "from fetcher")

			jobs <- &job

		}
	}
}
