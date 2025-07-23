package planner

import (
	"sync"
	"time"
)

type Planner struct {
	tasks      map[string]*Task
	tasksQueue chan struct{}
	mu         sync.Mutex
}

type Task struct {
	Id          string
	Status      string // "new", "in_progress", "done", "failed"
	Files       []FileInfo
	ZipLink     string
	CreatedTime time.Time
}

type FileInfo struct {
	URL      string
	Status   string // "downloaded", "failed"
	ErrorMsg string
}

func NewPlanner() *Planner {
	return &Planner{
		tasks:      make(map[string]*Task),
		tasksQueue: make(chan struct{}, 3),
	}
}

//TODO: make new task
//TODO: add obj to new task
//TODO: start task
//TODO:
