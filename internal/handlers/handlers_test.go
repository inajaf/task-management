package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/devenairevo/task-management/internal/queue"
	"github.com/devenairevo/task-management/internal/task"
	"github.com/devenairevo/task-management/internal/types"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHTTPHandlers_Workflow(t *testing.T) {
	// Setup dependencies
	wg := &sync.WaitGroup{}
	taskMng := task.NewLocalTaskManager()
	channel := queue.NewChannel(10, wg)

	// Start queue worker (in goroutine)
	queue.StartQueueWorker(channel, taskMng)

	// Helper mux for testing multiple routes
	mux := http.NewServeMux()
	taskHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTasksHandler(taskMng, channel)(w, r)
		case http.MethodPost:
			PostTasksHandler(taskMng, channel, wg)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/task", taskHandler)
	mux.HandleFunc("/task/create", taskHandler)
	mux.HandleFunc("/tasks", GetTasksHandler(taskMng, channel))
	mux.HandleFunc("/tasks/", TaskByID(taskMng, channel))

	// 1. GET /tasks initially (should return empty list)
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 OK on empty list, got %d", rr.Code)
	}
	var tasks []*task.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}

	// 2. POST /task with custom task JSON
	customTaskReq := CreateTaskRequest{
		UserID:   42,
		UserName: "John",
		TaskID:   999,
		Name:     "CustomTaskName",
	}
	bodyBytes, _ := json.Marshal(customTaskReq)
	req = httptest.NewRequest(http.MethodPost, "/task", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", rr.Code)
	}

	var createdTasks []*task.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &createdTasks); err != nil {
		t.Fatalf("failed to decode created tasks: %v", err)
	}
	if len(createdTasks) != 1 || createdTasks[0].ID != 999 || createdTasks[0].Name != "CustomTaskName" {
		t.Errorf("created task fields do not match request: %+v", createdTasks)
	}

	// 3. GET /tasks/999 (Get task status)
	req = httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", rr.Code)
	}

	var fetchedTask task.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &fetchedTask); err != nil {
		t.Fatalf("failed to decode fetched task: %v", err)
	}
	if fetchedTask.ID != 999 || fetchedTask.Name != "CustomTaskName" {
		t.Errorf("fetched task mismatch: %+v", fetchedTask)
	}

	// 4. GET /tasks/888 (Missing task ID, should return 404)
	req = httptest.NewRequest(http.MethodGet, "/tasks/888", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found for missing task, got %d", rr.Code)
	}

	// 5. PUT /tasks/999 (Update status to Processing)
	req = httptest.NewRequest(http.MethodPut, "/tasks/999", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", rr.Code)
	}

	var putTask task.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &putTask); err != nil {
		t.Fatalf("failed to decode PUT response: %v", err)
	}
	if putTask.Status != types.Processing {
		t.Errorf("expected task status to be Processing, got %s", putTask.Status)
	}

	// 6. Test invalid route patterns
	req = httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request for invalid ID syntax, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/tasks/123/extra", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request for extra subpaths, got %d", rr.Code)
	}
}

func TestHTTPHandlers_PostMockFallback(t *testing.T) {
	wg := &sync.WaitGroup{}
	taskMng := task.NewLocalTaskManager()
	channel := queue.NewChannel(10, wg)

	// Route handler
	handler := PostTasksHandler(taskMng, channel, wg)

	// POST /task with no body (should fallback to generating 4 mock tasks)
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}

	var createdTasks []*task.Task
	if err := json.Unmarshal(rr.Body.Bytes(), &createdTasks); err != nil {
		t.Fatalf("failed to decode fallback mock tasks: %v", err)
	}

	if len(createdTasks) != 4 {
		t.Errorf("expected 4 mock tasks to be created, got %d", len(createdTasks))
	}

	// Clean up pending queue wg
	for range createdTasks {
		channel.TaskDone()
	}
}
