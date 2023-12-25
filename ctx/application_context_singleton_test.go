package ctx

import (
	"testing"
)

func TestStartApplicationBlocking(t *testing.T) {
	type s1 struct {
	}
	type s2 struct {
		s1 *s1 `inject:""`
	}

	is1 := &s1{}
	is2 := &s2{}

	notifier := newLifecycleNotifier()

	application := CreateContextualizedApplication(PackageOf(is1, is2, notifier))

	go application.Join()

	<-notifier.ch

	if is2.s1 != is1 {
		t.Fail()
	}

	application.Stop()

	if <-notifier.ch != false {
		t.Fail()
	}

	application.Join()
}

func TestStartApplicationAsync(t *testing.T) {
	type s1 struct {
	}
	type s2 struct {
		s1 *s1 `inject:""`
	}

	is1 := &s1{}
	is2 := &s2{}

	application := CreateContextualizedApplication(PackageOf(is1, is2))

	if is2.s1 != is1 {
		t.Fail()
	}

	application.Stop().Join()
}

func TestApplicationRestart(t *testing.T) {
	type s1 struct {
		ctx AppContext `inject:"CTX"`
	}
	type s2 struct {
		s1 *s1 `inject:""`
	}

	is1 := &s1{}
	is2 := &s2{}

	application := CreateContextualizedApplication(PackageOf(is1, is2))

	code, name := is1.ctx.State()
	if code != 2 || name != "initialized" {
		t.Fail()
	}

	if is2.s1 != is1 {
		t.Fail()
	}

	application.Stop().Join()

	code, name = is1.ctx.State()
	if code != -1 || name != "used" {
		t.Fail()
	}

	oldCtx := is1.ctx

	application = CreateContextualizedApplication(PackageOf(is1, is2))

	if is1.ctx == oldCtx {
		t.Fail()
	}

	code, name = is1.ctx.State()
	if code != 2 || name != "initialized" {
		t.Fail()
	}

	application.Stop().Join()
}

type genericService[T any] struct {
	data T
}

func (g *genericService[T]) AfterStart() {
	println(g.data)
}

func (g *genericService[T]) BeforeStop() {
}

func Test_GenericService(t *testing.T) {
	app := CreateContextualizedApplication(PackageOf(&genericService[string]{data: "data"}))
	defer func() { app.Stop().Join() }()

	s1 := GetTypedService[*genericService[string]]()

	println(s1.data)
}
