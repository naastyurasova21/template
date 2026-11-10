package trip

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

func (s Status) String() string {
	return string(s)
}
