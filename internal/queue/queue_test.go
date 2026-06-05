package queue

import (
	"github.com/devenairevo/task-management/internal/task"
	"github.com/devenairevo/task-management/internal/types"
	"sync"
	"testing"
	"time"
)

func TestChannel_Operations(t *testing.T) {
	var wg sync.WaitGroup
	ch := NewChannel(5, &wg)

	// Verify empty queue state
	if !ch.IsEmpty() {
		t.Error("expected queue to be empty initially")
	}
	if ch.Size() != 0 {
		t.Errorf("expected queue size to be 0, got %d", ch.Size())
	}

	task1 := &task.Task{ID: 10, Name: "Task 10", Status: types.Created}

	// Enqueue
	err := ch.Enqueue(task1)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	if ch.IsEmpty() {
		t.Error("expected queue to not be empty after enqueue")
	}
	if ch.Size() != 1 {
		t.Errorf("expected queue size to be 1, got %d", ch.Size())
	}

	// Dequeue
	dequeued, err := ch.Dequeue()
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}
	if dequeued.ID != 10 {
		t.Errorf("expected dequeued task ID to be 10, got %d", dequeued.ID)
	}

	// Call TaskDone to decrement the wg added during Enqueue
	ch.TaskDone()

	// Wait with timeout to verify wg was decremented
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Error("wait group did not complete in time")
	}
}
