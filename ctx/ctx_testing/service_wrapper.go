package ctx_testing

type ServiceWrapper[T any] struct {
	name     string
	instance T
}
