package docker

import (
	"context"
	"testing"

	"bw_util/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogRepository is a mock implementation of the LogRepository interface
type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) Store(ctx context.Context, entry *models.LogEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockLogRepository) Search(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]models.LogEntry), args.Bool(1), args.Error(2)
}

func (m *MockLogRepository) GetContainers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func TestNewCollector(t *testing.T) {
	t.Run("should create new collector", func(t *testing.T) {
		mockClient := &Client{}
		mockRepo := &MockLogRepository{}

		collector := NewCollector(mockClient, mockRepo)

		assert.NotNil(t, collector)
		assert.Equal(t, mockClient, collector.client)
		assert.Equal(t, mockRepo, collector.repository)
	})
}

func TestCollector_StartCollection(t *testing.T) {
	t.Run("should handle empty container list", func(t *testing.T) {
		mockClient := &Client{}
		mockRepo := &MockLogRepository{}
		collector := NewCollector(mockClient, mockRepo)

		containers := []models.Container{}

		ctx := context.Background()
		err := collector.StartCollection(ctx, containers)

		assert.NoError(t, err)
	})
}

func TestCollector_Wait(t *testing.T) {
	t.Run("should wait for goroutines to finish", func(t *testing.T) {
		mockClient := &Client{}
		mockRepo := &MockLogRepository{}
		collector := NewCollector(mockClient, mockRepo)

		// This should not block since no goroutines are running
		collector.Wait()

		// Test passes if Wait() returns without hanging
		assert.True(t, true)
	})
}
