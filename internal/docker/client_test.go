package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("should create new Docker client", func(t *testing.T) {
		// Note: This test will try to create a real Docker client
		// In a real environment with Docker daemon running, this should work
		// In CI/CD or environments without Docker, this might fail
		client, err := NewClient()

		if err != nil {
			// If Docker is not available, we should get a specific error
			assert.Contains(t, err.Error(), "docker")
			assert.Nil(t, client)
		} else {
			// If Docker is available, client should be created successfully
			assert.NotNil(t, client)
			assert.NotNil(t, client.client)

			// Clean up
			client.Close()
		}
	})
}

func TestClient_GetClient(t *testing.T) {
	t.Run("should return underlying Docker client", func(t *testing.T) {
		client, err := NewClient()
		if err != nil {
			t.Skip("Docker not available, skipping test")
		}
		defer client.Close()

		dockerClient := client.GetClient()
		assert.NotNil(t, dockerClient)
		assert.Equal(t, client.client, dockerClient)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("should close Docker client successfully", func(t *testing.T) {
		client, err := NewClient()
		if err != nil {
			t.Skip("Docker not available, skipping test")
		}

		err = client.Close()
		assert.NoError(t, err)
	})

	t.Run("should handle closing nil client", func(t *testing.T) {
		client := &Client{client: nil}

		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestClient_ListRunningContainers(t *testing.T) {
	t.Run("should list running containers successfully", func(t *testing.T) {
		client, err := NewClient()
		if err != nil {
			t.Skip("Docker not available, skipping test")
		}
		defer client.Close()

		ctx := context.Background()
		containers, err := client.ListRunningContainers(ctx)

		// Should not error even if no containers are running
		assert.NoError(t, err)
		assert.NotNil(t, containers)
		// containers slice can be empty if no containers are running
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		client, err := NewClient()
		if err != nil {
			t.Skip("Docker not available, skipping test")
		}
		defer client.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		containers, err := client.ListRunningContainers(ctx)

		// Should handle cancellation gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
		// containers might be nil or empty depending on timing
		_ = containers
	})
}

func TestClient_StructValidation(t *testing.T) {
	t.Run("should validate Client struct", func(t *testing.T) {
		// Test the Client struct itself
		client := &Client{}

		assert.NotNil(t, client)
		assert.Nil(t, client.client) // Should be nil until initialized

		// Test that GetClient returns nil for uninitialized client
		dockerClient := client.GetClient()
		assert.Nil(t, dockerClient)

		// Test that Close doesn't panic with nil client
		err := client.Close()
		assert.NoError(t, err)
	})
}
