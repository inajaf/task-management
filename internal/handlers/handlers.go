package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/devenairevo/task-management/internal/contracts/queuer"
	"github.com/devenairevo/task-management/internal/contracts/tasker"
	"github.com/devenairevo/task-management/internal/task"
	"github.com/devenairevo/task-management/internal/types"
	"github.com/devenairevo/task-management/test/data"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type CreateTaskRequest struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	TaskID   int    `json:"task_id"`
	Name     string `json:"name"`
}

func PostTasksHandler(taskMng tasker.Tasker, channel queuer.Queuer, wg *sync.WaitGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("Adding tasks to the queue....")

		var tasksList []*task.Task

		// Check if JSON request body was provided
		if r.Body != nil && r.ContentLength > 0 {
			var req CreateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.TaskID > 0 && req.Name != "" {
				params := &task.CreateTaskParams{
					UserID:   req.UserID,
					UserName: req.UserName,
					TaskID:   req.TaskID,
					Name:     req.Name,
				}
				t, err := taskMng.CreateTask(params)
				if err == nil {
					tasksList = append(tasksList, t)
				} else {
					http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusBadRequest)
					return
				}
			}
		}

		// Fallback to generating mock tasks if no specific task was successfully created
		if len(tasksList) == 0 {
			tasksList = data.GenerateMockTasks(taskMng)
		}

		for _, t := range tasksList {
			err := channel.Enqueue(t)
			if err != nil {
				http.Error(w, "Issue with enqueuing task", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(tasksList); err != nil {
			http.Error(w, "Issue with encoding", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}
}

func GetTasksHandler(taskMng tasker.Tasker, channel queuer.Queuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("Getting all tasks....")
		w.Header().Set("Content-Type", "application/json")

		tasks, err := taskMng.ListTasks()
		if err != nil {
			// If no tasks found, return empty array rather than 500 Internal Server Error
			if strings.Contains(err.Error(), "no tasks found") || strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode([]*task.Task{})
				return
			}
			http.Error(w, "Issue with the task lists", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			http.Error(w, "Issue with encoding", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		for _, v := range tasks {
			fmt.Printf("Task with an id %d, name '%s' and status - %s\n", v.ID, v.Name, v.Status)
		}

		fmt.Printf("Queue size: %d\n", channel.Size())
		fmt.Printf("The queue got empty?: %t\n", channel.IsEmpty())
	}
}

func TaskByID(taskMng tasker.Tasker, channel queuer.Queuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clean and validate request URL path
		if !strings.HasPrefix(r.URL.Path, "/tasks/") {
			http.Error(w, "Incorrect path", http.StatusBadRequest)
			return
		}
		idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
		idStr = strings.TrimSuffix(idStr, "/")
		if strings.Contains(idStr, "/") || idStr == "" {
			http.Error(w, "Incorrect path", http.StatusBadRequest)
			return
		}

		taskID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			fmt.Println("Getting task status....")
			taskObj, err := taskMng.DescribeTask(taskID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Task with ID %d not found", taskID), http.StatusNotFound)
				fmt.Println(err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(taskObj)
			if err != nil {
				http.Error(w, "Issue with the encoding task", http.StatusInternalServerError)
				fmt.Println(err)
			}

			fmt.Printf("Found ✅ : %s with the ID %d and status - %s\n", taskObj.Name, taskObj.ID, taskObj.Status)

		case http.MethodPut:
			taskObj, err := taskMng.DescribeTask(taskID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Task with ID %d not found", taskID), http.StatusNotFound)
				fmt.Println(err)
				return
			}

			fmt.Println("Updating task....")

			taskObj.Status = types.Processing
			err = taskMng.UpdateTask(taskObj)
			if err != nil {
				http.Error(w, "Issue with the updating task", http.StatusInternalServerError)
				fmt.Println(err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(taskObj)
			if err != nil {
				http.Error(w, "Issue with the encoding task", http.StatusInternalServerError)
				fmt.Println(err)
				return
			}

			fmt.Printf("Queue size: %d\n", channel.Size())
			fmt.Printf("The queue got empty?: %t\n", channel.IsEmpty())
			fmt.Printf("Found ✅ : %s with the ID %d and status - %s\n", taskObj.Name, taskObj.ID, taskObj.Status)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
