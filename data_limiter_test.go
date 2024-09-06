package main

import (
	"sync"
	"testing"
	"time"
)

func TestNewDataLimiter(t *testing.T) {
	limit := int64(100)
	dl := NewDataLimiter(limit)

	if dl.limit != limit {
		t.Errorf("expected limit %d, got %d", limit, dl.limit)
	}

	if len(dl.usage) != 0 {
		t.Errorf("expected empty usage map, got size %d", len(dl.usage))
	}
}

func TestAddUsage(t *testing.T) {
	dl := NewDataLimiter(100)

	dl.AddUsage("user1", 50)

	if dl.GetUsage("user1") != 50 {
		t.Errorf("expected usage 50, got %d", dl.GetUsage("user1"))
	}
}

func TestConsumeUsage(t *testing.T) {
	dl := NewDataLimiter(100)

	reportChan := dl.ConsumeUsage()

	go func() {
		reportChan <- NewUsageReport("user1", 30)
		close(reportChan)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Wait some time for consumer to consume message
		time.Sleep(100 * time.Millisecond)
		defer wg.Done()
	}()

	wg.Wait()

	if dl.GetUsage("user1") != 30 {
		t.Errorf("expected usage 30, got %d", dl.GetUsage("user1"))
	}
}

func TestIsLimitReached(t *testing.T) {
	dl := NewDataLimiter(100)

	if dl.IsLimitReached("user1") {
		t.Errorf("expected limit not to be reached initially")
	}

	dl.AddUsage("user1", 80)
	if dl.IsLimitReached("user1") {
		t.Errorf("expected limit not to be reached with 80 usage")
	}

	dl.AddUsage("user1", 30) // 80 + 30 = 110, exceeding the limit
	if !dl.IsLimitReached("user1") {
		t.Errorf("expected limit to be reached with 110 usage")
	}
}

func TestConcurrentUsage(t *testing.T) {
	dl := NewDataLimiter(100)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dl.AddUsage("user1", 10)
		}()
	}

	wg.Wait()

	if dl.GetUsage("user1") != 100 {
		t.Errorf("expected usage 100, got %d", dl.GetUsage("user1"))
	}
}
