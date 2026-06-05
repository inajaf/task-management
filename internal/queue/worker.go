package queue

import (
	"fmt"
	"github.com/devenairevo/task-management/internal/contracts/queuer"
	"github.com/devenairevo/task-management/internal/contracts/tasker"
	"github.com/devenairevo/task-management/internal/types"
	"time"
)

func StartQueueWorker(channel queuer.Queuer, taskMng tasker.Tasker) {
	go func() {
		for {
			queuedTask, err := channel.Dequeue()
			if err != nil {
				// queue closed or error, wait a bit and retry
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if queuedTask == nil {
				return
			}

			// Describe the task to get its latest state
			t, err := taskMng.DescribeTask(queuedTask.ID)
			if err != nil {
				fmt.Printf("⚠️ Worker: task with ID %d not found in manager: %v\n", queuedTask.ID, err)
				channel.TaskDone()
				continue
			}

			// 1. Transition to Processing
			t.Status = types.Processing
			if err := taskMng.UpdateTask(t); err != nil {
				fmt.Printf("⚠️ Worker: failed to update task %d to Processing: %v\n", t.ID, err)
			} else {
				fmt.Printf("⚙️ Task with ID %d ('%s') started processing\n", t.ID, t.Name)
			}

			// 2. Simulate task execution
			time.Sleep(2 * time.Second)

			// 3. Transition to Done
			t.Status = types.Done
			if err := taskMng.UpdateTask(t); err != nil {
				fmt.Printf("⚠️ Worker: failed to update task %d to Done: %v\n", t.ID, err)
			} else {
				fmt.Printf("✅ Task with ID %d ('%s') finished processing\n", t.ID, t.Name)
			}

			// 4. Notify completion
			channel.TaskDone()
		}
	}()
}
