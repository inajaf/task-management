package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devenairevo/task-management/internal/task"
	"github.com/devenairevo/task-management/internal/types"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	dirPath string
}

type TaskFile task.Task

func NewManager(dirName string) *Manager {
	return &Manager{dirPath: dirName}
}

func (m *Manager) NewFile(task *task.Task) error {
	if err := os.MkdirAll(m.dirPath, 0755); err != nil {
		return errors.New("issue with the creating directory")
	}

	jsonData, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return errors.New("issue with marshaling")
	}

	filename := fmt.Sprintf("task_%s_%d.json", task.Name, task.ID)
	fullPath := filepath.Join(m.dirPath, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return errors.New("error with creating file")
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return errors.New("issue with a writing data")
	}

	return nil
}

func (m *Manager) CreateTask(params *task.CreateTaskParams) (*task.Task, error) {
	if params.UserID <= 0 || params.TaskID == 0 || params.Name == "" {
		return nil, errors.New("couldn't create user with task")
	}

	t, err := task.NewTask(params.TaskID, params.Name, types.Created)
	if err != nil {
		return nil, err
	}

	if err := m.NewFile(t); err != nil {
		return nil, errors.New("couldn't create a file")
	}

	return t, nil
}

func (m *Manager) ListTasks() ([]*task.Task, error) {
	if _, err := os.Stat(m.dirPath); os.IsNotExist(err) {
		return nil, errors.New("no tasks found")
	}

	files, err := os.ReadDir(m.dirPath)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	var tasks []*task.Task

	if len(files) > 0 {
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			filePath := filepath.Join(m.dirPath, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("%s", err)
			}

			var t task.Task
			if err := json.Unmarshal(data, &t); err != nil {
				return nil, fmt.Errorf("%s", err)
			}

			tasks = append(tasks, &t)
		}
	} else {
		return nil, fmt.Errorf("files not found in folder %s", m.dirPath)
	}

	return tasks, nil
}

func (m *Manager) DescribeTask(taskID int) (*task.Task, error) {
	path := filepath.Join(m.dirPath, fmt.Sprintf("*_%d.json", taskID))
	file, err := matchFile(path, taskID)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	var t task.Task
	if err := json.Unmarshal(file, &t); err != nil {
		return nil, fmt.Errorf("issue with an unmarshaling the task %d", taskID)
	}

	return &t, nil
}

func (m *Manager) UpdateTask(t *task.Task) error {
	if t == nil || t.ID < 0 {
		return errors.New("no task found")
	}

	path := filepath.Join(m.dirPath, fmt.Sprintf("*_%d.json", t.ID))
	fileBytes, err := matchFile(path, t.ID)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	var existingTask task.Task
	if err := json.Unmarshal(fileBytes, &existingTask); err != nil {
		return fmt.Errorf("issue with an unmarshaling the task %d", t.ID)
	}

	existingTask.Status = t.Status
	existingTask.Name = t.Name

	updatedData, err := json.MarshalIndent(existingTask, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprintf("issue with an marshaling the task %d", t.ID))
	}

	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("file Task with ID %d not found", t.ID)
	}

	if err := os.WriteFile(matches[0], updatedData, 0644); err != nil {
		return fmt.Errorf("error writing updated task %d: %w", t.ID, err)
	}

	return nil
}

func matchFile(path string, taskID int) ([]byte, error) {
	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("file Task with ID %d not found", taskID)
	}

	fileBytes, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("file Task with ID %d not found", taskID)
	}

	return fileBytes, nil
}
