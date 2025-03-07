package repo

const (
	Pending  int64 = 1
	Approved int64 = 2
	Rejected int64 = 3
	Revoked  int64 = 4
	Issued   int64 = 5
)

func (c *CertificateRequestsAndHashAlgorithmRow) StatusString() string {
	switch c.Status.Int64 {
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
	default:
		return "Unknown"
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
