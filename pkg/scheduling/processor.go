package scheduling

import (
	"encoding/json"
	"fmt"
	"log"
)

type ProcArgs interface {
	Kind() string
}

type Processor[A ProcArgs] interface {
	Run(log *log.Logger, args A) (*Job, error)
}

// type ErrorProcessor struct{}

// func (e *ErrorProcessor) Run(log *log.Logger, args ErrorArguments) error {
// 	log.Println("attempting to do a thing")
// 	return fmt.Errorf("failed to do a thing")
// }

// type ErrorArguments struct{}

// func (e ErrorArguments) Kind() string { return "ErrorProcessor" }

// type HelloWorldProcessor struct{}

// func (h *HelloWorldProcessor) Run(log *log.Logger, args HelloWorldArguments) error {
// 	fmt.Println("started sleeping for 8 seconds")
// 	time.Sleep(8 * time.Second)
// 	fmt.Println(args.Greeting, "from processor")
// 	return nil
// }

// type HelloWorldArguments struct {
// 	Greeting string `json:"greeting"`
// }

// func (h HelloWorldArguments) Kind() string { return "HelloWorldProcessor" }

func NewProcessorRegistry() *ProcRegistry {
	registry := make(ProcRegistry)
	return &registry
}

type ProcessorFactory[a ProcArgs] func([]byte) (Processor[ProcArgs], ProcArgs, error)

type ProcessorRegistry map[string]ProcessorFactory[ProcArgs]

func (p *ProcessorRegistry) Get(name string, args []byte) (Processor[ProcArgs], ProcArgs, error) {
	return (*p)[name](args)
}

type runable[A ProcArgs] struct {
	args     ProcArgs
	proc     Processor[A]
	metadata Metadata
}

func (p *runable[A]) Metadata() Metadata {
	return p.metadata
}

func (p *runable[A]) Run(log *log.Logger) (*Job, error) {
	args, ok := p.args.(A)
	if !ok {
		return nil, fmt.Errorf("invalid argument type")
	}
	return p.proc.Run(log, args)
}

func (p *runable[A]) Args() ProcArgs {
	return p.args
}

type RunableJob interface {
	Run(log *log.Logger) (*Job, error)
	Metadata() Metadata
	Args() ProcArgs
}

type ProcFactory interface {
	CreateProcessor([]byte, Metadata) (RunableJob, error)
}

type procFactory[A ProcArgs] struct {
	proc Processor[A]
}

func (p *procFactory[A]) CreateProcessor(input []byte, metadata Metadata) (RunableJob, error) {
	var args A
	err := json.Unmarshal(input, &args)
	if err != nil {
		return nil, err
	}
	return &runable[A]{args: args, proc: p.proc, metadata: metadata}, nil
}

type ProcRegistry map[string]ProcFactory

func RegisterProcessor[A ProcArgs](reg *ProcRegistry, proc Processor[A]) {
	var args A
	(*reg)[args.Kind()] = &procFactory[A]{proc: proc} // TODO validate uniqueness
}

func (p *ProcRegistry) Get(name string, args []byte, metadata Metadata) (RunableJob, error) {
	return (*p)[name].CreateProcessor(args, metadata)
}
