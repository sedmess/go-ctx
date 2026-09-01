package channels_test

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sedmess/go-ctx/utils/channels"
)

func TestMapMethodPreservesResultTypeAndOrder(t *testing.T) {
	var mapped channels.StreamingChan[string] = channels.SliceToChannel([]int{1, 2, 3}).Map(strconv.Itoa)

	got, err := mapped.CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"1", "2", "3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("mapped values = %v, want %v", got, want)
	}
}

func TestFlatMapMethodPreservesResultTypeAndOrder(t *testing.T) {
	var mapped channels.StreamingChan[string] = channels.SliceToChannel([]int{1, 2}).FlatMap(
		func(value int) channels.StreamingChan[string] {
			return channels.SliceToChannel([]string{strconv.Itoa(value), strconv.Itoa(value * 10)})
		},
	)

	got, err := mapped.CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"1", "10", "2", "20"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("flat-mapped values = %v, want %v", got, want)
	}
}

func TestGenericMethodValuesAndExpressions(t *testing.T) {
	explicitSource := channels.SliceToChannel([]int{4})
	contextualSource := channels.SliceToChannel([]int{5})
	mapValue := explicitSource.Map[string]
	var contextualValue func(func(int) string) channels.StreamingChan[string] = contextualSource.Map
	mapExpression := channels.StreamingChan[int].Map[string]

	for name, mapped := range map[string]channels.StreamingChan[string]{
		"explicit value":      mapValue(strconv.Itoa),
		"contextual value":    contextualValue(strconv.Itoa),
		"explicit expression": mapExpression(channels.SliceToChannel([]int{6}), strconv.Itoa),
	} {
		got, err := mapped.CollectToSlice()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(got) != 1 {
			t.Fatalf("%s: values = %v", name, got)
		}
	}
}

func TestMapMethodSupportsAnyAndEmptyStreams(t *testing.T) {
	var anyStream channels.StreamingChan[any] = channels.SliceToChannel([]int{7}).Map(
		func(value int) any { return strconv.Itoa(value) },
	)
	gotAny, err := anyStream.CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	if want := []any{"7"}; !reflect.DeepEqual(gotAny, want) {
		t.Fatalf("any values = %v, want %v", gotAny, want)
	}

	empty, err := channels.SliceToChannel([]int{}).Map(strconv.Itoa).CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty map values = %v", empty)
	}
}

func TestMapAndFlatMapPropagateErrors(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		want := errors.New("source failed")
		_, got := channels.SingleElemChannelErr(0, want).Map(strconv.Itoa).CollectToSlice()
		if !errors.Is(got, want) {
			t.Fatalf("error = %v, want %v", got, want)
		}
	})

	t.Run("nested", func(t *testing.T) {
		want := errors.New("nested failed")
		_, got := channels.SingleElemChannel(1).FlatMap(
			func(int) channels.StreamingChan[string] {
				return channels.SingleElemChannelErr("", want)
			},
		).CollectToSlice()
		if !errors.Is(got, want) {
			t.Fatalf("error = %v, want %v", got, want)
		}
	})
}

func TestFlatMapPackageNamesAreEquivalent(t *testing.T) {
	mapper := func(value int) channels.StreamingChan[string] {
		return channels.SingleElemChannel(strconv.Itoa(value))
	}
	canonical, err := channels.FlatMap(channels.SliceToChannel([]int{1, 2}), mapper).CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := channels.FlapMap(channels.SliceToChannel([]int{1, 2}), mapper).CollectToSlice()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(canonical, legacy) {
		t.Fatalf("FlatMap = %v, FlapMap = %v", canonical, legacy)
	}
}

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
