package testutil

import (
	"context"
	"errors"
	"io"

	"bw_util/internal/models"
)

// MockDockerClient implements a mock for our Docker client wrapper
type MockDockerClient struct {
	ListRunningContainersFunc func(ctx context.Context) ([]models.Container, error)
	GetClientFunc             func() interface{} // Returns the underlying client
	CloseFunc                 func() error
}

func (m *MockDockerClient) ListRunningContainers(ctx context.Context) ([]models.Container, error) {
	if m.ListRunningContainersFunc != nil {
		return m.ListRunningContainersFunc(ctx)
	}
	return []models.Container{}, nil
}

func (m *MockDockerClient) GetClient() interface{} {
	if m.GetClientFunc != nil {
		return m.GetClientFunc()
	}
	return nil
}

func (m *MockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockDockerClientError returns a mock that always errors
func NewMockDockerClientError() *MockDockerClient {
	return &MockDockerClient{
		ListRunningContainersFunc: func(ctx context.Context) ([]models.Container, error) {
			return nil, errors.New("docker daemon not available")
		},
		CloseFunc: func() error {
			return nil
		},
	}
}

// MockRepository implements a mock repository for testing
type MockRepository struct {
	StoreFunc         func(ctx context.Context, entry *models.LogEntry) error
	SearchFunc        func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error)
	GetContainersFunc func(ctx context.Context) ([]models.Container, error)
}

func (m *MockRepository) Store(ctx context.Context, entry *models.LogEntry) error {
	if m.StoreFunc != nil {
		return m.StoreFunc(ctx, entry)
	}
	return nil
}

func (m *MockRepository) Search(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, params)
	}
	return []models.LogEntry{}, false, nil
}

func (m *MockRepository) GetContainers(ctx context.Context) ([]string, error) {
	if m.GetContainersFunc != nil {
		containers, err := m.GetContainersFunc(ctx)
		if err != nil {
			return nil, err
		}
		// Convert []models.Container to []string
		var names []string
		for _, container := range containers {
			names = append(names, container.Name)
		}
		return names, nil
	}
	return []string{}, nil
}

// MockLogReader implements io.ReadCloser for simulating Docker log streams
type MockLogReader struct {
	data   []byte
	pos    int
	closed bool
}

func NewMockLogReader(data string) *MockLogReader {
	return &MockLogReader{
		data: []byte(data),
		pos:  0,
	}
}

func (m *MockLogReader) Read(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.EOF
	}

	if m.pos >= len(m.data) {
		return 0, io.EOF
	}

	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *MockLogReader) Close() error {
	m.closed = true
	return nil
}
