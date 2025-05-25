package web

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"bw_util/internal/models"
	"bw_util/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	t.Run("should create new handler", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Parse("test"))
		templateManager := &TemplateManager{tmpl: tmpl}
		dbPath := "/test/path.db"

		handler := NewHandler(nil, templateManager, dbPath)

		assert.NotNil(t, handler)
		assert.Equal(t, templateManager, handler.templateManager)
		assert.Equal(t, dbPath, handler.dbPath)
	})
}

func TestHandler_ServeLogsPage_MethodValidation(t *testing.T) {
	t.Run("should handle POST request with method not allowed", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Parse("test"))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(nil, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})

	t.Run("should handle PUT request with method not allowed", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Parse("test"))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(nil, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodPut, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})

	t.Run("should handle DELETE request with method not allowed", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Parse("test"))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(nil, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})
}

func TestHandler_ServeLogsPage_Success(t *testing.T) {
	t.Run("should serve logs page successfully", func(t *testing.T) {
		// Create a mock repository
		mockRepo := &testutil.MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				return testutil.SampleLogEntries, false, nil
			},
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
		}

		// Create a simple template
		tmpl := template.Must(template.New("logs.html").Parse(`
			<h1>Logs</h1>
			<p>DB: {{.DBPath}}</p>
			<p>Results: {{.ResultCount}}</p>
			{{range .Logs}}
				<div>{{.ContainerName}}: {{.LogMessage}}</div>
			{{end}}
		`))
		templateManager := &TemplateManager{tmpl: tmpl}

		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Logs")
		assert.Contains(t, w.Body.String(), "/test/db.db")
		assert.Contains(t, w.Body.String(), "web-server")
	})

	t.Run("should handle search query", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				// Verify search parameters
				assert.Equal(t, "test search", params.SearchQuery)
				return testutil.SampleLogEntries[:2], false, nil
			},
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`Results: {{.ResultCount}}`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/?search=test+search", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Results: 2")
	})

	t.Run("should handle container filtering", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				// Verify container filtering
				assert.Equal(t, []string{"web-server", "database"}, params.SelectedContainers)
				return testutil.SampleLogEntries[:3], false, nil
			},
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`Results: {{.ResultCount}}`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/?containers=web-server,database", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Results: 3")
	})

	t.Run("should handle pagination", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				// Verify pagination parameters
				assert.Equal(t, 2, params.CurrentPage)
				assert.Equal(t, 100, params.Offset)         // (2-1) * 100
				return testutil.SampleLogEntries, true, nil // hasNext = true
			},
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`
			Page: {{.CurrentPage}}
			HasNext: {{.HasNext}}
			HasPrev: {{.HasPrev}}
			NextURL: {{.NextPageURL}}
			PrevURL: {{.PrevPageURL}}
		`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "Page: 2")
		assert.Contains(t, body, "HasNext: true")
		assert.Contains(t, body, "HasPrev: true")
		assert.Contains(t, body, "NextURL: /?page=3")
		assert.Contains(t, body, "PrevURL: /?page=1")
	})

	t.Run("should handle page limit", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				// Verify page limit is enforced
				assert.Equal(t, 1000, params.CurrentPage) // Should be capped at 1000
				return testutil.SampleLogEntries, false, nil
			},
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`Page: {{.CurrentPage}}`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/?page=9999", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Page: 1000")
	})
}

func TestHandler_ServeLogsPage_Errors(t *testing.T) {
	t.Run("should handle repository GetContainers error", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return nil, assert.AnError
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`test`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to load container list")
	})

	t.Run("should handle repository Search error", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				return nil, false, assert.AnError
			},
		}

		tmpl := template.Must(template.New("logs.html").Parse(`test`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve logs")
	})

	t.Run("should handle template execution error", func(t *testing.T) {
		mockRepo := &testutil.MockRepository{
			GetContainersFunc: func(ctx context.Context) ([]models.Container, error) {
				return testutil.SampleContainers, nil
			},
			SearchFunc: func(ctx context.Context, params *models.SearchParams) ([]models.LogEntry, bool, error) {
				return testutil.SampleLogEntries, false, nil
			},
		}

		// Create a template that will fail execution
		tmpl := template.Must(template.New("logs.html").Parse(`{{.NonExistentField.BadAccess}}`))
		templateManager := &TemplateManager{tmpl: tmpl}
		handler := NewHandler(mockRepo, templateManager, "/test/db.db")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeLogsPage(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to render page")
	})
}

