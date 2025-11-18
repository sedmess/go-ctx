package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/logger"
)

type serviceA struct {
	l logger.Logger `ctx:""`
}

func (s *serviceA) getVal() string {
	s.l.Info("call serviceA.getVal")
	return "A"
}

type serviceB struct {
	l logger.Logger `ctx:""`
	a *serviceA     `ctx:""`
}

func (s *serviceB) getAVal() string {
	s.l.Info("call serviceB.getAVal")
	return s.a.getVal()
}

type task struct {
	l logger.Logger `ctx:""`

	c string `env:"C"`

	b *serviceB `ctx:""`
}

func (t *task) Run() {
	t.l.Info("start task...")
	_ = ctx.GetEnv("A").AsString()
	t.l.Info("C =", t.c)
	t.l.Info(t.b.getAVal())
	t.l.Info("...task finished")
}

func main() {
	ctx.Run(new(task), ctx.PackageOf(new(serviceA), new(serviceB)))
}
