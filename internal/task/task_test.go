package task

import (
	"fmt"
	"github.com/devenairevo/task-management/internal/types"
	"sync"
	"testing"
)

func TestLocalTaskManager_Operations(t *testing.T) {
	mgr := NewLocalTaskManager()

	// 1. Create a task
	params := &CreateTaskParams{
		UserID:   1,
		UserName: "Alice",
		TaskID:   100,
		Name:     "Test Task 1",
	}

	taskObj, err := mgr.CreateTask(params)
	if err != nil {
		t.Fatalf("unexpected error creating task: %v", err)
	}

	if taskObj.ID != 100 || taskObj.Name != "Test Task 1" || taskObj.Status != types.Created {
		t.Errorf("created task fields do not match params: %+v", taskObj)
	}

	// 2. Describe task
	described, err := mgr.DescribeTask(100)
	if err != nil {
		t.Fatalf("unexpected error describing task: %v", err)
	}
	if described.ID != 100 || described.Name != "Test Task 1" || described.Status != types.Created {
		t.Errorf("described task fields do not match: %+v", described)
	}

	// Verify that DescribeTask returns a copy, not a shared reference
	described.Name = "Mutated Name"
	freshCheck, _ := mgr.DescribeTask(100)
	if freshCheck.Name == "Mutated Name" {
		t.Error("DescribeTask returned a shared reference instead of a copy/clone")
	}

	// 3. Update task
	taskObj.Status = types.Processing
	taskObj.Name = "Test Task 1 Updated"
	err = mgr.UpdateTask(taskObj)
	if err != nil {
		t.Fatalf("unexpected error updating task: %v", err)
	}

	updated, err := mgr.DescribeTask(100)
	if err != nil {
		t.Fatalf("unexpected error describing updated task: %v", err)
	}
	if updated.Status != types.Processing || updated.Name != "Test Task 1 Updated" {
		t.Errorf("task fields were not updated: %+v", updated)
	}

	// 4. List tasks
	list, err := mgr.ListTasks()
	if err != nil {
		t.Fatalf("unexpected error listing tasks: %v", err)
	}
	if len(list) != 1 || list[0].ID != 100 {
		t.Errorf("expected 1 task with ID 100, got: %+v", list)
	}

	// 5. Get user tasks
	userTasks := mgr.GetUserTasks(1)
	if len(userTasks) != 1 || userTasks[0].ID != 100 {
		t.Errorf("expected 1 user task with ID 100, got: %+v", userTasks)
	}
}

func TestLocalTaskManager_Concurrency(t *testing.T) {
	mgr := NewLocalTaskManager()
	var wg sync.WaitGroup

	numGoroutines := 50
	wg.Add(numGoroutines)

	for i := 1; i <= numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			params := &CreateTaskParams{
				UserID:   id,
				UserName: fmt.Sprintf("User%d", id),
				TaskID:   id,
				Name:     fmt.Sprintf("Task %d", id),
			}
			_, _ = mgr.CreateTask(params)
			_, _ = mgr.ListTasks()
			_, _ = mgr.DescribeTask(id)
		}(i)
	}

	wg.Wait()

	list, err := mgr.ListTasks()
	if err != nil {
		t.Fatalf("unexpected error listing tasks: %v", err)
	}
	if len(list) != numGoroutines {
		t.Errorf("expected %d tasks, got %d", numGoroutines, len(list))
	}
}
