package channels_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/sedmess/go-ctx/utils/channels"
)

func TestToChanReturnsReceiveOnlyOrderedOutputs(t *testing.T) {
	var values <-chan int
	var failures <-chan error
	values, failures = channels.SliceToChannel([]int{1, 2, 3}).ToChan(1)

	var got []int
	for value := range values {
		got = append(got, value)
		time.Sleep(time.Millisecond)
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("values = %v", got)
	}
	if failure, open := <-failures; open || failure != nil {
		t.Fatalf("success error channel = (%v, open=%t)", failure, open)
	}
}

func TestToChanDeliversSourceErrorAndCloses(t *testing.T) {
	want := errors.New("stream failed")
	stream := channels.CreateChannel(func(sink func(int, context.Context) bool) error {
		if !sink(7, context.Background()) {
			t.Fatal("sink rejected value")
		}
		return want
	})
	values, failures := stream.ToChan(0)
	if got := <-values; got != 7 {
		t.Fatalf("value = %d, want 7", got)
	}
	if _, open := <-values; open {
		t.Fatal("value channel remained open")
	}
	if got := <-failures; !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
	if _, open := <-failures; open {
		t.Fatal("error channel remained open")
	}
}
