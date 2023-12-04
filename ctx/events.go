package ctx

const (
	eUnhandledPanic = iota
	eSuppressedPanic
	eStop
)

type event struct {
	kind    int
	payload any
}

type panicPayload struct {
	reason any
	stack  []byte
}
