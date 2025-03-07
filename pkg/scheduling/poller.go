package scheduling

import (
	"context"
	"sync"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/repo"
)

type Poller struct {
	queue    *Queue
	queries  *repo.Queries
	registry *ProcRegistry
	stop     chan bool
	*sync.WaitGroup
}

func NewPoller(queue *Queue, queries *repo.Queries, registry *ProcRegistry) *Poller {
	return &Poller{
		queue,
		queries,
		registry,
		make(chan bool),
		&sync.WaitGroup{},
	}
}

func (p *Poller) Run() {
	p.Add(1)
	go p.poll()
}

func (p *Poller) Close() {
	p.stop <- true
}

func (p *Poller) poll() {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-p.stop:
			p.Done()
			return
		case <-timer.C:
			scheduled, err := p.queries.ListScheduledSetShouldPerform(context.Background())
			if err != nil {
				timer.Reset(5 * time.Second)
				continue
			}

			for _, s := range scheduled {
				if time.Until(s.PerformAt) > (1 * time.Second) {
					continue
				}

				// job := &Job[ProcArgs]{
				// 	Id:         s.ID,
				// 	Retry:      s.Retry,
				// 	RetryCount: s.RetryCount,
				// 	CreatedAt:  s.CreatedAt,
				// }

				_, err = p.queries.CreateSchedulerQueueJob(context.Background(), repo.CreateSchedulerQueueJobParams{
					ID:         s.ID,
					Retry:      s.Retry,
					RetryCount: s.RetryCount,
					CreatedAt:  s.CreatedAt,
					EnqueuedAt: time.Now(),
					Arguments:  s.Arguments,
					Processor:  s.Processor,
				})
				if err != nil {
					continue
				}

				err = p.queries.DeleteScheduledSet(context.Background(), s.ID)
				if err != nil {
					continue
				}
			}

			timer.Reset(5 * time.Second)
		}
	}
}
