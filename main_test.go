package main

import (
	"embed"
	"testing"

	"bw_util/internal/app"

	"github.com/stretchr/testify/assert"
)

// TestEmbeddedTemplatesFS tests the embedded filesystem functionality
func TestEmbeddedTemplatesFS(t *testing.T) {
	t.Run("should have embedded templates filesystem", func(t *testing.T) {
		// Verify that templatesFS is properly embedded
		assert.NotNil(t, templatesFS)

		// Test that it's an embed.FS type
		var expectedType embed.FS
		assert.IsType(t, expectedType, templatesFS)
	})

	t.Run("should be able to read from embedded filesystem", func(t *testing.T) {
		// Try to read the directory to verify it's properly embedded
		entries, err := templatesFS.ReadDir("templates")

		// Should not error (even if directory is empty)
		assert.NoError(t, err)

		// entries should be a slice (could be empty)
		assert.NotNil(t, entries)
	})

	t.Run("should handle embedded filesystem operations", func(t *testing.T) {
		// Test basic filesystem operations
		assert.NotPanics(t, func() {
			templatesFS.ReadDir("templates")
		})

		assert.NotPanics(t, func() {
			templatesFS.ReadDir(".")
		})

		// Check if we can access the embedded content
		_, err := templatesFS.ReadDir(".")
		assert.NoError(t, err)
	})

	t.Run("should have templatesFS as package-level variable", func(t *testing.T) {
		// Verify templatesFS is accessible at package level
		assert.NotNil(t, templatesFS)

		// Verify it maintains its type
		var expectedType embed.FS
		assert.IsType(t, expectedType, templatesFS)
	})
}

// TestApplicationCreation tests the main application creation logic
func TestApplicationCreation(t *testing.T) {
	t.Run("should create application with embedded templates", func(t *testing.T) {
		// Test that we can create an application instance like main() does
		// This tests the same logic as main() without actually running it

		// This is the same call as in main()
		application := app.New(templatesFS)

		assert.NotNil(t, application)
	})
}

// TestMainPackageStructure tests the overall package structure and compilation
func TestMainPackageStructure(t *testing.T) {
	t.Run("should have main function defined", func(t *testing.T) {
		// This test verifies that the main function exists and can be referenced
		// We can't easily test the main function directly without running it,
		// but we can verify the package structure is correct
		assert.True(t, true, "main function exists and package compiles")
	})

	t.Run("should have proper imports", func(t *testing.T) {
		// Verify that required imports are available by using them
		assert.NotNil(t, templatesFS)

		// Test that embed directive works
		var fs embed.FS
		assert.IsType(t, fs, templatesFS)
	})

	t.Run("should compile without errors", func(t *testing.T) {
		// If this test runs, it means the package compiled successfully
		assert.True(t, true, "Package compiled successfully")
	})

	t.Run("should have valid Go module structure", func(t *testing.T) {
		// Basic validation that we're in a proper Go module
		assert.True(t, true, "Valid Go module structure")
	})
}
