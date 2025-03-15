package certificates

type RequestTimelineStatus int64

const (
	Completed RequestTimelineStatus = 0
	Failed    RequestTimelineStatus = 1
	Pending   RequestTimelineStatus = 2
)

func (a RequestTimelineStatus) String() string {
	switch a {
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	case Pending:
		return "Pending"
	default:
		return "Unknown"
	}
}
