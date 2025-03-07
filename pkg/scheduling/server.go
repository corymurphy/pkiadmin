package scheduling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/repo"
	"github.com/google/uuid"
)

type Metadata struct {
	Id         uuid.UUID `json:"id"`
	Retry      bool      `json:"retry"`
	RetryCount int64     `json:"retryCount"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Job struct {
	Id         uuid.UUID `json:"id"`
	Retry      bool      `json:"retry"`
	RetryCount int64     `json:"retryCount"`
	CreatedAt  time.Time `json:"createdAt"`
	Metadata   Metadata  `json:"metadata"`
	Arguments  ProcArgs  `json:"arguments"`
}

func (j *Job) PrettyArgs() string {
	args, err := json.MarshalIndent(j.Arguments, "", "")
	if err != nil {
		return ""
	}
	return string(args)
}

type Result struct {
	worker    *Worker
	metadata  Metadata
	completed bool
	err       error
	output    string
	args      ProcArgs
	next      *Job
}

type Server struct {
	running bool
	workers []*Worker
	results chan *Result
	poller  *Poller
	fetcher *Fetcher
	jobs    chan *RunableJob
	queue   *Queue

	processors *ProcRegistry
}

func (s *Server) Workers() []*Worker {
	return s.workers
}

func (s *Server) GetProcessor(name string, args []byte, metadata Metadata) (RunableJob, error) {
	return s.processors.Get(name, args, metadata)
}

func New(options ...func(*Server)) *Server {
	s := &Server{
		running:    false,
		processors: NewProcessorRegistry(),
		workers:    []*Worker{},
		results:    make(chan *Result),
	}

	// TODO: enforce default options
	for _, option := range options {
		option(s)
	}

	s.poller = NewPoller(s.queue, s.queue.queries, s.processors)
	s.fetcher = NewFetcher(s.queue, s.processors)

	return s
}

func WithWorkerCount(workerCount int) func(*Server) {
	return func(s *Server) {
		for i := 0; i < workerCount; i++ {
			s.workers = append(s.workers, &Worker{
				Id:       i,
				results:  s.results,
				closed:   make(chan bool),
				registry: s.processors,
			})
		}
	}
}

func WithWorkers(workers []*Worker) func(*Server) {
	return func(s *Server) {
		s.workers = workers
	}
}

func WithQueue(queue *Queue) func(*Server) {
	return func(s *Server) {
		s.queue = queue
	}
}

func WithProcessor[A ProcArgs](proc Processor[A]) func(*Server) {
	return func(s *Server) {
		RegisterProcessor(s.processors, proc)
	}
}

// func NewServer(workerCount int, queue *Queue) *Server {

// 	registry := NewProcessorRegistry()
// 	RegisterProcessor(registry, &HelloWorldProcessor{})
// 	RegisterProcessor(registry, &ErrorProcessor{})
// 	RegisterProcessor(registry, &certificates.CreateRsaCsrJob{})

// 	workers := []*Worker{}
// 	results := make(chan *Result)
// 	for i := 0; i < workerCount; i++ {
// 		workers = append(workers, &Worker{
// 			Id:       i,
// 			results:  results,
// 			closed:   make(chan bool),
// 			registry: registry,
// 		})
// 	}

// 	return &Server{
// 		running:    false,
// 		workers:    workers,
// 		results:    results,
// 		processors: registry,
// 		poller:     NewPoller(queue, queue.queries, registry),
// 		fetcher:    NewFetcher(queue, registry),
// 		queue:      queue,
// 	}
// }

func (s *Server) ProcessNow(job *RunableJob) {
	fmt.Println("processing job immediately")
	s.jobs <- job
}

func (s *Server) Run() error {
	if s.running {
		return errors.New("server is already running")
	}
	fmt.Println("⇨ starting scheduling server with", len(s.workers), "workers")

	s.jobs = make(chan *RunableJob)
	workerReady := make(chan bool)

	s.running = true
	s.poller.Run()
	s.fetcher.Run(s.jobs, workerReady)

	for _, worker := range s.workers {
		go worker.Run(s.jobs, workerReady)
	}

	for result := range s.results {

		ctx := context.Background()
		// TODO this is where the retry and error handling should be
		// s.processResult(result)

		s.queue.queries.DeleteInProgressSet(ctx, result.metadata.Id)

		if result.err == nil {
			// s.queue.queries.DeleteSchedulerQueueJob(context.Background(), result.metadata.Id)
			_, err := s.queue.queries.CreateCompletedJob(ctx, repo.CreateCompletedJobParams{
				ID:         result.metadata.Id,
				Retry:      result.metadata.Retry,
				RetryCount: result.metadata.RetryCount,
				CreatedAt:  result.metadata.CreatedAt,
				Processor:  result.args.Kind(),
				Log:        result.output,
			})
			if err != nil {
				fmt.Println("error creating completed job", err)
			}
			if result.next == nil {
				continue
			}
			if err := s.queue.EnqueueJob(ctx, time.Now(), result.next); err != nil {
				fmt.Println("error enqueuing next job", err)
			}
			continue
		}

		if !result.metadata.Retry || result.metadata.RetryCount >= 4 {
			_, err := s.queue.queries.CreateFailedJob(ctx, repo.CreateFailedJobParams{
				ID:         result.metadata.Id,
				Retry:      result.metadata.Retry,
				RetryCount: result.metadata.RetryCount,
				CreatedAt:  result.metadata.CreatedAt,
				Processor:  result.args.Kind(),
				Log:        result.output,
			})
			if err != nil {
				fmt.Println("error creating failed job", err)
			}
			continue
		}

		err := s.queue.EnqueueJob(
			ctx,
			time.Now().Add(10*time.Second),
			&Job{
				Id:         result.metadata.Id,
				Retry:      result.metadata.Retry,
				RetryCount: result.metadata.RetryCount + 1,
				CreatedAt:  result.metadata.CreatedAt,
				Metadata:   result.metadata,
				Arguments:  result.args,
			})
		if err != nil {
			fmt.Println("error enqueuing job", err)
		}

	}

	return nil
}

func (s *Server) Close() (err error) {

	fmt.Println("⇨ scheduling server is done")

	go s.poller.Close()
	s.poller.Wait()

	go s.fetcher.Close()
	s.fetcher.Wait()

	s.running = false
	close(s.results)
	for _, w := range s.workers {
		w.Close()
	}

	return err
}

// func (s *Server) processResult(result *Result) {
// 	fmt.Println("processing result")
// }
