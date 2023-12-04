package main

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/ctxtesting"
	"github.com/sedmess/go-ctx/logger"
	"github.com/sedmess/go-ctx/u"
	"testing"
)

func TestStartContextualizedApplication(t *testing.T) {
	type service1 struct {
		l logger.Logger `logger:""`
	}

	type service2 struct {
		l        logger.Logger `logger:""`
		service1 *service1     `inject:""`
	}

	s1 := &service1{}
	s2 := &service2{}

	application := ctxtesting.TestingContextualizedApplication().WithServices(s1, s2)

	application.Start()
	defer application.Stop()

	if ctx.GetService(u.GetInterfaceName[*service1]()) != s1 {
		t.Fail()
	}
	if ctx.GetService(u.GetInterfaceName[*service2]()) != s2 {
		t.Fail()
	}
	if ctx.GetService(u.GetInterfaceName[*service2]()).(*service2).service1 != s1 {
		t.Fail()
	}
	if s1.l == nil {
		t.Fail()
	}
	if s2.l == nil {
		t.Fail()
	}
}
