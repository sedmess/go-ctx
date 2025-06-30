package ctx_testing

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/autoctx"
	"os"
	"testing"
)

func init() {
	autoctx.S(&aService{data: "test_a_value"})
	autoctx.S(ctx.Typed[BService]())
}

type AService interface {
	Data() string
}

type aService struct {
	AService `ctx:""`
	data     string
}

func (instance *aService) Data() string {
	return instance.data
}

type BService struct {
	aService AService `ctx:""`
	value    string   `env:"TEST_ENV"`
}

func (instance *BService) Data() string {
	return instance.aService.Data()
}

type aServiceStub struct {
}

func (stub *aServiceStub) Data() string {
	return "stub"
}

func TestMain(m *testing.M) {
	_ = os.Setenv("SLOG_HANDLER", "legacy")
	_ = os.Setenv("SLOG_LEVEL", "debug")
	os.Exit(
		CreateAutoTestingApplication().
			WithParameter("TEST_ENV", "test_env_value").
			WithTestingService(Instead[AService](&aServiceStub{})).
			Run(m.Run),
	)
}

func Test_AService(t *testing.T) {
	aService, found := ctx.GetTypedService[AService]()
	if !found {
		t.Fail()
	}
	if aService.Data() != "stub" {
		t.Fail()
	}
}

func Test_BService(t *testing.T) {
	bService, found := ctx.GetTypedService[*BService]()
	if !found {
		t.Fail()
	}
	if bService.Data() != "stub" {
		t.Fail()
	}
	if bService.value != "test_env_value" {
		t.Fail()
	}
}
