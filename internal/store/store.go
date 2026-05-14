package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

type Store struct {
	path string
	mu   sync.Mutex
	lock *flock.Flock
}

func Open() (*Store, error) {
	p := ConfigFilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return nil, err
	}
	return &Store{
		path: p,
		lock: flock.New(p + ".lock"),
	}, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) Load() (Schema, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.acquireRead(); err != nil {
		return DefaultSchema(), err
	}
	defer s.lock.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return DefaultSchema(), nil
	}
	if err != nil {
		return DefaultSchema(), err
	}
	if len(data) == 0 {
		return DefaultSchema(), nil
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return DefaultSchema(), err
	}
	if schema.Aliases == nil {
		schema.Aliases = []Alias{}
	}
	if schema.Settings.DefaultLocalPortStart == 0 {
		schema.Settings.DefaultLocalPortStart = 9000
	}
	return schema, nil
}

func (s *Store) Save(schema Schema) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.acquireWrite(); err != nil {
		return err
	}
	defer s.lock.Unlock()

	data, err := json.MarshalIndent(schema, "", "\t")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Update(fn func(*Schema)) (Schema, error) {
	schema, err := s.Load()
	if err != nil {
		return schema, err
	}
	fn(&schema)
	return schema, s.Save(schema)
}

func (s *Store) acquireRead() error {
	return s.acquire(false)
}

func (s *Store) acquireWrite() error {
	return s.acquire(true)
}

func (s *Store) acquire(write bool) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		var ok bool
		var err error
		if write {
			ok, err = s.lock.TryLock()
		} else {
			ok, err = s.lock.TryRLock()
		}
		if ok {
			return nil
		}
		if err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return errors.New("config file is locked by another process")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
