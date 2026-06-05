package task

import (
	"errors"
	"fmt"
	"github.com/devenairevo/task-management/internal/types"
	"github.com/devenairevo/task-management/internal/user"
	"sync"
)

type Task struct {
	ID     int          `json:"id"`
	Name   string       `json:"name"`
	Status types.Status `json:"status"`
}

func (t *Task) Clone() *Task {
	if t == nil {
		return nil
	}
	return &Task{
		ID:     t.ID,
		Name:   t.Name,
		Status: t.Status,
	}
}

type CreateTaskParams struct {
	UserID   int
	UserName string
	TaskID   int
	Name     string
}

func NewTask(id int, name string, status types.Status) (*Task, error) {
	if id != 0 && name != "" {
		return &Task{ID: id, Name: name, Status: status}, nil
	}

	return nil, errors.New("something wrong with creating a new task")
}

func (lt *LocalTaskManager) CreateTask(params *CreateTaskParams) (*Task, error) {
	if params.UserID <= 0 || params.TaskID == 0 || params.Name == "" {
		return nil, errors.New("couldn't create user with task")
	}

	task, err := NewTask(params.TaskID, params.Name, types.Created)
	if err != nil {
		return nil, err
	}

	u, err := user.New(params.UserID, params.UserName)
	if err != nil {
		return nil, err
	}

	userTask, err := NewUserTask(u, task)
	if err != nil {
		return nil, err
	}

	lt.AddTask(userTask)

	return task.Clone(), nil
}

func NewUserTask(user *user.User, task *Task) (*UserTask, error) {
	if user.ID > 0 && task.ID != 0 && task.Name != "" {
		return &UserTask{
			User: user,
			Task: task,
		}, nil
	}
	return nil, errors.New("couldn't create user with task")
}

type LocalTaskManager struct {
	mu        sync.RWMutex
	userTasks map[int]*UserTask
}

func NewLocalTaskManager() *LocalTaskManager {
	return &LocalTaskManager{
		userTasks: make(map[int]*UserTask),
	}
}

func (lt *LocalTaskManager) AddTask(userTasks ...*UserTask) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	for _, v := range userTasks {
		if v != nil && v.Task != nil {
			lt.userTasks[v.Task.ID] = v
		}
	}
}

func (lt *LocalTaskManager) GetUserTasks(userID int) []*Task {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	var userTasks []*Task

	for _, userTask := range lt.userTasks {
		if userTask.User.ID == userID {
			userTasks = append(userTasks, userTask.Task.Clone())
		}
	}

	return userTasks
}

func (lt *LocalTaskManager) DescribeTask(taskID int) (*Task, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	userTask, exists := lt.userTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("❌  Task with ID %d not found", taskID)
	}

	return userTask.Task.Clone(), nil
}

func (lt *LocalTaskManager) ListTasks() ([]*Task, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	if len(lt.userTasks) == 0 {
		return nil, errors.New("no tasks found")
	}

	var tasks []*Task
	for _, userTask := range lt.userTasks {
		if userTask.Task != nil {
			tasks = append(tasks, userTask.Task.Clone())
		}
	}

	return tasks, nil
}

func (lt *LocalTaskManager) UpdateTask(task *Task) error {
	if task == nil || task.ID < 0 {
		return errors.New("no task found")
	}

	lt.mu.Lock()
	defer lt.mu.Unlock()
	userTask, exists := lt.userTasks[task.ID]
	if !exists {
		return fmt.Errorf("task with ID %d not found", task.ID)
	}

	userTask.Task.Status = task.Status
	userTask.Task.Name = task.Name

	return nil
}
