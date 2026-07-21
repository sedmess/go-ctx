package ctx_testing

import (
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/autoctx"
	"os"
	"strings"
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

func TestTestingPackagesAllowOneExplicitSubstitution(t *testing.T) {
	services := mergeServicePackages(
		[]ctx.ServicePackage{ctx.PackageOf(ctx.WithName("service", &BService{}))},
		[]ctx.ServicePackage{ctx.PackageOf(ctx.WithName("service", &aServiceStub{}))},
	)
	if len(services) != 1 {
		t.Fatalf("merged services = %d, want 1", len(services))
	}
}

func TestTestingPackagesRejectDuplicateBaseAndTestingEntries(t *testing.T) {
	tests := []struct {
		name    string
		base    []ctx.ServicePackage
		testing []ctx.ServicePackage
		want    string
	}{
		{
			name: "base",
			base: []ctx.ServicePackage{ctx.PackageOf(
				ctx.WithName("same", &BService{}),
				ctx.WithName("same", &BService{}),
			)},
			want: "duplicate base",
		},
		{
			name: "testing",
			base: []ctx.ServicePackage{ctx.PackageOf(ctx.WithName("same", &BService{}))},
			testing: []ctx.ServicePackage{
				ctx.PackageOf(ctx.WithName("same", &aServiceStub{})),
				ctx.PackageOf(ctx.WithName("same", &aServiceStub{})),
			},
			want: "duplicate testing",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recovered == nil || !strings.Contains(recovered.(string), test.want) {
					t.Fatalf("panic = %v, want containing %q", recovered, test.want)
				}
			}()
			mergeServicePackages(test.base, test.testing)
		})
	}
}

func TestParameterOverridesRestoreEnvironmentPresenceAndValue(t *testing.T) {
	const key = "GO_CTX_TEST_RESTORE"
	tests := []struct {
		name    string
		present bool
		value   string
	}{
		{name: "absent"},
		{name: "present empty", present: true},
		{name: "present value", present: true, value: "original"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.present {
				if err := os.Setenv(key, test.value); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Unsetenv(key); err != nil {
				t.Fatal(err)
			}

			if code := withParameters(map[string]string{key: "override"}, func() int {
				value, present := os.LookupEnv(key)
				if !present || value != "override" {
					t.Fatalf("override = (%q, %t)", value, present)
				}
				return 17
			}); code != 17 {
				t.Fatalf("return code = %d", code)
			}

			value, present := os.LookupEnv(key)
			if present != test.present || value != test.value {
				t.Fatalf("restored = (%q, %t), want (%q, %t)", value, present, test.value, test.present)
			}
		})
	}
	_ = os.Unsetenv(key)
}
