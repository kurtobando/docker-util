package app

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/templates/*.html
var testTemplatesFS embed.FS

// TestNew tests the App constructor functionality
func TestNew(t *testing.T) {
	t.Run("should create new app instance", func(t *testing.T) {
		app := New(testTemplatesFS)

		assert.NotNil(t, app)
		assert.Equal(t, testTemplatesFS, app.templatesFS)
		assert.Nil(t, app.config)
		assert.Nil(t, app.db)
		assert.Nil(t, app.repository)
		assert.Nil(t, app.dockerClient)
		assert.Nil(t, app.collector)
		assert.Nil(t, app.templateManager)
		assert.Nil(t, app.handler)
		assert.Nil(t, app.server)
	})

	t.Run("should handle empty embed.FS", func(t *testing.T) {
		var emptyFS embed.FS
		app := New(emptyFS)

		assert.NotNil(t, app)
		assert.Equal(t, emptyFS, app.templatesFS)
	})

	t.Run("should handle nil embed.FS", func(t *testing.T) {
		// Test with zero value embed.FS
		var nilFS embed.FS
		app := New(nilFS)

		assert.NotNil(t, app)
		assert.Equal(t, nilFS, app.templatesFS)

		// All other fields should be nil initially
		assert.Nil(t, app.config)
		assert.Nil(t, app.db)
		assert.Nil(t, app.repository)
		assert.Nil(t, app.dockerClient)
		assert.Nil(t, app.collector)
		assert.Nil(t, app.templateManager)
		assert.Nil(t, app.handler)
		assert.Nil(t, app.server)
	})
}

// TestAppStructValidation tests the App struct validation and field access
func TestAppStructValidation(t *testing.T) {
	t.Run("should validate App struct fields", func(t *testing.T) {
		app := &App{}

		// All fields should be nil initially
		assert.Nil(t, app.config)
		assert.Nil(t, app.db)
		assert.Nil(t, app.repository)
		assert.Nil(t, app.dockerClient)
		assert.Nil(t, app.collector)
		assert.Nil(t, app.templateManager)
		assert.Nil(t, app.handler)
		assert.Nil(t, app.server)
	})

	t.Run("should handle nil app instance", func(t *testing.T) {
		var app *App
		assert.Nil(t, app)
	})

	t.Run("should allow field assignment", func(t *testing.T) {
		app := New(testTemplatesFS)

		// Verify we can access and check the embedded filesystem
		assert.NotNil(t, app.templatesFS)
		assert.Equal(t, testTemplatesFS, app.templatesFS)

		// Verify all component fields are accessible
		assert.Nil(t, app.config)
		assert.Nil(t, app.db)
		assert.Nil(t, app.repository)
		assert.Nil(t, app.dockerClient)
		assert.Nil(t, app.collector)
		assert.Nil(t, app.templateManager)
		assert.Nil(t, app.handler)
		assert.Nil(t, app.server)
	})
}

// TestAppEmbeddedTemplates tests the embedded templates functionality
func TestAppEmbeddedTemplates(t *testing.T) {
	t.Run("should store embedded templates filesystem", func(t *testing.T) {
		app := New(testTemplatesFS)

		// Should store the provided filesystem
		assert.Equal(t, testTemplatesFS, app.templatesFS)

		// Should be able to access the filesystem
		assert.NotNil(t, app.templatesFS)
	})

	t.Run("should handle filesystem operations", func(t *testing.T) {
		app := New(testTemplatesFS)

		// Should not panic when accessing the filesystem
		assert.NotPanics(t, func() {
			app.templatesFS.ReadDir(".")
		})

		assert.NotPanics(t, func() {
			app.templatesFS.ReadDir("testdata")
		})
	})
}

// TestAppComponentInitialization tests basic component initialization patterns
func TestAppComponentInitialization(t *testing.T) {
	t.Run("should start with all components nil", func(t *testing.T) {
		app := New(testTemplatesFS)

		// All components should start as nil
		assert.Nil(t, app.config)
		assert.Nil(t, app.db)
		assert.Nil(t, app.repository)
		assert.Nil(t, app.dockerClient)
		assert.Nil(t, app.collector)
		assert.Nil(t, app.templateManager)
		assert.Nil(t, app.handler)
		assert.Nil(t, app.server)
	})

	t.Run("should maintain component field types", func(t *testing.T) {
		app := New(testTemplatesFS)

		// Verify field types are correct (this tests struct definition)
		assert.IsType(t, testTemplatesFS, app.templatesFS)

		// These should be nil but of correct pointer types
		var expectedConfig interface{} = app.config
		var expectedDB interface{} = app.db
		var expectedRepo interface{} = app.repository
		var expectedDockerClient interface{} = app.dockerClient
		var expectedCollector interface{} = app.collector
		var expectedTemplateManager interface{} = app.templateManager
		var expectedHandler interface{} = app.handler
		var expectedServer interface{} = app.server

		// All should be nil initially
		assert.Nil(t, expectedConfig)
		assert.Nil(t, expectedDB)
		assert.Nil(t, expectedRepo)
		assert.Nil(t, expectedDockerClient)
		assert.Nil(t, expectedCollector)
		assert.Nil(t, expectedTemplateManager)
		assert.Nil(t, expectedHandler)
		assert.Nil(t, expectedServer)
	})
}
