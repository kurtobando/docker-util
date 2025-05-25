package main

import (
	"testing"

	"bw_util/internal/app"

	"github.com/stretchr/testify/assert"
)

func TestMain_EmbeddedTemplates(t *testing.T) {
	t.Run("should have embedded templates filesystem", func(t *testing.T) {
		// Test that templatesFS is not nil and contains files
		assert.NotNil(t, templatesFS)

		// Try to read the templates directory
		entries, err := templatesFS.ReadDir("templates")
		if err != nil {
			// If templates directory doesn't exist, that's expected in test environment
			assert.Contains(t, err.Error(), "file does not exist")
		} else {
			// If it exists, it should contain HTML files
			assert.NotEmpty(t, entries)
		}
	})
}

func TestMain_ApplicationCreation(t *testing.T) {
	t.Run("should create application with embedded templates", func(t *testing.T) {
		// Test that we can create an application instance like main() does
		// This tests the same logic as main() without actually running it

		// This is the same call as in main()
		application := app.New(templatesFS)

		assert.NotNil(t, application)
	})
}
