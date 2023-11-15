package ctx

const (
	eUnhandledPanic = iota
	eSuppressedPanic
)

type event struct {
	kind    int
	payload any
}

type panicPayload struct {
	reason any
	stack  []byte
}
