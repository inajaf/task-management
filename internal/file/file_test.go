package file

import (
	"github.com/devenairevo/task-management/internal/task"
	"github.com/devenairevo/task-management/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestFileManager_Operations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "task_manager_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgr := NewManager(tempDir)

	// Test 1: List empty directory
	_, err = mgr.ListTasks()
	if err == nil {
		t.Error("expected error listing from empty directory, got nil")
	}

	// Test 2: Create task
	params := &task.CreateTaskParams{
		UserID:   1,
		UserName: "Alice",
		TaskID:   200,
		Name:     "FileTask",
	}

	created, err := mgr.CreateTask(params)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if created.ID != 200 || created.Name != "FileTask" || created.Status != types.Created {
		t.Errorf("incorrect fields on created task: %+v", created)
	}

	// Verify file was written
	filename := "task_FileTask_200.json"
	filePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected task JSON file to exist at %s, but it was not found", filePath)
	}

	// Test 3: Describe task
	described, err := mgr.DescribeTask(200)
	if err != nil {
		t.Fatalf("failed to describe task: %v", err)
	}
	if described.ID != 200 || described.Name != "FileTask" || described.Status != types.Created {
		t.Errorf("incorrect fields on described task: %+v", described)
	}

	// Test 4: Update task status & name
	described.Status = types.Processing
	described.Name = "FileTaskUpdated"
	err = mgr.UpdateTask(described)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	updated, err := mgr.DescribeTask(200)
	if err != nil {
		t.Fatalf("failed to describe updated task: %v", err)
	}
	if updated.Status != types.Processing || updated.Name != "FileTaskUpdated" {
		t.Errorf("incorrect fields on updated task: %+v", updated)
	}

	// Test 5: List tasks
	list, err := mgr.ListTasks()
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(list) != 1 || list[0].ID != 200 {
		t.Errorf("expected 1 task with ID 200, got: %+v", list)
	}
}
