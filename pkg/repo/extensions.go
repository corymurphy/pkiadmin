package repo

const (
	Requested int64 = 0
	Pending   int64 = 1
	Approved  int64 = 2
	Rejected  int64 = 3
	Revoked   int64 = 4
	Issued    int64 = 5
	Failed    int64 = 6
	Completed int64 = 7
)

func (c *CertificatesAndHashAlgorithmPaginatedRow) StatusString() string {
	switch c.Status {
	case Requested:
		return "Requested"
	case Pending:
		return "Pending"
	case Approved:
		return "Approved"
	case Rejected:
		return "Rejected"
	case Revoked:
		return "Revoked"
	case Issued:
		return "Issued"
	case Failed:
		return "Failed"
	case Completed:
		return "Completed"
	default:
		return "Unknown"
	}
}

func (c *CertificatesAndHashAlgorithmPaginatedRow) StatusStyle() string {
	switch c.Status {
	case Requested:
		return "amber"
	case Pending:
		return "amber"
	case Approved:
		return "amber"
	case Rejected:
		return "red"
	case Revoked:
		return "red"
	case Issued:
		return "green"
	case Failed:
		return "red"
	case Completed:
		return "green"
	default:
		return "amber"
	}
}

func (s *SchedulerScheduledset) PrettyArgs() string {
	// json.Un
	// args, err := json.MarshalIndent(s.Arguments, "", "  ")
	// if err != nil {
	// 	return ""
	// }
	return string(s.Arguments)
}

func (s *SchedulerQueue) PrettyArgs() string {
	return string(s.Arguments)
}

func (s *SchedulerScheduledset) ProcessorKind() string {
	return s.Processor
}

func (s *SchedulerScheduledset) ProcessorArguments() []byte {
	return s.Arguments
}

func (s *SchedulerInprogressset) ProcessorKind() string {
	return s.Processor
}

func (s *SchedulerInprogressset) ProcessorArguments() []byte {
	return s.Arguments
}

func (s *SchedulerQueue) ProcessorKind() string {
	return s.Processor
}

func (s *SchedulerQueue) ProcessorArguments() []byte {
	return s.Arguments
}