func TestHandler_buildPaginationURL(t *testing.T) {
	handler := &Handler{}

	t.Run("should build URL with page only", func(t *testing.T) {
		url := handler.buildPaginationURL(2, "", []string{})
		assert.Equal(t, "/?page=2", url)
	})

	t.Run("should build URL with page and search", func(t *testing.T) {
		url := handler.buildPaginationURL(3, "test search", []string{})
		assert.Equal(t, "/?page=3&search=test+search", url)
	})

	t.Run("should build URL with page and containers", func(t *testing.T) {
		url := handler.buildPaginationURL(1, "", []string{"container1", "container2"})
		assert.Equal(t, "/?page=1&containers=container1%2Ccontainer2", url)
	})

	t.Run("should build URL with all parameters", func(t *testing.T) {
		url := handler.buildPaginationURL(5, "test query", []string{"app", "db"})
		expected := "/?page=5&search=test+query&containers=app%2Cdb"
		assert.Equal(t, expected, url)
	})

	t.Run("should handle special characters in search", func(t *testing.T) {
		url := handler.buildPaginationURL(1, "test & query", []string{})
		assert.Contains(t, url, "test+%26+query")
	})

	t.Run("should handle special characters in containers", func(t *testing.T) {
		url := handler.buildPaginationURL(1, "", []string{"container-1", "container_2"})
		assert.Contains(t, url, "container-1")
		assert.Contains(t, url, "container_2")
	})

	t.Run("should handle empty parameters", func(t *testing.T) {
		url := handler.buildPaginationURL(1, "", []string{})
		assert.Equal(t, "/?page=1", url)
	})

	t.Run("should handle high page numbers", func(t *testing.T) {
		url := handler.buildPaginationURL(999, "", []string{})
		assert.Equal(t, "/?page=999", url)
	})
}

func TestHandler_StructValidation(t *testing.T) {
	t.Run("should validate Handler struct", func(t *testing.T) {
		handler := &Handler{}

		assert.NotNil(t, handler)
		assert.Nil(t, handler.repository)
		assert.Nil(t, handler.templateManager)
		assert.Empty(t, handler.dbPath)
	})

	t.Run("should handle nil dependencies", func(t *testing.T) {
		handler := NewHandler(nil, nil, "")

		assert.NotNil(t, handler)
		assert.Nil(t, handler.repository)
		assert.Nil(t, handler.templateManager)
		assert.Empty(t, handler.dbPath)
	})
}

func TestHandler_ParameterParsing(t *testing.T) {
	t.Run("should parse search parameters correctly", func(t *testing.T) {
		// Test URL parameter parsing logic
		testURL := "/?search=test+query&containers=app,db&page=2"
		parsedURL, err := url.Parse(testURL)
		assert.NoError(t, err)

		searchQuery := parsedURL.Query().Get("search")
		assert.Equal(t, "test query", searchQuery)

		containersParam := parsedURL.Query().Get("containers")
		assert.Equal(t, "app,db", containersParam)

		pageParam := parsedURL.Query().Get("page")
		assert.Equal(t, "2", pageParam)
	})

	t.Run("should handle URL encoding", func(t *testing.T) {
		testURL := "/?search=" + url.QueryEscape("test & special chars")
		parsedURL, err := url.Parse(testURL)
		assert.NoError(t, err)

		searchQuery := parsedURL.Query().Get("search")
		assert.Equal(t, "test & special chars", searchQuery)
	})
}
