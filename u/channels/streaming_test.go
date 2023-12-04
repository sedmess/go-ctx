package channels

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestSingleElemChannel(t *testing.T) {
	slice, err := SingleElemChannel("data").Map(func(data string) any {
		return "mapped " + data
	}).CollectToSlice()
	if err != nil {
		t.FailNow()
	}
	if len(slice) != 1 {
		t.FailNow()
	}
	if slice[0] != "mapped data" {
		t.FailNow()
	}
}

func TestSingleElemChannelErr(t *testing.T) {
	slice, err := SingleElemChannelErr("data", nil).CollectToSlice()
	if err != nil {
		t.FailNow()
	}
	if len(slice) != 1 || slice[0] != "data" {
		t.FailNow()
	}
	slice, err = SingleElemChannelErr("data", errors.New("test error")).CollectToSlice()
	if err == nil {
		t.FailNow()
	}
	if len(slice) != 0 {
		t.FailNow()
	}
}

func TestSliceToChannel(t *testing.T) {
	i := 0
	expected := []int{2, 3, 4}
	err := Map(SliceToChannel([]int{1, 2, 3}), func(data int) int {
		return data + 1
	}).ForEachChanElem(func(data int) error {
		if expected[i] != data {
			t.FailNow()
		}
		i++
		if i == 3 {
			return errors.New("test error")
		} else {
			return nil
		}
	})
	if i != 3 {
		t.FailNow()
	}
	if err == nil || err.Error() != "test error" {
		t.FailNow()
	}
}

func TestCreateChannel(t *testing.T) {
	slice, err := CreateChannel(func(sink func(data string, context context.Context) bool) error {
		for i := 0; i < 10; i++ {
			if !sink(strconv.Itoa(i), context.Background()) {
				return BrokenSinkError()
			}
			<-time.After(10 * time.Millisecond)
		}
		return nil
	}).Map(func(data string) any {
		return data + data
	}).CollectToSlice()
	if err != nil {
		t.FailNow()
	}
	if len(slice) != 10 {
		t.FailNow()
	}
	for i := 0; i < 10; i++ {
		if slice[i] != strconv.Itoa(i)+strconv.Itoa(i) {
			t.FailNow()
		}
	}
}

func TestCreateChannelBuffered(t *testing.T) {
	slice, err := CreateChannelBuffered(5, func(sink func(data []string, context context.Context) bool) error {
		for i := 0; i < 10; i++ {
			if !sink([]string{strconv.Itoa(i)}, context.Background()) {
				return BrokenSinkError()
			}
			<-time.After(10 * time.Millisecond)
		}
		return nil
	}).Map(func(data string) any {
		return data + data
	}).CollectToSlice()
	if err != nil {
		t.FailNow()
	}
	if len(slice) != 10 {
		t.FailNow()
	}
	for i := 0; i < 10; i++ {
		if slice[i] != strconv.Itoa(i)+strconv.Itoa(i) {
			t.FailNow()
		}
	}
}

func TestFlapMap(t *testing.T) {
	slice, err := CreateChannel(func(sink func(data int, context context.Context) bool) error {
		for i := 0; i < 10; i++ {
			if !sink(i, context.Background()) {
				return BrokenSinkError()
			}
		}
		return nil
	}).FlatMap(func(data int) StreamingChan[any] {
		return CreateChannel(func(sink func(data any, context context.Context) bool) error {
			for i := 0; i < 5; i++ {
				if !sink(data+i, context.Background()) {
					return BrokenSinkError()
				}
			}
			return nil
		})
	}).CollectToSlice()
	if err != nil {
		t.FailNow()
	}
	if len(slice) != 50 {
		t.FailNow()
	}
}

func TestCreateChannelGeneratorError(t *testing.T) {
	count := 0
	err := CreateChannel(func(sink func(data int, context context.Context) bool) error {
		for i := 0; i < 5; i++ {
			if !sink(i, context.Background()) {
				return BrokenSinkError()
			}
		}
		return errors.New("test error 1")
	}).ForEachChanElem(func(data int) error {
		count++
		return nil
	})
	if err == nil || err.Error() != "test error 1" {
		t.FailNow()
	}
}

func TestCreateChannelContextTimeout(t *testing.T) {
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelFunc()

	count := 0

	err := CreateChannel(func(sink func(data int, context context.Context) bool) error {
		for i := 0; i < 10; i++ {
			if !sink(i, timeout) {
				return BrokenSinkError()
			}
		}
		return nil
	}).ForEachChanElem(func(data int) error {
		<-time.After(2 * time.Millisecond)
		count++
		return nil
	})

	if err == nil {
		t.FailNow()
	}
	if count == 10 {
		t.FailNow()
	}
}
