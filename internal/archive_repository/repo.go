package archives

import (
	"errors"
	"sync"
	errvalues "testcase/internal/errors"
	"testcase/models"
)

var (
	maxTasks int
	maxFiles int
)

type Manager struct {
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func New(maxTasksCount, maxFilesCount int) *Manager {
	maxTasks = maxTasksCount
	maxFiles = maxFilesCount
	return &Manager{
		mu:    sync.RWMutex{},
		tasks: make(map[string]*models.Task),
	}
}

// Registers new task if manager has less than maxTasks and returns task-id,
// otherwise returns ErrManyTasks
func (m *Manager) CreateTask() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tasks) == maxTasks {
		return "", errvalues.ErrManyTasks
	}
	task := &models.Task{
		TaskID: generateID(),
		Files:  make([]*models.FileRequest, 0, maxFiles),
		Status: models.Staged,
	}
	m.tasks[task.TaskID] = task
	return task.TaskID, nil
}

// Searchs task with provided taskID. If there there is one, adds file to task
// or returns ErrTaskFull if tasks limit exceeded
func (m *Manager) AddFile(taskID string, file *models.FileRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, exist := m.tasks[taskID]
	if !exist {
		return errvalues.ErrNoSuchTask
	}
	if task.Status == models.Completed {
		return errvalues.ErrTaskFull
	}
	task.Files = append(task.Files, file)
	if len(task.Files) == maxFiles {
		task.Status = models.Completed
	}
	m.tasks[taskID] = task
	return nil
}

// Returns task status if there is task with provided ID, otherwise returns ErrNoSuchTask
func (m *Manager) GetTaskStatus(taskID string) (models.TaskStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, exist := m.tasks[taskID]
	if !exist {
		return "", errvalues.ErrNoSuchTask
	}
	return task.Status, nil
}

// Returns archive data if there is task with given ID, then deleted task
func (m *Manager) GetArchive(taskID string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, exist := m.tasks[taskID]
	if !exist {
		return nil, errvalues.ErrNoSuchTask
	}
	result, err := getRawArchive(task.Files)
	if err != nil {
		return nil, errors.New("getting result archive error(s): " + err.Error())
	}
	delete(m.tasks, taskID)
	return result, nil
}
