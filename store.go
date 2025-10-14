package main

import (
	"encoding/json"
	"os"       
	"errors"
	"sync"
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type diskImage struct {
	NextID int64           `json:"next_id"`
	Data   map[int64]Task  `json:"data"`
}

var ErrNotFound = errors.New("not found")

type TaskStore struct {
	mu     sync.RWMutex
	data   map[int64]Task
	nextID int64
	filePath string
}

func NewTaskStore(filePath string) *TaskStore {
	s := &TaskStore{
		data:     make(map[int64]Task),
		nextID:   1,
		filePath: filePath,
	}
	_ = s.Load()
	return s
}

func (s *TaskStore) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	t := Task{ID: id, Title: title, Done: false}
	s.data[id] = t
	return t
}

func (s *TaskStore) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.data))
	for _, t := range s.data {
		out = append(out, t)
	}
	return out
}

func (s *TaskStore) Get(id int64) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.data[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (s *TaskStore) Update(t Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[t.ID]; !ok {
		return ErrNotFound
	}
	s.data[t.ID] = t
	return nil
}

func (s *TaskStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return nil
}

func (s *TaskStore) Save() error {
	if s.filePath == "" {
		return nil
	}
	s.mu.RLock()
	img := diskImage{
		NextID: s.nextID,
		Data:   s.data,
	}
	s.mu.RUnlock()

	b, err := json.MarshalIndent(img, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, b, 0644)
}

func (s *TaskStore) Load() error {
	if s.filePath == "" {
		return nil
	}
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var img diskImage
	if err := json.Unmarshal(b, &img); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = img.Data
	if s.data == nil {
		s.data = make(map[int64]Task)
	}
	if img.NextID > 0 {
		s.nextID = img.NextID
	} else {
		var max int64 = 0
		for id := range s.data {
			if id > max {
				max = id
			}
		}
		s.nextID = max + 1
	}
	return nil
}