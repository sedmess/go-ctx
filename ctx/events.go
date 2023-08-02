package ctx

const (
	ePanic = iota
)

type event struct {
	kind    int
	payload any
}

type panicPayload struct {
	reason any
	stack  []byte
}
