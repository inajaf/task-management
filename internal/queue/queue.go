package queue

import (
	"errors"
	"fmt"
	"github.com/devenairevo/task-management/internal/task"
	"sync"
)

type Channel struct {
	mu       sync.Mutex
	wg       *sync.WaitGroup
	BuffSize int
	Task     *task.Task
	chanel   chan *task.Task
}

func NewChannel(buffSize int, wg *sync.WaitGroup) *Channel {
	return &Channel{
		BuffSize: buffSize,
		chanel:   make(chan *task.Task, buffSize),
		wg:       wg,
	}
}

func (c *Channel) Enqueue(task *task.Task) error {
	if task.ID <= 0 || task.Name == "" {
		return errors.New("error with adding task to the queue")
	}
	if c.chanel == nil {
		return errors.New("channel not created yet")
	}

	c.wg.Add(1)
	c.chanel <- task

	fmt.Printf("Task with ID %d and name '%s' added to the queue\n", task.ID, task.Name)
	return nil
}

func (c *Channel) Dequeue() (*task.Task, error) {
	if c.chanel == nil {
		return nil, errors.New("channel not created yet")
	}

	t, ok := <-c.chanel
	if !ok {
		return nil, errors.New("channel closed")
	}

	return t, nil
}

func (c *Channel) TaskDone() {
	if c.wg != nil {
		c.wg.Done()
	}
}

func (c *Channel) IsEmpty() bool {
	return len(c.chanel) == 0
}

func (c *Channel) Size() int {
	return len(c.chanel)
}
