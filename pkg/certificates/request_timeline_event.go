package certificates

type RequestTimelineEvent int64

const (
	Requested RequestTimelineEvent = 0
	Approved  RequestTimelineEvent = 1
	Generated RequestTimelineEvent = 2
	Submitted RequestTimelineEvent = 3
	Issued    RequestTimelineEvent = 4
)

func (a RequestTimelineEvent) String() string {
	switch a {
	case Requested:
		return "Requested"
	case Approved:
		return "Approved"
	case Generated:
		return "Generated"
	case Submitted:
		return "Submitted"
	case Issued:
		return "Issued"
	default:
		return "Unknown"
	}
}
