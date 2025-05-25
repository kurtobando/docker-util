package app

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/templates/*.html
var testTemplatesFS embed.FS

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
}

func TestApp_StructValidation(t *testing.T) {
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
}
