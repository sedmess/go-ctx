package nopanic

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsPanicWrapperErrorDirectAndWrapped(t *testing.T) {
	direct := Run(func() { panic("boom") })
	if !IsPanicWrapperError(direct) {
		t.Fatal("direct captured panic was not recognized")
	}
	wrapped := fmt.Errorf("outer: %w", direct)
	if !IsPanicWrapperError(wrapped) {
		t.Fatal("wrapped captured panic was not recognized")
	}
}

func TestIsPanicWrapperErrorRejectsNilAndOrdinaryErrors(t *testing.T) {
	if IsPanicWrapperError(nil) {
		t.Fatal("nil classified as a panic wrapper")
	}
	if IsPanicWrapperError(errors.New("ordinary")) {
		t.Fatal("ordinary error classified as a panic wrapper")
	}
	if err := RunE(func() error { return errors.New("returned") }); IsPanicWrapperError(err) {
		t.Fatal("returned error classified as a panic wrapper")
	}
}

func TestRunResultCapturesPanic(t *testing.T) {
	value, err := RunResult(func() (int, error) { panic("result") })
	if value != 0 || !IsPanicWrapperError(err) {
		t.Fatalf("RunResult = (%d, %v)", value, err)
	}
}
