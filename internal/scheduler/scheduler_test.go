package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerTicks(t *testing.T) {
	var count int32
	s := New(10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}, nil)
	s.Start()
	time.Sleep(35 * time.Millisecond)
	s.Stop()
	if atomic.LoadInt32(&count) == 0 {
		t.Fatalf("expected at least one tick")
	}
}

func TestSchedulerReschedule(t *testing.T) {
	s := New(1*time.Second, func(ctx context.Context) error { return nil }, nil)
	if s.Interval() != 1*time.Second {
		t.Fatalf("unexpected interval")
	}
	s.Reschedule(2 * time.Second)
	if s.Interval() != 2*time.Second {
		t.Fatalf("expected rescheduled interval")
	}
}
