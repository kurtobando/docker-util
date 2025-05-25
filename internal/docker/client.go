package docker

import (
	"context"
	"log"
	"strings"

	"bw_util/internal/models"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Client wraps the Docker client
type Client struct {
	client *client.Client
}

// NewClient creates a new Docker client
// This preserves the exact same client creation logic from the original main function
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{client: cli}, nil
}

// ListRunningContainers returns all currently running containers
// This preserves the exact same container listing logic from the original main function
func (c *Client) ListRunningContainers(ctx context.Context) ([]models.Container, error) {
	containers, err := c.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, err
	}

	var result []models.Container
	for _, cont := range containers {
		containerName := "unknown"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
		}

		result = append(result, models.Container{
			ID:    cont.ID,
			Name:  containerName,
			Image: cont.Image,
		})
	}

	if len(result) == 0 {
		log.Println("No running containers found to monitor.")
	} else {
		log.Printf("Found %d running containers. Starting log streaming...\n", len(result))
		for _, cont := range result {
			log.Printf("Starting log stream for: ID=%s, Name=%s, Image=%s\n", cont.ID[:12], cont.Name, cont.Image)
		}
	}

	return result, nil
}

// GetClient returns the underlying Docker client
func (c *Client) GetClient() *client.Client {
	return c.client
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
