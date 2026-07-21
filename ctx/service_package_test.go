package ctx

import (
	"strings"
	"testing"
)

type duplicateService struct{}
type otherDuplicateService struct{}

func registerPackageForTest(context *appContext, pkg ServicePackage) error {
	var result error
	pkg.ForEach(func(service any, name string) {
		if result == nil {
			result = context.register(service, name)
		}
	})
	return result
}

func TestServicePackagePreservesDuplicateEntriesForValidation(t *testing.T) {
	first := &duplicateService{}
	second := &otherDuplicateService{}
	pkg := PackageOf(WithName("same", first), WithName("same", second))

	var entries []any
	pkg.ForEach(func(service any, _ string) { entries = append(entries, service) })
	if len(entries) != 2 || entries[0] != first || entries[1] != second {
		t.Fatalf("package entries = %v; duplicates/order were not preserved", entries)
	}

	context := newApplicationContext()
	if err := registerPackageForTest(context, pkg); err == nil || !strings.Contains(err.Error(), "duplication") {
		t.Fatalf("duplicate registration error = %v", err)
	}
}

func TestServiceRegistrationRejectsDerivedCrossPackageMixedAndReservedNames(t *testing.T) {
	tests := []struct {
		name     string
		packages []ServicePackage
		want     string
	}{
		{
			name:     "derived",
			packages: []ServicePackage{PackageOf(&duplicateService{}, &duplicateService{})},
			want:     "duplication",
		},
		{
			name: "cross package",
			packages: []ServicePackage{
				PackageOf(WithName("same", &duplicateService{})),
				PackageOf(WithName("same", &otherDuplicateService{})),
			},
			want: "duplication",
		},
		{
			name: "mixed",
			packages: []ServicePackage{
				PackageOf(&duplicateService{}),
				PackageOf(WithName("*ctx.duplicateService", &otherDuplicateService{})),
			},
			want: "duplication",
		},
		{
			name:     "reserved",
			packages: []ServicePackage{PackageOf(WithName(ctxTag, &duplicateService{}))},
			want:     "reserved",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := newApplicationContext()
			var err error
			for _, pkg := range test.packages {
				if err = registerPackageForTest(context, pkg); err != nil {
					break
				}
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("registration error = %v, want containing %q", err, test.want)
			}
		})
	}
}
