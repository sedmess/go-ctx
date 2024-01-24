package ctx_testing

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/u"
	"os"
	"testing"
)

type AService interface {
	Data() string
}

type aService struct {
	AService `implement:""`
	data     string
}

func (instance *aService) Data() string {
	return instance.data
}

type BService struct {
	aService AService `inject:""`
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
	os.Exit(
		CreateTestingApplication(ctx.PackageOf(
			&aService{data: "test_a_value"},
			&BService{},
		)).
			WithParameter("TEST_ENV", "test_env_value").
			WithTestingService(Instead[AService](&aServiceStub{})).
			Run(m.Run),
	)
}

func Test_AService(t *testing.T) {
	aService := ctx.GetService(u.GetInterfaceName[AService]()).(AService)
	if aService.Data() != "stub" {
		t.Fail()
	}
}

func Test_BService(t *testing.T) {
	bService := ctx.GetService(u.GetInterfaceName[*BService]()).(*BService)
	if bService.Data() != "stub" {
		t.Fail()
	}
	if bService.value != "test_env_value" {
		t.Fail()
	}
}
