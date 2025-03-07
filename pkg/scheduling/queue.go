package scheduling

import (
	"context"
	"encoding/json"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/repo"
)

type Queue struct {
	// fetcher     *Fetcher
	// poller      *Poller
	// redisClient *redis.Client
	// name     string
	queries  *repo.Queries
	jobReady chan bool
}

func NewQueue(queries *repo.Queries) *Queue {
	return &Queue{
		jobReady: make(chan bool),
		queries:  queries,
	}
}

func (q *Queue) EnqueueJob(ctx context.Context, performAt time.Time, job *Job) (err error) {

	args, err := json.Marshal(job.Arguments)
	if err != nil {
		return err
	}

	if time.Until(performAt) > (1 * time.Second) {
		_, err := q.queries.CreateScheduledSet(ctx, repo.CreateScheduledSetParams{
			ID:         job.Id,
			Retry:      job.Retry,
			RetryCount: job.RetryCount,
			CreatedAt:  job.CreatedAt,
			EnqueuedAt: time.Now(),
			PerformAt:  performAt,
			Processor:  job.Arguments.Kind(),
			Arguments:  args,
		})
		return err
	}

	q.queries.CreateSchedulerQueueJob(ctx, repo.CreateSchedulerQueueJobParams{
		ID:         job.Id,
		Retry:      job.Retry,
		RetryCount: job.RetryCount,
		CreatedAt:  job.CreatedAt,
		EnqueuedAt: time.Now(),
		Processor:  job.Arguments.Kind(),
		Arguments:  args,
	})

	return nil
}
